package main

import (
	"bufio"
	_ "embed"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"text/template"
)

type EstvIncomeRate struct {
	CommuneID         int
	CommuneName       string
	CommuneMultiplier int
}

const cantonalRateString = "	98.00	"

func runEstvIncomeRate() (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("runQuellenSteuer: %w", err)
		}
	}()
	f, err := os.Open(fmt.Sprintf("./taxes/estv_income_rates_20%s.txt", year))
	if err != nil {
		return fmt.Errorf("os.Open: %v", err)
	}
	defer f.Close()

	eirList := []*EstvIncomeRate{}
	var row string
	b := bufio.NewScanner(f)
	for b.Scan() {
		row = strings.TrimSpace(b.Text())
		splits := strings.Split(row, cantonalRateString)
		if len(splits) != 2 {
			return fmt.Errorf("expected 2 splits, got %d", len(splits))
		}
		communeMultiplier, err := strconv.Atoi(strings.TrimSuffix(splits[1], ".00"))
		if err != nil {
			return fmt.Errorf("strconv.Atoi communeRate: %w", err)
		}
		communes := strings.Split(splits[0], "	")
		if len(communes) != 2 {
			return fmt.Errorf("expected 2 commune splits, got %d", len(communes))
		}
		communeID, err := strconv.Atoi(strings.TrimSpace(communes[0]))
		if err != nil {
			return fmt.Errorf("strconv.Atoi communeID: %w", err)
		}
		eirList = append(eirList, &EstvIncomeRate{
			CommuneID:         communeID,
			CommuneName:       strings.TrimSpace(communes[1]),
			CommuneMultiplier: communeMultiplier,
		})
	}
	return templateEstvIncomeRate(eirList)
}

//go:embed template/estv_income_rates.go
var templateEstvIncomeRatesGo string

func templateEstvIncomeRate(eirList []*EstvIncomeRate) error {
	t, err := template.New("estv_income_rates").Parse(templateEstvIncomeRatesGo)
	if err != nil {
		return fmt.Errorf("template.Parse: %w", err)
	}

	fileName := fmt.Sprintf("./taxes/estv_income_rates_%s.go", year)
	if err := os.Remove(fileName); err != nil {
		slog.Warn("os.Remove", "err", err)
	}
	f, err := os.Create(fileName)
	if err != nil {
		return fmt.Errorf("os.Create: %v", err)
	}
	defer f.Close()

	if err := t.Execute(f, map[string]any{
		"Year": year,
		"List": eirList,
	}); err != nil {
		return fmt.Errorf("t.Execute: %w", err)
	}

	return nil
}
