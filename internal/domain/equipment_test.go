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
		wantCategory    domain.ItemCategory
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
			wantCategory: domain.CategoryChargingStation,
		},
	}

	for _, test := range tests {
		t.Run(string(test.kind), func(t *testing.T) {
			item, err := domain.NewEquipment(test.kind, test.inventoryNumber)
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
			if _, ok := item.(domain.IConsumptor); ok {
				t.Error("equipment must not implement IConsumptor")
			}
		})
	}
}

func TestNewEquipmentRejectsInvalidData(t *testing.T) {
	if _, err := domain.NewEquipment(domain.KindHelmet, "  "); !errors.Is(err, domain.ErrEmptyInventoryNumber) {
		t.Errorf("empty inventory number error = %v, want %v", err, domain.ErrEmptyInventoryNumber)
	}
	if _, err := domain.NewEquipment("jacket", "JACKET-001"); !errors.Is(err, domain.ErrUnknownEquipmentKind) {
		t.Errorf("unknown kind error = %v, want %v", err, domain.ErrUnknownEquipmentKind)
	}
}
