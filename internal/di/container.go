// Package di собирает граф зависимостей приложения.
//
// Контейнер намеренно ручной и явный: для текущего размера приложения
// библиотека внедрения зависимостей была бы лишней абстракцией.
package di

import (
	"fmt"
	"io"

	"github.com/Polsky75/hse-scoot/internal/console"
	"github.com/Polsky75/hse-scoot/internal/fleet"
	"github.com/Polsky75/hse-scoot/internal/inspection"
	"github.com/Polsky75/hse-scoot/internal/interactive"
	"github.com/Polsky75/hse-scoot/internal/report"
	"github.com/Polsky75/hse-scoot/internal/source"
	"github.com/Polsky75/hse-scoot/internal/source/textfile"
)

// Container хранит собранный граф зависимостей.
type Container struct {
	file        *console.App
	interactive *interactive.App
}

// New собирает приложение с текстовым источником инвентаря.
// Если output равен nil, отчёт выводится в out.
func New(out io.Writer, devicesPath string, output report.Output) (*Container, error) {
	return NewWithSource(out, textfile.New(devicesPath), inspection.DefaultMinCondition, output)
}

// NewWithSource собирает приложение с произвольным источником данных
// и выбранным способом вывода отчёта.
// Точка расширения: JSON, БД или другой источник подключается здесь,
// без изменений в домене, сервисе и консоли.
func NewWithSource(
	out io.Writer,
	src source.Source,
	minCondition int,
	output report.Output,
) (*Container, error) {
	inspector, err := inspection.NewInspector(minCondition)
	if err != nil {
		return nil, fmt.Errorf("create inspector: %w", err)
	}
	if output == nil {
		output = report.NewConsoleOutput(out)
	}

	service := fleet.NewService(src, inspector)
	return &Container{
		file:        console.New(service, output, out),
		interactive: interactive.New(service, output, out),
	}, nil
}

// Run запускает файловый сценарий.
func (c *Container) Run() error {
	return c.file.Run()
}

// RunInteractive запускает интерактивный ввод.
func (c *Container) RunInteractive(in io.Reader) error {
	return c.interactive.Run(in)
}
