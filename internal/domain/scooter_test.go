package domain_test

import (
	"errors"
	"testing"

	"github.com/Polsky75/hse-scoot/internal/domain"
)

func TestNewScooter(t *testing.T) {
	scooter, err := domain.NewScooter("  SCOOT-001  ", 1.25, 8)
	if err != nil {
		t.Fatalf("NewScooter() error = %v", err)
	}

	if got := scooter.GetInventoryNumber(); got != "SCOOT-001" {
		t.Errorf("GetInventoryNumber() = %q, want %q", got, "SCOOT-001")
	}
	if got := scooter.GetCategory(); got != domain.CategoryScooter {
		t.Errorf("GetCategory() = %q, want %q", got, domain.CategoryScooter)
	}
	if got := scooter.GetDailyConsumptionKWh(); got != 1.25 {
		t.Errorf("GetDailyConsumptionKWh() = %v, want 1.25", got)
	}
	if got := scooter.GetEaseOfUse(); got != 8 {
		t.Errorf("GetEaseOfUse() = %d, want 8", got)
	}
}

func TestNewScooterRejectsInvalidData(t *testing.T) {
	tests := []struct {
		name            string
		inventoryNumber string
		consumption     float64
		easeOfUse       int
		wantErr         error
	}{
		{
			name: "empty inventory number", inventoryNumber: "", consumption: 1,
			easeOfUse: 5, wantErr: domain.ErrEmptyInventoryNumber,
		},
		{
			name: "blank inventory number", inventoryNumber: "   ", consumption: 1,
			easeOfUse: 5, wantErr: domain.ErrEmptyInventoryNumber,
		},
		{
			name: "negative consumption", inventoryNumber: "SCOOT-001", consumption: -0.1,
			easeOfUse: 5, wantErr: domain.ErrNegativeConsumption,
		},
		{
			name: "ease below minimum", inventoryNumber: "SCOOT-001", consumption: 1,
			easeOfUse: 0, wantErr: domain.ErrInvalidEaseOfUse,
		},
		{
			name: "ease above maximum", inventoryNumber: "SCOOT-001", consumption: 1,
			easeOfUse: 11, wantErr: domain.ErrInvalidEaseOfUse,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := domain.NewScooter(test.inventoryNumber, test.consumption, test.easeOfUse)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("NewScooter() error = %v, want %v", err, test.wantErr)
			}
		})
	}
}
