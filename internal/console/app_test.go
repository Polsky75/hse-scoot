package console_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/Polsky75/hse-scoot/internal/console"
	"github.com/Polsky75/hse-scoot/internal/domain"
	"github.com/Polsky75/hse-scoot/internal/fleet"
	"github.com/Polsky75/hse-scoot/internal/report"
	"github.com/Polsky75/hse-scoot/internal/source"
)

type fakeSource struct {
	dataset source.Dataset
	err     error
}

func (s fakeSource) Load() (source.Dataset, error) {
	return s.dataset, s.err
}

var errInspection = errors.New("техосмотр не пройден (fake)")

// fakeInspector — тестовый дубль техосмотра с управляемым отказом.
type fakeInspector struct {
	rejectInventoryNumber string
}

func (f fakeInspector) Inspect(device domain.IInventoriable, _ int) error {
	if device.GetInventoryNumber() == f.rejectInventoryNumber {
		return errInspection
	}
	return nil
}

func newApp(src source.Source, reportOut, statusOut *bytes.Buffer) *console.App {
	return console.New(
		fleet.NewService(src, fakeInspector{rejectInventoryNumber: "SCOOT-002"}),
		report.NewConsoleOutput(reportOut),
		statusOut,
	)
}

func TestRunWritesStatusAndReport(t *testing.T) {
	var reportOut, statusOut bytes.Buffer
	app := newApp(fakeSource{dataset: source.Dataset{
		Registered: []source.DeviceSpec{
			{Kind: domain.KindScooter, InventoryNumber: "SCOOT-001", DailyConsumptionKWh: 1.5, Condition: 95, EaseOfUse: 8},
		},
		Incoming: []source.DeviceSpec{
			{Kind: domain.KindEBike, InventoryNumber: "EBIKE-001", DailyConsumptionKWh: 2, Condition: 80, EaseOfUse: 5},
			{Kind: domain.KindScooter, InventoryNumber: "SCOOT-002", DailyConsumptionKWh: 1, Condition: 40, EaseOfUse: 9},
			{Kind: domain.KindBicycle, InventoryNumber: "BIKE-002", Condition: 90, EaseOfUse: 6},
		},
		Equipment: []source.EquipmentSpec{
			{Kind: domain.KindHelmet, InventoryNumber: "HELMET-001"},
			{Kind: domain.KindKneePads, InventoryNumber: "KNEE-001"},
			{Kind: domain.KindChargingStation, InventoryNumber: "STATION-001", DailyConsumptionKWh: 3.2},
		},
	}}, &reportOut, &statusOut)

	if err := app.Run(); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	status := statusOut.String()
	for _, want := range []string{
		"Постановка на учёт:",
		"принят",
		"EBIKE-001 (ebike)",
		"отклонён",
		"SCOOT-002 (scooter)",
		"техосмотр не пройден",
	} {
		if !strings.Contains(status, want) {
			t.Errorf("status does not contain %q:\n%s", want, status)
		}
	}

	reportText := reportOut.String()
	for _, want := range []string{
		"hse-scoot — отчёт по инвентарю",
		"Всего вещей на учёте: 6",
		"Оборудование: 3",
		"Транспорт для профессионалов: 2",
		"Транспорт для новичков: 1",
		"Самокаты: 1",
		"Велосипеды: 1",
		"Электровелосипеды: 1",
		"Шлемы: 1",
		"Наколенники: 1",
		"Зарядные станции: 1",
		"Суммарное суточное потребление: 6.70 кВт·ч",
		"Всё оборудование (3):",
		"Транспорт для профессионалов (2):",
		"Транспорт для новичков (1):",
		"SCOOT-001 — самокат",
		"HELMET-001 — шлем",
		"KNEE-001 — наколенники",
		"BIKE-002 — велосипед",
	} {
		if !strings.Contains(reportText, want) {
			t.Errorf("report does not contain %q:\n%s", want, reportText)
		}
	}
}

func TestRunWritesEmptyInventory(t *testing.T) {
	var reportOut, statusOut bytes.Buffer
	if err := newApp(fakeSource{}, &reportOut, &statusOut).Run(); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if !strings.Contains(statusOut.String(), "нет новых устройств") {
		t.Errorf("status does not contain empty registration message:\n%s", statusOut.String())
	}

	reportText := reportOut.String()
	for _, unwanted := range []string{
		"Самокаты: 0", "Велосипеды: 0", "Электровелосипеды: 0",
		"Шлемы: 0", "Наколенники: 0", "Зарядные станции: 0",
	} {
		if strings.Contains(reportText, unwanted) {
			t.Errorf("report contains zero category line %q:\n%s", unwanted, reportText)
		}
	}
	for _, want := range []string{
		"Всего вещей на учёте: 0",
		"Оборудование: 0",
		"Транспорт для профессионалов: 0",
		"Транспорт для новичков: 0",
		"Суммарное суточное потребление: 0.00 кВт·ч",
		"Всё оборудование (0):",
	} {
		if !strings.Contains(reportText, want) {
			t.Errorf("report does not contain %q:\n%s", want, reportText)
		}
	}
}

func TestRunReturnsSourceError(t *testing.T) {
	err := newApp(fakeSource{err: errors.New("source failed")}, &bytes.Buffer{}, &bytes.Buffer{}).Run()
	if err == nil || !strings.Contains(err.Error(), "sync inventory") {
		t.Fatalf("Run() error = %v, want wrapped sync error", err)
	}
}

type failingOutput struct{}

func (failingOutput) Write(report.Document) error {
	return errors.New("output failed")
}

func TestRunReturnsReportError(t *testing.T) {
	app := console.New(
		fleet.NewService(fakeSource{}, fakeInspector{}),
		failingOutput{},
		&bytes.Buffer{},
	)

	if err := app.Run(); err == nil || !strings.Contains(err.Error(), "write report") {
		t.Fatalf("Run() error = %v, want wrapped report error", err)
	}
}
