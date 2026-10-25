package domain_test

import (
	"errors"
	"testing"

	"github.com/Polsky75/hse-scoot/internal/domain"
)

func TestParseTransportKind(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		want    domain.TransportKind
		wantErr error
	}{
		{name: "scooter", value: "scooter", want: domain.KindScooter},
		{name: "bicycle", value: "bicycle", want: domain.KindBicycle},
		{name: "ebike", value: "ebike", want: domain.KindEBike},
		{name: "case and spaces", value: "  EBike  ", want: domain.KindEBike},
		{name: "unknown", value: "truck", wantErr: domain.ErrUnknownTransportKind},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := domain.ParseTransportKind(test.value)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("ParseTransportKind() error = %v, want %v", err, test.wantErr)
			}
			if got != test.want {
				t.Errorf("ParseTransportKind() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestNewTransportCreatesSupportedKinds(t *testing.T) {
	tests := []struct {
		name            string
		kind            domain.TransportKind
		inventoryNumber string
		consumptionKWh  float64
		easeOfUse       int
		wantCategory    domain.ItemCategory
		wantConsumptor  bool
	}{
		{
			name: "scooter", kind: domain.KindScooter, inventoryNumber: "SCOOT-001",
			consumptionKWh: 1.25, easeOfUse: 8, wantCategory: domain.CategoryScooter,
			wantConsumptor: true,
		},
		{
			name: "bicycle", kind: domain.KindBicycle, inventoryNumber: "BIKE-001",
			easeOfUse: 9, wantCategory: domain.CategoryBicycle, wantConsumptor: false,
		},
		{
			name: "ebike", kind: domain.KindEBike, inventoryNumber: "EBIKE-001",
			consumptionKWh: 2.10, easeOfUse: 7, wantCategory: domain.CategoryEBike,
			wantConsumptor: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			device, err := domain.NewTransport(
				test.kind,
				test.inventoryNumber,
				test.consumptionKWh,
				test.easeOfUse,
			)
			if err != nil {
				t.Fatalf("NewTransport() error = %v", err)
			}
			if got := device.GetInventoryNumber(); got != test.inventoryNumber {
				t.Errorf("GetInventoryNumber() = %q, want %q", got, test.inventoryNumber)
			}
			if got := device.GetCategory(); got != test.wantCategory {
				t.Errorf("GetCategory() = %q, want %q", got, test.wantCategory)
			}
			if got := device.GetEaseOfUse(); got != test.easeOfUse {
				t.Errorf("GetEaseOfUse() = %d, want %d", got, test.easeOfUse)
			}

			consumer, ok := device.(domain.IConsumptor)
			if ok != test.wantConsumptor {
				t.Fatalf("device implements IConsumptor = %v, want %v", ok, test.wantConsumptor)
			}
			if ok && consumer.GetDailyConsumptionKWh() != test.consumptionKWh {
				t.Errorf(
					"GetDailyConsumptionKWh() = %v, want %v",
					consumer.GetDailyConsumptionKWh(), test.consumptionKWh,
				)
			}
		})
	}
}

func TestNewTransportRejectsBicycleWithConsumption(t *testing.T) {
	_, err := domain.NewTransport(domain.KindBicycle, "BIKE-001", 1, 5)
	if !errors.Is(err, domain.ErrBicycleConsumption) {
		t.Fatalf("NewTransport() error = %v, want %v", err, domain.ErrBicycleConsumption)
	}
}

func TestNewTransportRejectsInvalidEaseOfUse(t *testing.T) {
	_, err := domain.NewTransport(domain.KindScooter, "SCOOT-001", 1, 11)
	if !errors.Is(err, domain.ErrInvalidEaseOfUse) {
		t.Fatalf("NewTransport() error = %v, want %v", err, domain.ErrInvalidEaseOfUse)
	}
}

func TestBicycleIsNotConsumptor(t *testing.T) {
	bicycle, err := domain.NewBicycle("BIKE-001", 9)
	if err != nil {
		t.Fatalf("NewBicycle() error = %v", err)
	}
	if _, ok := any(bicycle).(domain.IConsumptor); ok {
		t.Fatal("Bicycle must not implement IConsumptor")
	}
	if _, ok := any(bicycle).(domain.ITransport); !ok {
		t.Fatal("Bicycle must implement ITransport")
	}
	if _, ok := any(bicycle).(domain.IInventoriable); !ok {
		t.Fatal("Bicycle must implement IInventoriable")
	}
	if got := bicycle.GetCategory(); got != domain.CategoryBicycle {
		t.Errorf("GetCategory() = %q, want %q", got, domain.CategoryBicycle)
	}
}
