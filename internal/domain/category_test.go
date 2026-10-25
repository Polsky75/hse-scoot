package domain_test

import (
	"testing"

	"github.com/Polsky75/hse-scoot/internal/domain"
)

func TestItemCategoryLabel(t *testing.T) {
	tests := map[domain.ItemCategory]string{
		domain.CategoryScooter:         "самокат",
		domain.CategoryBicycle:         "велосипед",
		domain.CategoryEBike:           "электровелосипед",
		domain.CategoryHelmet:          "шлем",
		domain.CategoryKneePads:        "наколенники",
		domain.CategoryChargingStation: "зарядная станция",
	}

	for category, want := range tests {
		if got := category.Label(); got != want {
			t.Errorf("%q.Label() = %q, want %q", category, got, want)
		}
		if got := category.String(); got != string(category) {
			t.Errorf("%q.String() = %q, want %q", category, got, string(category))
		}
	}
}
