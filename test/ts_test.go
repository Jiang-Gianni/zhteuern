package test

import (
	"bytes"
	"encoding/base64"
	"strings"
	"testing"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/input"
	"github.com/skip2/go-qrcode"
	"github.com/ysmood/gson"
)

const (
	host = "http://localhost:3000"
)

// go test ./test -rod=show
func TestTaxSimulation(t *testing.T) {
	browser := rod.New().MustConnect()
	defer browser.MustClose()

	page1 := browser.MustPage(host).MustWaitDOMStable()

	// Start Tax Simulation button
	page1.MustElement("button[key-down='s']")
	if err := page1.Keyboard.Press(input.KeyS); err != nil {
		t.Fatalf("s key press: %v", err)
	}
	page1.MustWaitDOMStable()

	i := page1.MustInfo()
	tsID := strings.TrimSuffix(
		strings.TrimPrefix(i.URL, host+"/tax-simulation/"),
		"#income",
	)
	if len(tsID) != 16 {
		t.Fatalf("expected tsID of length 8 for url %s", i.URL)
	}

	// Check QR Code
	if err := page1.Keyboard.Press(input.KeyQ); err != nil {
		t.Fatalf("q key press: %v", err)
	}
	page1.MustWaitDOMStable()
	qrCodeSrc, err := page1.MustElement("img#qr_code_image").Attribute("src")
	if err != nil {
		t.Fatalf("img qr code: %v", err)
	}
	if qrCodeSrc == nil || len(*qrCodeSrc) == 0 {
		t.Fatalf("nil or empty src attribute for qr code img")
	}
	png, err := qrcode.Encode(i.URL, qrcode.Highest, 512)
	if err != nil {
		t.Fatalf("qrcode.Encode[%T]: %v", err, err)
	}
	gotQrCode, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(*qrCodeSrc, "data:image/png;base64,"))
	if err != nil {
		t.Fatalf("decode base64 qr code[%T]: %v", err, err)
	}
	if bytes.Equal(png, gotQrCode) {
		t.Fatalf("unexpected qr code base64")
	}

	// Check live SSE updates on the same tax simulation on year change
	yearSelector := "select#year"
	yearVerify := &verify{selector: yearSelector, property: "value"}
	page2 := browser.MustPage(i.URL).MustWaitDOMStable()
	year := verifyStringProp(t, yearVerify, page1, page2)
	if year.String() != "2025" {
		t.Fatalf("expected year 2025 as default but got %s", year.String())
	}

	page2.MustElement(yearSelector).MustSelect("2024")
	page2.MustWaitDOMStable()
	year = verifyStringProp(t, yearVerify, page1, page2)
	if year.String() != "2024" {
		t.Fatalf("expected year 2024 after update but got %s", year.String())
	}
}

type verify struct {
	selector string
	property string
}

func verifyStringProp(t *testing.T, v *verify, pages ...*rod.Page) gson.JSON {
	var value string
	var gsonJson gson.JSON
	var err error
	for _, p := range pages {
		gsonJson, err = p.MustElement(v.selector).Property(v.property)
		if err != nil {
			t.Fatalf("%s selector, %s property: %v", v.selector, v.property, err)
		}
		if value == "" {
			value = gsonJson.String()
		}
		if value != gsonJson.String() {
			t.Fatalf("%s selector, %s property: %s and %s are not equal", v.selector, v.property, value, gsonJson.String())
		}
	}
	return gsonJson
}
