package fleet_test

import (
	"testing"

	"github.com/Polsky75/hse-scoot/internal/domain"
	"github.com/Polsky75/hse-scoot/internal/fleet"
	"github.com/Polsky75/hse-scoot/internal/source"
)

func newTransport(t *testing.T, kind domain.TransportKind, inventoryNumber string, consumption float64, ease int) domain.IInventoriable {
	t.Helper()
	device, err := domain.NewTransport(kind, inventoryNumber, consumption, ease)
	if err != nil {
		t.Fatalf("NewTransport(%s) error = %v", inventoryNumber, err)
	}
	return device
}

func newEquipment(t *testing.T, kind domain.EquipmentKind, inventoryNumber string) domain.IInventoriable {
	t.Helper()
	item, err := domain.NewEquipment(kind, inventoryNumber)
	if err != nil {
		t.Fatalf("NewEquipment(%s) error = %v", inventoryNumber, err)
	}
	return item
}

func TestBeginnerDevicesFiltersByEaseAndSumsConsumers(t *testing.T) {
	devices := []domain.IInventoriable{
		newTransport(t, domain.KindScooter, "SCOOT-001", 1.5, 8),
		newTransport(t, domain.KindBicycle, "BIKE-001", 0, 9),
		newTransport(t, domain.KindEBike, "EBIKE-001", 2.0, 7),
		newTransport(t, domain.KindScooter, "SCOOT-002", 3.0, 6),
		newTransport(t, domain.KindEBike, "EBIKE-002", 4.0, 1),
		newEquipment(t, domain.KindHelmet, "HELMET-001"),
		newEquipment(t, domain.KindKneePads, "KNEE-001"),
		newEquipment(t, domain.KindChargingStation, "STATION-001"),
	}

	selected, total := fleet.BeginnerDevices(devices)

	if len(selected) != 3 {
		t.Fatalf("selected count = %d, want 3", len(selected))
	}
	wantIDs := []string{"SCOOT-001", "BIKE-001", "EBIKE-001"}
	for i, want := range wantIDs {
		if got := selected[i].GetInventoryNumber(); got != want {
			t.Errorf("selected[%d] = %q, want %q", i, got, want)
		}
	}

	// BIKE-001 не IConsumptor, поэтому в сумму попадают только 1.5 и 2.0.
	// Снаряжение не ITransport и в выборку не попадает.
	if total != 3.5 {
		t.Errorf("total = %v, want 3.5", total)
	}
}

func TestBeginnerDevicesHandlesEmptyList(t *testing.T) {
	selected, total := fleet.BeginnerDevices(nil)

	if len(selected) != 0 {
		t.Errorf("selected = %v, want empty", selected)
	}
	if total != 0 {
		t.Errorf("total = %v, want 0", total)
	}
}

func TestServiceBeginnerDevicesUsesCurrentFleet(t *testing.T) {
	service := fleet.NewService(fakeSource{dataset: source.Dataset{
		Registered: []source.DeviceSpec{
			{Kind: domain.KindScooter, InventoryNumber: "SCOOT-001", DailyConsumptionKWh: 1.5, EaseOfUse: 8},
			{Kind: domain.KindBicycle, InventoryNumber: "BIKE-001", EaseOfUse: 9},
			{Kind: domain.KindEBike, InventoryNumber: "EBIKE-001", DailyConsumptionKWh: 2, EaseOfUse: 4},
		},
		Equipment: []source.EquipmentSpec{
			{Kind: domain.KindHelmet, InventoryNumber: "HELMET-001"},
			{Kind: domain.KindChargingStation, InventoryNumber: "STATION-001"},
		},
	}}, &fakeInspector{})

	if _, err := service.Sync(); err != nil {
		t.Fatalf("Sync() error = %v", err)
	}

	selected, total := service.BeginnerDevices()
	if len(selected) != 2 {
		t.Fatalf("selected count = %d, want 2", len(selected))
	}
	if got := selected[0].GetInventoryNumber(); got != "SCOOT-001" {
		t.Errorf("selected[0] = %q, want SCOOT-001", got)
	}
	if got := selected[1].GetInventoryNumber(); got != "BIKE-001" {
		t.Errorf("selected[1] = %q, want BIKE-001", got)
	}
	if total != 1.5 {
		t.Errorf("total = %v, want 1.5", total)
	}
}
