package fleet_test

import (
	"errors"
	"slices"
	"testing"

	"github.com/Polsky75/hse-scoot/internal/domain"
	"github.com/Polsky75/hse-scoot/internal/fleet"
	"github.com/Polsky75/hse-scoot/internal/source"
)

type fakeSource struct {
	dataset source.Dataset
	err     error
}

func (s fakeSource) Load() (source.Dataset, error) {
	return s.dataset, s.err
}

var errInspection = errors.New("inspection failed")

type inspectCall struct {
	inventoryNumber string
	condition       int
}

// fakeInspector — управляемый тестовый дубль техосмотра.
type fakeInspector struct {
	inspect func(domain.IInventoriable, int) error
	calls   []inspectCall
}

func (f *fakeInspector) Inspect(device domain.IInventoriable, condition int) error {
	f.calls = append(f.calls, inspectCall{
		inventoryNumber: device.GetInventoryNumber(),
		condition:       condition,
	})
	if f.inspect != nil {
		return f.inspect(device, condition)
	}
	return nil
}

func TestSyncRestoresInventoryAndRegistersInspectedTransport(t *testing.T) {
	inspector := &fakeInspector{
		inspect: func(device domain.IInventoriable, _ int) error {
			if device.GetInventoryNumber() == "SCOOT-002" {
				return errInspection
			}
			return nil
		},
	}
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
			{Kind: domain.KindChargingStation, InventoryNumber: "STATION-001", DailyConsumptionKWh: 3.2},
		},
	}}, inspector)

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
	if !errors.Is(results[1].Err, errInspection) {
		t.Errorf("SCOOT-002 error = %v, want %v", results[1].Err, errInspection)
	}
	if !results[2].Accepted() {
		t.Errorf("BIKE-002 must be accepted, got %v", results[2].Err)
	}

	wantCalls := []inspectCall{
		{inventoryNumber: "EBIKE-001", condition: 70},
		{inventoryNumber: "SCOOT-002", condition: 69},
		{inventoryNumber: "BIKE-002", condition: 80},
	}
	if !slices.Equal(inspector.calls, wantCalls) {
		t.Errorf("inspection calls = %+v, want %+v", inspector.calls, wantCalls)
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

func TestSyncUsesFakeInspectorOnlyForIncomingTransport(t *testing.T) {
	inspector := &fakeInspector{}
	service := fleet.NewService(fakeSource{dataset: source.Dataset{
		Registered: []source.DeviceSpec{
			{Kind: domain.KindScooter, InventoryNumber: "SCOOT-001", DailyConsumptionKWh: 1, EaseOfUse: 8},
		},
		Equipment: []source.EquipmentSpec{
			{Kind: domain.KindHelmet, InventoryNumber: "HELMET-001"},
		},
		Incoming: []source.DeviceSpec{
			{Kind: domain.KindEBike, InventoryNumber: "EBIKE-001", DailyConsumptionKWh: 2, Condition: 80, EaseOfUse: 7},
		},
	}}, inspector)

	results, err := service.Sync()
	if err != nil {
		t.Fatalf("Sync() error = %v", err)
	}
	if len(results) != 1 || !results[0].Accepted() {
		t.Fatalf("results = %+v, want one accepted", results)
	}

	wantCalls := []inspectCall{{inventoryNumber: "EBIKE-001", condition: 80}}
	if !slices.Equal(inspector.calls, wantCalls) {
		t.Errorf("inspection calls = %+v, want only incoming EBIKE-001", inspector.calls)
	}
}

func TestSyncAddsEquipmentToInventory(t *testing.T) {
	inspector := &fakeInspector{}
	service := fleet.NewService(fakeSource{dataset: source.Dataset{
		Equipment: []source.EquipmentSpec{
			{Kind: domain.KindHelmet, InventoryNumber: "HELMET-001"},
			{Kind: domain.KindKneePads, InventoryNumber: "KNEE-001"},
			{Kind: domain.KindChargingStation, InventoryNumber: "STATION-001", DailyConsumptionKWh: 3.2},
		},
	}}, inspector)

	if _, err := service.Sync(); err != nil {
		t.Fatalf("Sync() error = %v", err)
	}
	if len(inspector.calls) != 0 {
		t.Errorf("equipment must not be inspected, calls = %+v", inspector.calls)
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
	inspector := &fakeInspector{}
	service := fleet.NewService(fakeSource{}, inspector)

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
	if len(inspector.calls) != 0 {
		t.Errorf("inspection calls = %+v, want none", inspector.calls)
	}
}

func TestSyncReturnsSourceError(t *testing.T) {
	sourceErr := errors.New("source unavailable")
	inspector := &fakeInspector{}
	service := fleet.NewService(fakeSource{err: sourceErr}, inspector)

	if _, err := service.Sync(); !errors.Is(err, sourceErr) {
		t.Fatalf("Sync() error = %v, want %v", err, sourceErr)
	}
	if len(inspector.calls) != 0 {
		t.Errorf("inspection calls = %+v, want none", inspector.calls)
	}
}

func TestSyncRejectsInvalidRegisteredDevice(t *testing.T) {
	inspector := &fakeInspector{}
	service := fleet.NewService(fakeSource{dataset: source.Dataset{
		Registered: []source.DeviceSpec{
			{Kind: domain.KindBicycle, InventoryNumber: "BIKE-001", DailyConsumptionKWh: 1, EaseOfUse: 5},
		},
	}}, inspector)

	if _, err := service.Sync(); !errors.Is(err, domain.ErrBicycleConsumption) {
		t.Fatalf("Sync() error = %v, want %v", err, domain.ErrBicycleConsumption)
	}
	if len(inspector.calls) != 0 {
		t.Errorf("inspection calls = %+v, want none", inspector.calls)
	}
}

func TestSyncRejectsInvalidInventoryNumber(t *testing.T) {
	inspector := &fakeInspector{}
	service := fleet.NewService(fakeSource{dataset: source.Dataset{
		Incoming: []source.DeviceSpec{
			{Kind: domain.KindScooter, InventoryNumber: "scoot-001", DailyConsumptionKWh: 1, Condition: 90, EaseOfUse: 5},
		},
	}}, inspector)

	results, err := service.Sync()
	if err != nil {
		t.Fatalf("Sync() error = %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("results = %d, want 1", len(results))
	}
	if results[0].Accepted() {
		t.Fatal("invalid inventory number must be rejected")
	}
	if !errors.Is(results[0].Err, domain.ErrInvalidInventoryNumber) {
		t.Errorf("error = %v, want %v", results[0].Err, domain.ErrInvalidInventoryNumber)
	}
	if len(inspector.calls) != 0 {
		t.Errorf("inspection calls = %+v, want none", inspector.calls)
	}
}
