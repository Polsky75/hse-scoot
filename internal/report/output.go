package report

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

// Output — способ вывода отчёта. Новый способ (например, JSON в stdout
// или отправка по HTTP) добавляется реализацией этого интерфейса.
type Output interface {
	Write(Document) error
}

// ConsoleOutput печатает отчёт в заданный поток.
type ConsoleOutput struct {
	out io.Writer
}

// NewConsoleOutput создаёт вывод в консоль.
func NewConsoleOutput(out io.Writer) *ConsoleOutput {
	return &ConsoleOutput{out: out}
}

// Write выводит текстовый отчёт.
func (o *ConsoleOutput) Write(document Document) error {
	if o.out == nil {
		return errors.New("console output is not configured")
	}
	if _, err := io.WriteString(o.out, RenderText(document)); err != nil {
		return fmt.Errorf("write console report: %w", err)
	}
	return nil
}

// FileOutput сохраняет отчёт в файл.
type FileOutput struct {
	path   string
	notice io.Writer
}

// NewFileOutput создаёт вывод в файл. notice может быть nil.
func NewFileOutput(path string, notice io.Writer) (*FileOutput, error) {
	if strings.TrimSpace(path) == "" {
		return nil, errors.New("report file path must not be empty")
	}
	return &FileOutput{path: path, notice: notice}, nil
}

// Write сохраняет текстовый отчёт и сообщает путь, если задан notice.
func (o *FileOutput) Write(document Document) error {
	if err := os.WriteFile(o.path, []byte(RenderText(document)), 0o644); err != nil {
		return fmt.Errorf("write report file %q: %w", o.path, err)
	}
	if o.notice != nil {
		fmt.Fprintf(o.notice, "Отчёт записан: %s\n", o.path)
	}
	return nil
}

// NewOutput создаёт способ вывода по его имени.
//
// Сейчас поддерживаются console и file. Добавление нового метода
// сводится к новой реализации Output и ветке в этой фабрике.
func NewOutput(target, path string, console io.Writer) (Output, error) {
	switch strings.ToLower(strings.TrimSpace(target)) {
	case "console":
		return NewConsoleOutput(console), nil
	case "file":
		return NewFileOutput(path, console)
	default:
		return nil, fmt.Errorf("unknown report output %q (want console or file)", target)
	}
}
