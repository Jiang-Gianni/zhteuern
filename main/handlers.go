package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/Jiang-Gianni/zhteuern/browser"
	"github.com/Jiang-Gianni/zhteuern/taxes"
	"github.com/andybalholm/brotli"
	"github.com/skip2/go-qrcode"
)

const (
	income     = "income"
	investment = "investment"
	deduction  = "deduction"
	result     = "result"
)

var (
	pageList = []string{
		income,
		investment,
		deduction,
		result,
	}
)

var ErrPageNotFound = errors.New("page-not-found")

func (s *Server) IndexHandler() httpHandlerFunc {
	brotliIndexHandler, err := s.BrotliTempl(ViewIndex())
	if err != nil {
		log.Panicf("brotli: %v", err)
	}
	return func(w http.ResponseWriter, r *http.Request) error {
		if r.URL.Path != "/" {
			return ErrPageNotFound
		}
		ctx := r.Context()
		switch r.Method {
		default:
			return ErrMethodNotAllowed
		case http.MethodGet:
			return brotliIndexHandler(w, r)
		case http.MethodPost:
			b := make([]byte, 8)
			_, err := rand.Read(b)
			if err != nil {
				return fmt.Errorf("rand.Read: %w", err)
			}
			tsID := strings.ToLower(hex.EncodeToString(b))
			err = s.writeDB.InsertTS(ctx, tsID)
			if err != nil {
				return err
			}
			targetURL := fmt.Sprintf("/tax-simulation/%s#%s", tsID, income)
			http.Redirect(w, r, targetURL, http.StatusFound)
			return nil
		}
	}
}

func (s *Server) TaxSimulationQRCodeHandler() httpHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		ctx := r.Context()
		tsID := r.PathValue("tsID")

		_, err := s.readDB.GetTS(ctx, tsID)
		if err != nil {
			return fmt.Errorf("get: %w", err)
		}

		png, err := qrcode.Encode(
			fmt.Sprintf("https://%s/tax-simulation/%s#income", s.Host, tsID),
			qrcode.Highest,
			512,
		)
		if err != nil {
			return fmt.Errorf("qrcode.Encode[%T]: %w", err, err)
		}

		w.Header().Set(ContentType, TextHtml)
		w.Header().Set(CacheControl, `public, max-age=31536000`)
		err = ImgQRCode(base64.StdEncoding.EncodeToString(png)).Render(ctx, w)
		if err != nil {
			return fmt.Errorf("imgQRCode[%T]: %w", err, err)
		}
		return nil
	}
}

func (s *Server) TaxSimulationHandler() httpHandlerFunc {
	brotliIndexHandler, err := s.BrotliTempl(ViewPage())
	if err != nil {
		log.Panicf("brotli: %v", err)
	}
	var onlineCount int = 0
	return func(w http.ResponseWriter, r *http.Request) error {
		ctx := r.Context()

		tsID := r.PathValue("tsID")
		_, err := s.readDB.GetTS(ctx, tsID)
		if err != nil {
			return fmt.Errorf("get: %w", err)
		}

		switch r.Method {
		default:
			return ErrMethodNotAllowed

		case http.MethodGet:
			return brotliIndexHandler(w, r)

		case http.MethodPost:
			inputMap := map[string]string{}
			err = json.NewDecoder(r.Body).Decode(&inputMap)
			if err != nil {
				return fmt.Errorf("json.Decode: %w", err)
			}
			defer r.Body.Close()

			if !slices.Contains(pageList, inputMap[page]) {
				return ErrPageNotFound
			}

			switch inputMap[page] {
			case income:
				err = s.writeDB.UpdateTSIncome(ctx, UpdateTSIncomeParams{
					GrossSalary:     formInt(inputMap[grossSalary], maxSalary),
					KtgBeitrag:      formInt(inputMap[ktgBeitrag], maxBeitrag),
					BvgBeitrag:      formInt(inputMap[bvgBeitrag], maxBeitrag),
					TaxableSalary:   formInt(inputMap[taxableSalary], maxSalary),
					Year:            formInt(inputMap[year], defaultYear),
					CommuneID:       formInt(inputMap[communeId], zurichCommuneID),
					TaxSimulationID: tsID,
				})
			case investment:
				err = s.writeDB.UpdateTSInvestment(ctx, UpdateTSInvestmentParams{
					Investment:      formInt(inputMap[investmentID], maxSalary),
					TaxSimulationID: tsID,
				})
			case deduction:
				err = s.writeDB.UpdateTSDeduction(ctx, UpdateTSDeductionParams{
					DeductionOther:           formInt(inputMap[deductionOther], maxSalary),
					DeductionTransport:       formInt(inputMap[deductionTransport], maxSalary),
					DeductionProfession:      formInt(inputMap[deductionProfession], maxSalary),
					DeductionThirdPillar:     formInt(inputMap[deductionThirdPillar], maxSalary),
					DeductionHealthInsurance: formInt(inputMap[deductionHealthInsurance], maxSalary),
					DeductionMeal:            formInt(inputMap[deductionMeal], maxSalary),
					TaxSimulationID:          tsID,
				})
			}

			if err != nil {
				return fmt.Errorf("update: %w", err)
			}
			return nil

		case http.MethodPatch:
			w.Header().Set(ContentEncoding, "br")

			b := brotli.NewWriterLevel(w, 3)
			// b := brotli.NewWriterLevel(&countPrinter{w: w}, 3)
			sse, err := StartSSE(w, r)
			if err != nil {
				return fmt.Errorf("startSSE: %w", err)
			}
			sse.rc.SetWriteDeadline(time.Time{})
			writer := b

			var viewOnlineCount = 0
			var viewTs *TaxSimulation
			onlineCount++
			defer func() { onlineCount-- }()
			ticker := time.NewTicker(100 * time.Millisecond)
			var browserUpdates []*browser.Update
			for {
				if viewOnlineCount != onlineCount {
					viewOnlineCount = onlineCount
					browserUpdates = append(browserUpdates, &browser.Update{
						Selector: "#" + onlineCountID,
						Integer: map[string]int{
							browser.TEXT_CONTENT: viewOnlineCount,
						},
					})
				}

				ts, err := s.readDB.GetTS(ctx, tsID)
				if err != nil {
					return fmt.Errorf("get: %w", err)
				}
				ts.AhvBeitrag = taxes.ContributionRatesByYear[ts.Year].AHV
				ts.AlvBeitrag = taxes.ContributionRatesByYear[ts.Year].ALV
				if viewTs == nil || ts.Version != viewTs.Version {
					browserUpdates = append(browserUpdates, ts.toBrowserUpdates()...)
				}
				viewTs = &ts

				if len(browserUpdates) > 0 {
					err = sse.WriteMessage(ctx, writer, &SSEventMessage{
						BrowserUpdates: browserUpdates,
					})
					if err != nil {
						return fmt.Errorf("sse.WriteMessage: %w", err)
					}
					browserUpdates = []*browser.Update{}

					err = sse.WriteMessage(ctx, writer, &SSEventMessage{
						Templ: ViewResult(viewTs),
					})
					if err != nil {
						return fmt.Errorf("sse.WriteMessage: %w", err)
					}
				}

				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-s.shutdownChan:
					return s.HijackAndClose(w)
				case <-ticker.C:
				}
			}

		}
	}
}

