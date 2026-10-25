package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/Polsky75/hse-scoot/internal/di"
	"github.com/Polsky75/hse-scoot/internal/report"
)

func main() {
	devicesPath := flag.String(
		"devices",
		"data/devices.txt",
		"путь к текстовому файлу с инвентарём",
	)
	reportTarget := flag.String(
		"report",
		"console",
		"куда вывести отчёт: console или file",
	)
	reportPath := flag.String(
		"report-path",
		"report.txt",
		"путь к файлу отчёта при -report=file",
	)
	flag.Parse()

	output, err := report.NewOutput(*reportTarget, *reportPath, os.Stdout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "hse-scoot: %v\n", err)
		os.Exit(1)
	}

	container, err := di.New(os.Stdout, *devicesPath, output)
	if err != nil {
		fmt.Fprintf(os.Stderr, "hse-scoot: %v\n", err)
		os.Exit(1)
	}

	if err := container.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "hse-scoot: %v\n", err)
		os.Exit(1)
	}
}
