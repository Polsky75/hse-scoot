package report_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Polsky75/hse-scoot/internal/report"
)

func testDocument(t *testing.T) report.Document {
	t.Helper()
	items, beginner := testInventory(t)
	return report.Build(items, beginner)
}

func TestConsoleOutputWritesText(t *testing.T) {
	var out bytes.Buffer
	if err := report.NewConsoleOutput(&out).Write(testDocument(t)); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if !strings.Contains(out.String(), "Всего вещей на учёте: 7") {
		t.Errorf("console output does not contain counts:\n%s", out.String())
	}
}

func TestFileOutputWritesFileAndNotice(t *testing.T) {
	path := filepath.Join(t.TempDir(), "report.txt")
	var notice bytes.Buffer

	output, err := report.NewFileOutput(path, &notice)
	if err != nil {
		t.Fatalf("NewFileOutput() error = %v", err)
	}
	if err := output.Write(testDocument(t)); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if !strings.Contains(string(content), "Транспорт для новичков (3):") {
		t.Errorf("file does not contain report:\n%s", content)
	}
	if !strings.Contains(notice.String(), path) {
		t.Errorf("notice = %q, want path %q", notice.String(), path)
	}
}

func TestNewOutput(t *testing.T) {
	console, err := report.NewOutput("console", "", &bytes.Buffer{})
	if err != nil || console == nil {
		t.Fatalf("NewOutput(console) = %v, %v", console, err)
	}

	file, err := report.NewOutput("file", filepath.Join(t.TempDir(), "report.txt"), &bytes.Buffer{})
	if err != nil || file == nil {
		t.Fatalf("NewOutput(file) = %v, %v", file, err)
	}

	if _, err := report.NewOutput("json", "", &bytes.Buffer{}); err == nil {
		t.Fatal("NewOutput(json) error = nil, want error")
	}
}

func TestNewFileOutputRejectsEmptyPath(t *testing.T) {
	if _, err := report.NewFileOutput("  ", nil); err == nil {
		t.Fatal("NewFileOutput() error = nil, want error")
	}
}