var _ io.Writer = (*countPrinter)(nil)

type countPrinter struct {
	w io.Writer
}

func (c *countPrinter) Write(b []byte) (int, error) {
	fmt.Println("count: ", len(b))
	return c.w.Write(b)
}

func (ts *TaxSimulation) taxableSalaryNoInvsNoDeds() float32 {
	taxableSalary := float32(ts.TaxableSalary)
	if taxableSalary == 0 {
		taxableSalary = float32(ts.GrossSalary) * (1 - float32(ts.AhvBeitrag+ts.AlvBeitrag+ts.KtgBeitrag+ts.BvgBeitrag)/10000.0)
	}
	if taxableSalary < 0 {
		return 0
	}
	return taxableSalary
}

func (ts *TaxSimulation) deductionValue(deduction string) (float32, float32) {
	deductionMap := map[string]int{
		deductionTransport:       ts.DeductionTransport,
		deductionMeal:            ts.DeductionMeal,
		deductionProfession:      ts.DeductionProfession,
		deductionThirdPillar:     ts.DeductionThirdPillar,
		deductionHealthInsurance: ts.DeductionHealthInsurance,
	}
	value, ok := deductionMap[deduction]
	if !ok {
		return 0, 0
	}
	dl, ok := deductionLimitByYear[ts.Year]
	if !ok {
		return 0, 0
	}
	d, ok := dl[deduction]
	if !ok {
		return 0, 0
	}
	return min(float32(value), d[0]), min(float32(value), d[1])
}

func (ts *TaxSimulation) taxableSalaryZurich() float32 {
	out := ts.taxableSalaryNoInvsNoDeds() + float32(ts.Investment-ts.DeductionOther)
	for _, deduction := range []string{
		deductionTransport,
		deductionMeal,
		deductionProfession,
		deductionThirdPillar,
		deductionHealthInsurance,
	} {
		cantonal, _ := ts.deductionValue(deduction)
		out -= cantonal
	}
	if out < 0 {
		return 0
	}
	return out
}

func (ts *TaxSimulation) taxableSalaryFederal() float32 {
	out := ts.taxableSalaryNoInvsNoDeds() + float32(ts.Investment-ts.DeductionOther)
	for _, deduction := range []string{
		deductionTransport,
		deductionMeal,
		deductionProfession,
		deductionThirdPillar,
		deductionHealthInsurance,
	} {
		_, federal := ts.deductionValue(deduction)
		out -= federal
	}
	if out < 0 {
		return 0
	}
	return out
}

func (ts *TaxSimulation) toBrowserUpdates() (out []*browser.Update) {
	valueByID := map[string]int{
		grossSalary:              ts.GrossSalary,
		ktgBeitrag:               ts.KtgBeitrag,
		bvgBeitrag:               ts.BvgBeitrag,
		taxableSalary:            ts.TaxableSalary,
		year:                     ts.Year,
		communeId:                ts.CommuneID,
		deductionTransport:       ts.DeductionTransport,
		deductionProfession:      ts.DeductionProfession,
		deductionThirdPillar:     ts.DeductionThirdPillar,
		deductionHealthInsurance: ts.DeductionHealthInsurance,
		deductionMeal:            ts.DeductionMeal,
		deductionOther:           ts.DeductionOther,
		investmentID:             ts.Investment,
	}
	for id, value := range valueByID {
		out = append(out, &browser.Update{
			Selector: "#" + id,
			Integer:  map[string]int{browser.VALUE: value},
		})
	}
	return out
}

const (
	maxSalary       = 500000
	maxBeitrag      = 1000
	zurichCommuneID = 261
	defaultYear     = 2025
)

func formInt(value string, maxValue int) int {
	v, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}
	if v > maxValue {
		return maxValue
	}
	if v < 0 {
		return 0
	}
	return v
}
