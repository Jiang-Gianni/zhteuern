package main

// go run taxes/cmd/*

import (
	"errors"
	"log"
)

const year = "25"

func main() {
	if err := run(); err != nil {
		log.Fatalf("run: %v", err)
	}
}

func run() error {
	return errors.Join(
		runQuellenSteuer(),
		runEstvIncomeRate(),
	)
}
