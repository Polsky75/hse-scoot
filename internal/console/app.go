// Package console запускает сценарий и выводит служебный статус.
package console

import (
	"fmt"
	"io"
	"strings"

	"github.com/Polsky75/hse-scoot/internal/fleet"
	"github.com/Polsky75/hse-scoot/internal/report"
)

// App загружает инвентарь, выполняет постановку на учёт и выводит отчёт.
type App struct {
	fleet  *fleet.Service
	output report.Output
	status io.Writer
}

// New создаёт приложение.
func New(fleetService *fleet.Service, output report.Output, status io.Writer) *App {
	return &App{
		fleet:  fleetService,
		output: output,
		status: status,
	}
}

// Run выполняет сценарий и выводит отчёт выбранным способом.
func (a *App) Run() error {
	results, err := a.fleet.Sync()
	if err != nil {
		return fmt.Errorf("sync inventory: %w", err)
	}

	if err := a.writeRegistrations(results); err != nil {
		return fmt.Errorf("write registration status: %w", err)
	}

	beginnerItems, _ := a.fleet.BeginnerDevices()
	document := report.Build(a.fleet.Items(), beginnerItems)
	if err := a.output.Write(document); err != nil {
		return fmt.Errorf("write report: %w", err)
	}
	return nil
}

func (a *App) writeRegistrations(results []fleet.RegistrationResult) error {
	if a.status == nil {
		return nil
	}

	var out strings.Builder
	out.WriteString("Постановка на учёт:\n")
	if len(results) == 0 {
		out.WriteString("  нет новых устройств\n")
	}
	for _, result := range results {
		if result.Accepted() {
			fmt.Fprintf(&out, "  %-8s %s (%s)\n", "принят", result.InventoryNumber, result.Kind)
			continue
		}
		fmt.Fprintf(
			&out,
			"  %-8s %s (%s): %v\n",
			"отклонён", result.InventoryNumber, result.Kind, result.Err,
		)
	}
	out.WriteString("\n")

	if _, err := io.WriteString(a.status, out.String()); err != nil {
		return err
	}
	return nil
}
