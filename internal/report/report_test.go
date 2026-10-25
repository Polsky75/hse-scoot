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

func equipment(t *testing.T, kind domain.EquipmentKind, number string, consumption float64) domain.IInventoriable {
	t.Helper()
	value, err := domain.NewEquipment(kind, number, consumption)
	if err != nil {
		t.Fatalf("NewEquipment(%s) error = %v", number, err)
	}
	return value
}

func assertCounts(t *testing.T, got, want report.Counts) {
	t.Helper()
	if got.Total != want.Total {
		t.Errorf("Total = %d, want %d", got.Total, want.Total)
	}
	if got.Transport != want.Transport {
		t.Errorf("Transport = %d, want %d", got.Transport, want.Transport)
	}
	if got.Equipment != want.Equipment {
		t.Errorf("Equipment = %d, want %d", got.Equipment, want.Equipment)
	}
	if !slices.Equal(got.Categories, want.Categories) {
		t.Errorf("Categories = %+v, want %+v", got.Categories, want.Categories)
	}
	if got.TotalDailyConsumptionKWh != want.TotalDailyConsumptionKWh {
		t.Errorf(
			"TotalDailyConsumptionKWh = %v, want %v",
			got.TotalDailyConsumptionKWh, want.TotalDailyConsumptionKWh,
		)
	}
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
		equipment(t, domain.KindHelmet, "HELMET-001", 0),
		equipment(t, domain.KindKneePads, "KNEE-001", 0),
		equipment(t, domain.KindChargingStation, "STATION-001", 2),
	)
	return items, beginner
}

func TestBuildCountsAndSections(t *testing.T) {
	items, beginner := testInventory(t)
	document := report.Build(items, beginner)

	wantCounts := report.Counts{
		Total:     7,
		Transport: 4,
		Equipment: 3,
		Categories: []report.CategoryCount{
			{Category: domain.CategoryScooter, Count: 2},
			{Category: domain.CategoryBicycle, Count: 1},
			{Category: domain.CategoryEBike, Count: 1},
			{Category: domain.CategoryHelmet, Count: 1},
			{Category: domain.CategoryKneePads, Count: 1},
			{Category: domain.CategoryChargingStation, Count: 1},
		},
		TotalDailyConsumptionKWh: 5,
	}
	assertCounts(t, document.Counts, wantCounts)

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
	if got := report.CountEquipment(items); got != 3 {
		t.Errorf("CountEquipment() = %d, want 3", got)
	}
	if got := report.CountCategory(items, domain.CategoryScooter); got != 2 {
		t.Errorf("CountCategory(scooter) = %d, want 2", got)
	}
	if got := report.CountCategory(items, domain.CategoryBicycle); got != 1 {
		t.Errorf("CountCategory(bicycle) = %d, want 1", got)
	}
	if got := report.CountCategory(items, domain.CategoryEBike); got != 1 {
		t.Errorf("CountCategory(ebike) = %d, want 1", got)
	}
	if got := report.CountCategory(items, domain.CategoryKneePads); got != 1 {
		t.Errorf("CountCategory(kneepads) = %d, want 1", got)
	}
	if got := report.CountCategory(items, domain.CategoryHelmet); got != 1 {
		t.Errorf("CountCategory(helmet) = %d, want 1", got)
	}
	wantCategories := []report.CategoryCount{
		{Category: domain.CategoryScooter, Count: 2},
		{Category: domain.CategoryBicycle, Count: 1},
		{Category: domain.CategoryEBike, Count: 1},
		{Category: domain.CategoryHelmet, Count: 1},
		{Category: domain.CategoryKneePads, Count: 1},
		{Category: domain.CategoryChargingStation, Count: 1},
	}
	if got := report.CountCategories(items); !slices.Equal(got, wantCategories) {
		t.Errorf("CountCategories() = %+v, want %+v", got, wantCategories)
	}
	if got := report.CountDailyConsumptionKWh(items); got != 5 {
		t.Errorf("CountDailyConsumptionKWh() = %v, want 5", got)
	}
}

func TestRenderTextContainsCountsAndCategories(t *testing.T) {
	items, beginner := testInventory(t)
	text := report.RenderText(report.Build(items, beginner))

	for _, want := range []string{
		"hse-scoot — отчёт по инвентарю",
		"Всего вещей на учёте: 7",
		"Транспорт: 4",
		"Оборудование: 3",
		"Самокаты: 2",
		"Велосипеды: 1",
		"Электровелосипеды: 1",
		"Шлемы: 1",
		"Наколенники: 1",
		"Зарядные станции: 1",
		"Суммарное суточное потребление: 5.00 кВт·ч",
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

func TestChargingStationCountedAsEquipment(t *testing.T) {
	items := []domain.IInventoriable{
		equipment(t, domain.KindChargingStation, "STATION-001", 3.2),
	}
	document := report.Build(items, nil)

	if document.Counts.Total != 1 || document.Counts.Transport != 0 || document.Counts.Equipment != 1 {
		t.Errorf("counts = %+v, want total=1 transport=0 equipment=1", document.Counts)
	}
	if len(document.Equipment) != 1 || len(document.Transport) != 0 {
		t.Errorf("equipment/transport lists = %d/%d, want 1/0", len(document.Equipment), len(document.Transport))
	}
	wantCategories := []report.CategoryCount{
		{Category: domain.CategoryChargingStation, Count: 1},
	}
	if !slices.Equal(document.Counts.Categories, wantCategories) {
		t.Errorf("categories = %+v, want %+v", document.Counts.Categories, wantCategories)
	}

	text := report.RenderText(document)
	for _, want := range []string{"Оборудование: 1", "Зарядные станции: 1", "STATION-001 — зарядная станция"} {
		if !strings.Contains(text, want) {
			t.Errorf("text does not contain %q:\n%s", want, text)
		}
	}
	for _, unwanted := range []string{
		"Наколенники: 0", "Самокаты: 0", "Велосипеды: 0",
		"Электровелосипеды: 0", "Шлемы: 0",
	} {
		if strings.Contains(text, unwanted) {
			t.Errorf("text contains zero category line %q:\n%s", unwanted, text)
		}
	}
}

type customItem struct {
	number   string
	category domain.ItemCategory
}

func (i customItem) GetInventoryNumber() string       { return i.number }
func (i customItem) GetCategory() domain.ItemCategory { return i.category }

func TestNewCategoryAppearsInStatisticsWithoutReportChanges(t *testing.T) {
	const category domain.ItemCategory = "cargo_bike"
	items := []domain.IInventoriable{
		customItem{number: "CARGO-001", category: category},
	}

	document := report.Build(items, nil)
	wantCategories := []report.CategoryCount{{Category: category, Count: 1}}
	if !slices.Equal(document.Counts.Categories, wantCategories) {
		t.Errorf("categories = %+v, want %+v", document.Counts.Categories, wantCategories)
	}
	if document.Counts.Equipment != 1 || document.Counts.Transport != 0 {
		t.Errorf("counts = %+v, want equipment=1 transport=0", document.Counts)
	}

	text := report.RenderText(document)
	if !strings.Contains(text, "cargo_bike: 1") {
		t.Errorf("text does not contain new category count:\n%s", text)
	}
}
