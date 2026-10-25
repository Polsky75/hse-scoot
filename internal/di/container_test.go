package di_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Polsky75/hse-scoot/internal/di"
	"github.com/Polsky75/hse-scoot/internal/report"
	"github.com/Polsky75/hse-scoot/internal/source"
)

type fakeSource struct {
	dataset source.Dataset
	err     error
}

func (s fakeSource) Load() (source.Dataset, error) {
	return s.dataset, s.err
}

func TestNewReadsTextFileAndRunsRegistration(t *testing.T) {
	path := filepath.Join(t.TempDir(), "devices.txt")
	content := `# test inventory
[registered]
scooter SCOOT-001 1.15 95 8
[incoming]
ebike EBIKE-001 2.10 80 7
scooter SCOOT-002 1.00 40 5
bicycle BIKE-001 0.00 90 9
[equipment]
helmet HELMET-001
kneepads KNEE-001
charging_station STATION-001
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	var out bytes.Buffer
	container, err := di.New(&out, path, report.NewConsoleOutput(&out))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if err := container.Run(); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	text := out.String()
	for _, want := range []string{
		"принят",
		"SCOOT-002",
		"техосмотр не пройден",
		"hse-scoot — отчёт по инвентарю",
		"Всего вещей на учёте: 6",
		"Транспорт: 3",
		"Самокаты: 1",
		"Наколенники: 1",
		"Суммарное суточное потребление: 3.25 кВт·ч",
		"Всё оборудование (3):",
		"Транспорт (3):",
		"Транспорт для новичков (3):",
		"SCOOT-001 — самокат",
		"HELMET-001 — шлем",
		"STATION-001 — зарядная станция",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("output does not contain %q:\n%s", want, text)
		}
	}
}

func TestNewWithSourceRuns(t *testing.T) {
	var out bytes.Buffer
	container, err := di.NewWithSource(&out, fakeSource{}, 70, nil)
	if err != nil {
		t.Fatalf("NewWithSource() error = %v", err)
	}
	if err := container.Run(); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	text := out.String()
	for _, want := range []string{"нет новых устройств", "Всего вещей на учёте: 0"} {
		if !strings.Contains(text, want) {
			t.Errorf("output does not contain %q:\n%s", want, text)
		}
	}
}

func TestNewWithFileOutput(t *testing.T) {
	reportPath := filepath.Join(t.TempDir(), "report.txt")
	var notice bytes.Buffer
	output, err := report.NewFileOutput(reportPath, &notice)
	if err != nil {
		t.Fatalf("NewFileOutput() error = %v", err)
	}

	var out bytes.Buffer
	container, err := di.NewWithSource(&out, fakeSource{}, 70, output)
	if err != nil {
		t.Fatalf("NewWithSource() error = %v", err)
	}
	if err := container.Run(); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	content, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if !strings.Contains(string(content), "Всего вещей на учёте: 0") {
		t.Errorf("file does not contain report:\n%s", content)
	}
	if !strings.Contains(notice.String(), reportPath) {
		t.Errorf("notice = %q, want path %q", notice.String(), reportPath)
	}
}

func TestNewWithSourceRejectsInvalidThreshold(t *testing.T) {
	if _, err := di.NewWithSource(&bytes.Buffer{}, fakeSource{}, 101, nil); err == nil {
		t.Fatal("NewWithSource() error = nil, want invalid threshold error")
	}
}
