package textfile_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Polsky75/hse-scoot/internal/domain"
	"github.com/Polsky75/hse-scoot/internal/source/textfile"
)

func writeDeviceFile(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "devices.txt")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	return path
}

func TestLoadParsesTransportAndEquipment(t *testing.T) {
	path := writeDeviceFile(t, `# hse-scoot inventory
[registered]
scooter SCOOT-001 1.15 95 8
bicycle BIKE-001 0 88 9

[incoming]
ebike EBIKE-001 1.80 42 7

[equipment]
helmet HELMET-001
kneepads KNEE-001
charging_station STATION-001
`)

	dataset, err := textfile.New(path).Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(dataset.Registered) != 2 {
		t.Fatalf("registered count = %d, want 2", len(dataset.Registered))
	}
	if got := dataset.Registered[0]; got.Kind != domain.KindScooter || got.InventoryNumber != "SCOOT-001" ||
		got.DailyConsumptionKWh != 1.15 || got.Condition != 95 || got.EaseOfUse != 8 {
		t.Errorf("registered[0] = %+v, want scooter SCOOT-001 1.15 95 8", got)
	}
	if got := dataset.Registered[1]; got.Kind != domain.KindBicycle || got.InventoryNumber != "BIKE-001" ||
		got.Condition != 88 || got.EaseOfUse != 9 {
		t.Errorf("registered[1] = %+v, want bicycle BIKE-001 88 9", got)
	}

	if len(dataset.Incoming) != 1 {
		t.Fatalf("incoming count = %d, want 1", len(dataset.Incoming))
	}
	if got := dataset.Incoming[0]; got.Kind != domain.KindEBike || got.InventoryNumber != "EBIKE-001" ||
		got.Condition != 42 || got.EaseOfUse != 7 {
		t.Errorf("incoming[0] = %+v, want ebike EBIKE-001 42 7", got)
	}

	if len(dataset.Equipment) != 3 {
		t.Fatalf("equipment count = %d, want 3", len(dataset.Equipment))
	}
	wantEquipment := []struct {
		kind            domain.EquipmentKind
		inventoryNumber string
	}{
		{kind: domain.KindHelmet, inventoryNumber: "HELMET-001"},
		{kind: domain.KindKneePads, inventoryNumber: "KNEE-001"},
		{kind: domain.KindChargingStation, inventoryNumber: "STATION-001"},
	}
	for i, want := range wantEquipment {
		if got := dataset.Equipment[i]; got.Kind != want.kind || got.InventoryNumber != want.inventoryNumber {
			t.Errorf("equipment[%d] = %+v, want %s %s", i, got, want.kind, want.inventoryNumber)
		}
	}
}

func TestLoadRejectsInvalidFiles(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr string
	}{
		{
			name:    "item outside section",
			content: "scooter SCOOT-001 1.15 95 8\n",
			wantErr: "outside a section",
		},
		{
			name:    "unknown section",
			content: "[unknown]\n",
			wantErr: "unknown section",
		},
		{
			name:    "malformed section",
			content: "[registered\n",
			wantErr: "malformed section header",
		},
		{
			name:    "wrong transport field count",
			content: "[incoming]\nscooter SCOOT-001 1.15 95\n",
			wantErr: "expected 5 fields",
		},
		{
			name:    "wrong equipment field count",
			content: "[equipment]\nhelmet\n",
			wantErr: "expected 2 fields",
		},
		{
			name:    "unknown transport kind",
			content: "[incoming]\ntruck TRUCK-001 1.15 95 5\n",
			wantErr: "unknown transport kind",
		},
		{
			name:    "unknown equipment kind",
			content: "[equipment]\njacket JACKET-001\n",
			wantErr: "unknown equipment kind",
		},
		{
			name:    "bad consumption",
			content: "[incoming]\nscooter SCOOT-001 nope 95 5\n",
			wantErr: "invalid daily consumption",
		},
		{
			name:    "negative consumption",
			content: "[incoming]\nscooter SCOOT-001 -1 95 5\n",
			wantErr: "non-negative finite number",
		},
		{
			name:    "non-finite consumption",
			content: "[incoming]\nscooter SCOOT-001 NaN 95 5\n",
			wantErr: "non-negative finite number",
		},
		{
			name:    "condition out of range",
			content: "[incoming]\nscooter SCOOT-001 1 101 5\n",
			wantErr: "between 0 and 100",
		},
		{
			name:    "bad ease of use",
			content: "[incoming]\nscooter SCOOT-001 1 95 nope\n",
			wantErr: "invalid ease of use",
		},
		{
			name:    "ease of use below minimum",
			content: "[incoming]\nscooter SCOOT-001 1 95 0\n",
			wantErr: "between 1 and 10",
		},
		{
			name:    "ease of use above maximum",
			content: "[incoming]\nscooter SCOOT-001 1 95 11\n",
			wantErr: "between 1 and 10",
		},
		{
			name:    "duplicate transport inventory number",
			content: "[registered]\nscooter SCOOT-001 1 95 8\n[incoming]\nscooter SCOOT-001 2 95 8\n",
			wantErr: "duplicate inventory number",
		},
		{
			name:    "duplicate inventory number across transport and equipment",
			content: "[registered]\nscooter SHARED-001 1 95 8\n[equipment]\nhelmet SHARED-001\n",
			wantErr: "duplicate inventory number",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := textfile.New(writeDeviceFile(t, test.content)).Load()
			if err == nil {
				t.Fatal("Load() error = nil, want error")
			}
			if !strings.Contains(err.Error(), test.wantErr) {
				t.Errorf("Load() error = %v, want substring %q", err, test.wantErr)
			}
		})
	}
}

func TestLoadReturnsErrorForMissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.txt")
	_, err := textfile.New(path).Load()
	if err == nil {
		t.Fatal("Load() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "open inventory file") {
		t.Errorf("Load() error = %v, want open error", err)
	}
}
