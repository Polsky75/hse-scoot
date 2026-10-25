package fleet_test

import (
	"errors"
	"slices"
	"testing"

	"github.com/Polsky75/hse-scoot/internal/domain"
	"github.com/Polsky75/hse-scoot/internal/fleet"
	"github.com/Polsky75/hse-scoot/internal/inspection"
	"github.com/Polsky75/hse-scoot/internal/source"
)

type fakeSource struct {
	dataset source.Dataset
	err     error
}

func (s fakeSource) Load() (source.Dataset, error) {
	return s.dataset, s.err
}

func newInspector(t *testing.T) *inspection.Inspector {
	t.Helper()
	inspector, err := inspection.NewInspector(inspection.DefaultMinCondition)
	if err != nil {
		t.Fatalf("NewInspector() error = %v", err)
	}
	return inspector
}

func TestSyncRestoresInventoryAndRegistersInspectedTransport(t *testing.T) {
	service := fleet.NewService(fakeSource{dataset: source.Dataset{
		Registered: []source.DeviceSpec{
			{Kind: domain.KindScooter, InventoryNumber: "SCOOT-001", DailyConsumptionKWh: 1.5, Condition: 95, EaseOfUse: 8},
			{Kind: domain.KindBicycle, InventoryNumber: "BIKE-001", Condition: 88, EaseOfUse: 9},
		},
		Incoming: []source.DeviceSpec{
			{Kind: domain.KindEBike, InventoryNumber: "EBIKE-001", DailyConsumptionKWh: 2, Condition: 70, EaseOfUse: 7},
			{Kind: domain.KindScooter, InventoryNumber: "SCOOT-002", DailyConsumptionKWh: 1, Condition: 69, EaseOfUse: 6},
			{Kind: domain.KindBicycle, InventoryNumber: "BIKE-002", Condition: 80, EaseOfUse: 8},
		},
		Equipment: []source.EquipmentSpec{
			{Kind: domain.KindHelmet, InventoryNumber: "HELMET-001"},
			{Kind: domain.KindChargingStation, InventoryNumber: "STATION-001"},
		},
	}}, newInspector(t))

	results, err := service.Sync()
	if err != nil {
		t.Fatalf("Sync() error = %v", err)
	}

	if len(results) != 3 {
		t.Fatalf("registration results = %d, want 3", len(results))
	}
	if !results[0].Accepted() {
		t.Errorf("EBIKE-001 must be accepted, got %v", results[0].Err)
	}
	if results[1].Accepted() {
		t.Error("SCOOT-002 must be rejected")
	}
	if !errors.Is(results[1].Err, inspection.ErrFailed) {
		t.Errorf("SCOOT-002 error = %v, want %v", results[1].Err, inspection.ErrFailed)
	}
	if !results[2].Accepted() {
		t.Errorf("BIKE-002 must be accepted, got %v", results[2].Err)
	}

	items := service.Items()
	if len(items) != 6 {
		t.Fatalf("inventory size = %d, want 6", len(items))
	}

	transports, equipment := 0, 0
	for _, item := range items {
		if _, ok := item.(domain.ITransport); ok {
			transports++
		} else {
			equipment++
		}
	}
	if slices.ContainsFunc(items, func(item domain.IInventoriable) bool {
		return item.GetInventoryNumber() == "SCOOT-002"
	}) {
		t.Error("rejected transport SCOOT-002 must not be in the inventory")
	}
	if transports != 4 || equipment != 2 {
		t.Errorf("transport/equipment = %d/%d, want 4/2", transports, equipment)
	}
}

func TestSyncAddsEquipmentToInventory(t *testing.T) {
	service := fleet.NewService(fakeSource{dataset: source.Dataset{
		Equipment: []source.EquipmentSpec{
			{Kind: domain.KindHelmet, InventoryNumber: "HELMET-001"},
			{Kind: domain.KindKneePads, InventoryNumber: "KNEE-001"},
			{Kind: domain.KindChargingStation, InventoryNumber: "STATION-001"},
		},
	}}, newInspector(t))

	if _, err := service.Sync(); err != nil {
		t.Fatalf("Sync() error = %v", err)
	}

	items := service.Items()
	if len(items) != 3 {
		t.Fatalf("inventory size = %d, want 3", len(items))
	}
	wantCategories := map[string]domain.ItemCategory{
		"HELMET-001":  domain.CategoryHelmet,
		"KNEE-001":    domain.CategoryKneePads,
		"STATION-001": domain.CategoryChargingStation,
	}
	for _, item := range items {
		want, ok := wantCategories[item.GetInventoryNumber()]
		if !ok {
			t.Errorf("unexpected item %q", item.GetInventoryNumber())
			continue
		}
		if item.GetCategory() != want {
			t.Errorf("category of %q = %q, want %q", item.GetInventoryNumber(), item.GetCategory(), want)
		}
		if _, ok := item.(domain.ITransport); ok {
			t.Errorf("equipment %q must not be transport", item.GetInventoryNumber())
		}
	}
}

func TestSyncHandlesEmptyDataset(t *testing.T) {
	service := fleet.NewService(fakeSource{}, newInspector(t))

	results, err := service.Sync()
	if err != nil {
		t.Fatalf("Sync() error = %v", err)
	}
	if len(results) != 0 {
		t.Errorf("registration results = %d, want 0", len(results))
	}
	if len(service.Items()) != 0 {
		t.Errorf("inventory is not empty: %+v", service.Items())
	}
}

func TestSyncReturnsSourceError(t *testing.T) {
	sourceErr := errors.New("source unavailable")
	service := fleet.NewService(fakeSource{err: sourceErr}, newInspector(t))

	if _, err := service.Sync(); !errors.Is(err, sourceErr) {
		t.Fatalf("Sync() error = %v, want %v", err, sourceErr)
	}
}

func TestSyncRejectsInvalidRegisteredDevice(t *testing.T) {
	service := fleet.NewService(fakeSource{dataset: source.Dataset{
		Registered: []source.DeviceSpec{
			{Kind: domain.KindBicycle, InventoryNumber: "BIKE-001", DailyConsumptionKWh: 1, EaseOfUse: 5},
		},
	}}, newInspector(t))

	if _, err := service.Sync(); !errors.Is(err, domain.ErrBicycleConsumption) {
		t.Fatalf("Sync() error = %v, want %v", err, domain.ErrBicycleConsumption)
	}
}
