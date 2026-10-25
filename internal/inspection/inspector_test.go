package inspection_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/Polsky75/hse-scoot/internal/domain"
	"github.com/Polsky75/hse-scoot/internal/inspection"
)

func newDevice(t *testing.T) *domain.Scooter {
	t.Helper()
	device, err := domain.NewScooter("SCOOT-001", 1.25, 5)
	if err != nil {
		t.Fatalf("NewScooter() error = %v", err)
	}
	return device
}

func TestInspectorAcceptsConditionAtThreshold(t *testing.T) {
	inspector, err := inspection.NewInspector(70)
	if err != nil {
		t.Fatalf("NewInspector() error = %v", err)
	}

	if err := inspector.Inspect(newDevice(t), 70); err != nil {
		t.Fatalf("Inspect() error = %v", err)
	}
}

func TestInspectorRejectsLowCondition(t *testing.T) {
	inspector, err := inspection.NewInspector(70)
	if err != nil {
		t.Fatalf("NewInspector() error = %v", err)
	}

	err = inspector.Inspect(newDevice(t), 69)
	if !errors.Is(err, inspection.ErrFailed) {
		t.Fatalf("Inspect() error = %v, want %v", err, inspection.ErrFailed)
	}
	if !strings.Contains(err.Error(), "69") || !strings.Contains(err.Error(), "70") {
		t.Errorf("Inspect() error = %v, want condition and threshold details", err)
	}
}

func TestInspectorRejectsInvalidCondition(t *testing.T) {
	inspector, err := inspection.NewInspector(70)
	if err != nil {
		t.Fatalf("NewInspector() error = %v", err)
	}

	for _, condition := range []int{-1, 101} {
		if err := inspector.Inspect(newDevice(t), condition); !errors.Is(err, inspection.ErrInvalidCondition) {
			t.Errorf("Inspect(%d) error = %v, want %v", condition, err, inspection.ErrInvalidCondition)
		}
	}
}

func TestNewInspectorRejectsInvalidThreshold(t *testing.T) {
	for _, threshold := range []int{-1, 101} {
		if _, err := inspection.NewInspector(threshold); !errors.Is(err, inspection.ErrInvalidCondition) {
			t.Errorf("NewInspector(%d) error = %v, want %v", threshold, err, inspection.ErrInvalidCondition)
		}
	}
}
