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
	if got.Equipment != want.Equipment {
		t.Errorf("Equipment = %d, want %d", got.Equipment, want.Equipment)
	}
	if got.ProfessionalTransport != want.ProfessionalTransport {
		t.Errorf(
			"ProfessionalTransport = %d, want %d",
			got.ProfessionalTransport, want.ProfessionalTransport,
		)
	}
	if got.BeginnerTransport != want.BeginnerTransport {
		t.Errorf("BeginnerTransport = %d, want %d", got.BeginnerTransport, want.BeginnerTransport)
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
		item(t, domain.KindScooter, "SCOOT-004", 7),
	}
	items := slices.Clone(beginner)
	items = append(items,
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
		Total:                 7,
		Equipment:             3,
		ProfessionalTransport: 1,
		BeginnerTransport:     3,
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
	if len(document.ProfessionalTransport) != 1 {
		t.Errorf("professional transport = %d, want 1", len(document.ProfessionalTransport))
	}
	if got := document.ProfessionalTransport[0]; got.InventoryNumber != "EBIKE-001" {
		t.Errorf("professional[0] = %+v, want EBIKE-001", got)
	}
	if len(document.BeginnerTransport) != 3 {
		t.Errorf("beginner transport = %d, want 3", len(document.BeginnerTransport))
	}
	if got := document.BeginnerTransport[0]; got.InventoryNumber != "SCOOT-001" || got.Category != domain.CategoryScooter {
		t.Errorf("beginner[0] = %+v, want SCOOT-001/scooter", got)
	}
}

func TestCountCategoriesAndConsumption(t *testing.T) {
	items, _ := testInventory(t)

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
		"Оборудование: 3",
		"Транспорт для профессионалов: 1",
		"Транспорт для новичков: 3",
		"Самокаты: 2",
		"Велосипеды: 1",
		"Электровелосипеды: 1",
		"Шлемы: 1",
		"Наколенники: 1",
		"Зарядные станции: 1",
		"Суммарное суточное потребление: 5.00 кВт·ч",
		"Оборудование (3):",
		"Транспорт для профессионалов (1):",
		"Транспорт для новичков (3):",
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

	if document.Counts.Total != 1 || document.Counts.Equipment != 1 ||
		document.Counts.ProfessionalTransport != 0 || document.Counts.BeginnerTransport != 0 {
		t.Errorf("counts = %+v, want total=1 equipment=1 transport=0/0", document.Counts)
	}
	if len(document.Equipment) != 1 ||
		len(document.ProfessionalTransport) != 0 || len(document.BeginnerTransport) != 0 {
		t.Errorf(
			"equipment/professional/beginner = %d/%d/%d, want 1/0/0",
			len(document.Equipment), len(document.ProfessionalTransport), len(document.BeginnerTransport),
		)
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
	if document.Counts.Equipment != 1 ||
		document.Counts.ProfessionalTransport != 0 || document.Counts.BeginnerTransport != 0 {
		t.Errorf("counts = %+v, want equipment=1 transport=0/0", document.Counts)
	}

	text := report.RenderText(document)
	if !strings.Contains(text, "cargo_bike: 1") {
		t.Errorf("text does not contain new category count:\n%s", text)
	}
}
