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

// I'll only scan for A0N (not married, 0 children, no church)
type QuellenSteuer struct {
	Start              int
	PercentageTimes100 int // ex 0.25% becomes 25
}

func runQuellenSteuer() (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("runQuellenSteuer: %w", err)
		}
	}()
	f, err := os.Open(fmt.Sprintf("./taxes/tar%szh.txt", year))
	if err != nil {
		return fmt.Errorf("os.Open: %v", err)
	}
	defer f.Close()

	var qsList []*QuellenSteuer
	var row string
	b := bufio.NewScanner(f)
	scannedFirst := false
	for b.Scan() {
		row = strings.TrimSpace(b.Text())
		if len(row) != 59 {
			continue
		}
		if row[6:9] != "A0N" {
			continue
		}
		perc, err := strconv.Atoi(row[55:])
		if err != nil {
			return fmt.Errorf("strconv.Atoi: %w", err)
		}
		start, err := strconv.Atoi(row[26:31])
		if err != nil {
			return fmt.Errorf("strconv.Atoi: %w", err)
		}
		if start == 1 {
			if !scannedFirst {
				start = 0 // first row
				scannedFirst = true
			} else {
				break // last row
			}
		}
		qsList = append(qsList, &QuellenSteuer{
			Start:              start,
			PercentageTimes100: perc,
		})
	}
	return templateQuellenSteuer(qsList)
}

//go:embed template/quellensteuer.go
var templateQuellenSteuerGo string

func templateQuellenSteuer(qsList []*QuellenSteuer) error {
	t, err := template.New("quellensteuer").Parse(templateQuellenSteuerGo)
	if err != nil {
		return fmt.Errorf("template.Parse: %w", err)
	}

	fileName := fmt.Sprintf("./taxes/quellensteuer_%s.go", year)
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
		"List": qsList,
	}); err != nil {
		return fmt.Errorf("t.Execute: %w", err)
	}

	return nil
}
