package domain_test

import (
	"errors"
	"testing"

	"github.com/Polsky75/hse-scoot/internal/domain"
)

func TestValidateInventoryNumber(t *testing.T) {
	tests := []struct {
		name    string
		number  string
		wantErr error
	}{
		{name: "letters digits dashes", number: "SCOOT-001"},
		{name: "only digits", number: "12345"},
		{name: "only dashes", number: "---"},
		{name: "mixed", number: "AB-CD-12"},
		{name: "empty", number: "", wantErr: domain.ErrEmptyInventoryNumber},
		{name: "lowercase", number: "scoot-001", wantErr: domain.ErrInvalidInventoryNumber},
		{name: "underscore", number: "SCOOT_001", wantErr: domain.ErrInvalidInventoryNumber},
		{name: "space", number: "SCOOT 001", wantErr: domain.ErrInvalidInventoryNumber},
		{name: "dot", number: "SCOOT.001", wantErr: domain.ErrInvalidInventoryNumber},
		{name: "cyrillic", number: "САМОКАТ-001", wantErr: domain.ErrInvalidInventoryNumber},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := domain.ValidateInventoryNumber(test.number)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("ValidateInventoryNumber(%q) error = %v, want %v", test.number, err, test.wantErr)
			}
		})
	}
}
