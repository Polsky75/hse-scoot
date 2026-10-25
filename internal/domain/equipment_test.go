package domain_test

import (
	"errors"
	"testing"

	"github.com/Polsky75/hse-scoot/internal/domain"
)

func TestParseEquipmentKind(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		want    domain.EquipmentKind
		wantErr error
	}{
		{name: "helmet", value: "helmet", want: domain.KindHelmet},
		{name: "kneepads", value: "kneepads", want: domain.KindKneePads},
		{name: "charging station", value: "charging_station", want: domain.KindChargingStation},
		{name: "case and spaces", value: "  HELMET  ", want: domain.KindHelmet},
		{name: "unknown", value: "jacket", wantErr: domain.ErrUnknownEquipmentKind},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := domain.ParseEquipmentKind(test.value)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("ParseEquipmentKind() error = %v, want %v", err, test.wantErr)
			}
			if got != test.want {
				t.Errorf("ParseEquipmentKind() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestNewEquipmentCreatesSupportedKinds(t *testing.T) {
	tests := []struct {
		kind            domain.EquipmentKind
		inventoryNumber string
		consumption     float64
		wantCategory    domain.ItemCategory
		wantConsumptor  bool
	}{
		{
			kind: domain.KindHelmet, inventoryNumber: "HELMET-001",
			wantCategory: domain.CategoryHelmet,
		},
		{
			kind: domain.KindKneePads, inventoryNumber: "KNEE-001",
			wantCategory: domain.CategoryKneePads,
		},
		{
			kind: domain.KindChargingStation, inventoryNumber: "STATION-001",
			consumption: 3.2, wantCategory: domain.CategoryChargingStation,
			wantConsumptor: true,
		},
	}

	for _, test := range tests {
		t.Run(string(test.kind), func(t *testing.T) {
			item, err := domain.NewEquipment(test.kind, test.inventoryNumber, test.consumption)
			if err != nil {
				t.Fatalf("NewEquipment() error = %v", err)
			}
			if got := item.GetInventoryNumber(); got != test.inventoryNumber {
				t.Errorf("GetInventoryNumber() = %q, want %q", got, test.inventoryNumber)
			}
			if got := item.GetCategory(); got != test.wantCategory {
				t.Errorf("GetCategory() = %q, want %q", got, test.wantCategory)
			}

			var gotKind string
			switch item.(type) {
			case *domain.Helmet:
				gotKind = string(domain.KindHelmet)
			case *domain.KneePads:
				gotKind = string(domain.KindKneePads)
			case *domain.ChargingStation:
				gotKind = string(domain.KindChargingStation)
			}
			if gotKind != string(test.kind) {
				t.Errorf("created type = %q, want %q", gotKind, test.kind)
			}

			if _, ok := item.(domain.ITransport); ok {
				t.Error("equipment must not implement ITransport")
			}

			consumer, ok := item.(domain.IConsumptor)
			if ok != test.wantConsumptor {
				t.Fatalf("equipment implements IConsumptor = %v, want %v", ok, test.wantConsumptor)
			}
			if ok && consumer.GetDailyConsumptionKWh() != test.consumption {
				t.Errorf(
					"GetDailyConsumptionKWh() = %v, want %v",
					consumer.GetDailyConsumptionKWh(), test.consumption,
				)
			}
		})
	}
}

func TestNewEquipmentRejectsInvalidData(t *testing.T) {
	tests := []struct {
		name            string
		kind            domain.EquipmentKind
		inventoryNumber string
		consumption     float64
		wantErr         error
	}{
		{
			name: "empty inventory number", kind: domain.KindHelmet,
			inventoryNumber: "  ", wantErr: domain.ErrEmptyInventoryNumber,
		},
		{
			name: "unknown kind", kind: "jacket",
			inventoryNumber: "JACKET-001", wantErr: domain.ErrUnknownEquipmentKind,
		},
		{
			name: "helmet with consumption", kind: domain.KindHelmet,
			inventoryNumber: "HELMET-001", consumption: 1,
			wantErr: domain.ErrEquipmentConsumption,
		},
		{
			name: "kneepads with consumption", kind: domain.KindKneePads,
			inventoryNumber: "KNEE-001", consumption: 1,
			wantErr: domain.ErrEquipmentConsumption,
		},
		{
			name: "negative station consumption", kind: domain.KindChargingStation,
			inventoryNumber: "STATION-001", consumption: -0.1,
			wantErr: domain.ErrNegativeConsumption,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := domain.NewEquipment(test.kind, test.inventoryNumber, test.consumption)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("NewEquipment() error = %v, want %v", err, test.wantErr)
			}
		})
	}
}
