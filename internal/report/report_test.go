package report_test

import (
	"slices"
	"strings"
	"testing"

	"github.com/Polsky75/hse-scoot/internal/domain"
	"github.com/Polsky75/hse-scoot/internal/report"
)

func item(t *testing.T, kind domain.TransportKind, number string, ease int) domain.IInventoriable {
	t.Helper()
	consumption := 1.0
	if kind == domain.KindBicycle {
		consumption = 0
	}
	value, err := domain.NewTransport(kind, number, consumption, ease)
	if err != nil {
		t.Fatalf("NewTransport(%s) error = %v", number, err)
	}
	return value
}

func equipment(t *testing.T, kind domain.EquipmentKind, number string) domain.IInventoriable {
	t.Helper()
	value, err := domain.NewEquipment(kind, number)
	if err != nil {
		t.Fatalf("NewEquipment(%s) error = %v", number, err)
	}
	return value
}

func testInventory(t *testing.T) ([]domain.IInventoriable, []domain.IInventoriable) {
	t.Helper()
	beginner := []domain.IInventoriable{
		item(t, domain.KindScooter, "SCOOT-001", 8),
		item(t, domain.KindBicycle, "BIKE-001", 9),
	}
	items := slices.Clone(beginner)
	items = append(items,
		item(t, domain.KindScooter, "SCOOT-004", 7),
		item(t, domain.KindEBike, "EBIKE-001", 5),
		equipment(t, domain.KindHelmet, "HELMET-001"),
		equipment(t, domain.KindKneePads, "KNEE-001"),
		equipment(t, domain.KindChargingStation, "STATION-001"),
	)
	return items, beginner
}

func TestBuildCountsAndSections(t *testing.T) {
	items, beginner := testInventory(t)
	document := report.Build(items, beginner)

	wantCounts := report.Counts{
		Total:                    7,
		Transport:                4,
		Scooters:                 2,
		KneePads:                 1,
		TotalDailyConsumptionKWh: 3,
	}
	if got := document.Counts; got != wantCounts {
		t.Errorf("counts = %+v, want %+v", got, wantCounts)
	}
	if len(document.Equipment) != 3 {
		t.Fatalf("equipment = %d, want 3", len(document.Equipment))
	}
	for _, item := range document.Equipment {
		switch item.Category {
		case domain.CategoryHelmet, domain.CategoryKneePads, domain.CategoryChargingStation:
		default:
			t.Errorf("equipment section contains non-equipment item: %+v", item)
		}
	}
	if len(document.Transport) != 4 {
		t.Errorf("transport = %d, want 4", len(document.Transport))
	}
	if len(document.BeginnerTransport) != 2 {
		t.Errorf("beginner transport = %d, want 2", len(document.BeginnerTransport))
	}
	if got := document.BeginnerTransport[0]; got.InventoryNumber != "SCOOT-001" || got.Category != domain.CategoryScooter {
		t.Errorf("beginner[0] = %+v, want SCOOT-001/scooter", got)
	}
}

func TestCountFunctions(t *testing.T) {
	items, _ := testInventory(t)

	if got := report.CountAll(items); got != 7 {
		t.Errorf("CountAll() = %d, want 7", got)
	}
	if got := report.CountTransport(items); got != 4 {
		t.Errorf("CountTransport() = %d, want 4", got)
	}
	if got := report.CountCategory(items, domain.CategoryScooter); got != 2 {
		t.Errorf("CountCategory(scooter) = %d, want 2", got)
	}
	if got := report.CountCategory(items, domain.CategoryKneePads); got != 1 {
		t.Errorf("CountCategory(kneepads) = %d, want 1", got)
	}
	if got := report.CountCategory(items, domain.CategoryHelmet); got != 1 {
		t.Errorf("CountCategory(helmet) = %d, want 1", got)
	}
	if got := report.CountDailyConsumptionKWh(items); got != 3 {
		t.Errorf("CountDailyConsumptionKWh() = %v, want 3", got)
	}
}

func TestRenderTextContainsCountsAndCategories(t *testing.T) {
	items, beginner := testInventory(t)
	text := report.RenderText(report.Build(items, beginner))

	for _, want := range []string{
		"hse-scoot — отчёт по инвентарю",
		"Всего вещей на учёте: 7",
		"Транспорт: 4",
		"Самокаты: 2",
		"Наколенники: 1",
		"Суммарное суточное потребление: 3.00 кВт·ч",
		"Всё оборудование (3):",
		"Транспорт (4):",
		"Транспорт для новичков (2):",
		"SCOOT-001 — самокат",
		"BIKE-001 — велосипед",
		"EBIKE-001 — электровелосипед",
		"HELMET-001 — шлем",
		"KNEE-001 — наколенники",
		"STATION-001 — зарядная станция",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("text does not contain %q:\n%s", want, text)
		}
	}
}
