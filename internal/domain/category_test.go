package domain_test

import (
	"testing"

	"github.com/Polsky75/hse-scoot/internal/domain"
)

func TestItemCategoryLabel(t *testing.T) {
	tests := []struct {
		category domain.ItemCategory
		label    string
		plural   string
	}{
		{domain.CategoryScooter, "самокат", "Самокаты"},
		{domain.CategoryBicycle, "велосипед", "Велосипеды"},
		{domain.CategoryEBike, "электровелосипед", "Электровелосипеды"},
		{domain.CategoryHelmet, "шлем", "Шлемы"},
		{domain.CategoryKneePads, "наколенники", "Наколенники"},
		{domain.CategoryChargingStation, "зарядная станция", "Зарядные станции"},
	}

	for _, test := range tests {
		if got := test.category.Label(); got != test.label {
			t.Errorf("%q.Label() = %q, want %q", test.category, got, test.label)
		}
		if got := test.category.LabelPlural(); got != test.plural {
			t.Errorf("%q.LabelPlural() = %q, want %q", test.category, got, test.plural)
		}
		if got := test.category.String(); got != string(test.category) {
			t.Errorf("%q.String() = %q, want %q", test.category, got, string(test.category))
		}
	}
}
