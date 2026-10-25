package interactive_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/Polsky75/hse-scoot/internal/domain"
	"github.com/Polsky75/hse-scoot/internal/fleet"
	"github.com/Polsky75/hse-scoot/internal/interactive"
	"github.com/Polsky75/hse-scoot/internal/report"
)

type fakeInspector struct {
	calls []inspectCall
}

type inspectCall struct {
	inventoryNumber string
	condition       int
}

func (f *fakeInspector) Inspect(device domain.IInventoriable, condition int) error {
	f.calls = append(f.calls, inspectCall{
		inventoryNumber: device.GetInventoryNumber(),
		condition:       condition,
	})
	if condition < 70 {
		return errors.New("техосмотр не пройден")
	}
	return nil
}

func newApp(t *testing.T) (*interactive.App, *bytes.Buffer, *fakeInspector) {
	t.Helper()
	inspector := &fakeInspector{}
	var out bytes.Buffer
	app := interactive.New(
		fleet.NewService(nil, inspector),
		report.NewConsoleOutput(&out),
		&out,
	)
	return app, &out, inspector
}

func TestInteractiveAddsRegisteredWithRetry(t *testing.T) {
	app, out, _ := newApp(t)

	// 9 — неверное действие, 0 — неверная группа, 0 — неверный вид,
	// abc — неверный расход.
	input := "9\n1\n0\n1\n0\n1\n\nabc\n1.5\n8\n0\n"
	if err := app.Run(strings.NewReader(input)); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	text := out.String()
	for _, want := range []string{
		"Ошибка: введите целое число от 0 до 3.",
		"Ошибка: введите целое число от 1 до 2.",
		"Ошибка: введите целое число от 1 до 3.",
		"Ошибка: введите неотрицательное число.",
		"Добавлено на учёт. Номер: SCOOT-001 — самокат",
		"Всего вещей на учёте: 1",
		"Суммарное суточное потребление: 1.50 кВт·ч",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("output does not contain %q:\n%s", want, text)
		}
	}
}

func TestInteractiveRejectsInvalidInventoryNumber(t *testing.T) {
	app, out, _ := newApp(t)

	input := "1\n1\n1\nscoot-01\nSCOOT-001\n1.5\n8\n0\n"
	if err := app.Run(strings.NewReader(input)); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	text := out.String()
	for _, want := range []string{
		"Ошибка: номер может состоять только из заглавных латинских букв, цифр и тире.",
		"Добавлено на учёт. Номер: SCOOT-001 — самокат",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("output does not contain %q:\n%s", want, text)
		}
	}
}

func TestInteractiveAcceptsAndRejectsTransportApplications(t *testing.T) {
	app, out, inspector := newApp(t)

	// Заявка на электровелосипед принята, заявка на самокат отклонена.
	// 101 — неверная оценка состояния, затем 40.
	input := "2\n3\n\n2\n7\n80\n2\n1\n\n1\n5\n101\n40\n0\n"
	if err := app.Run(strings.NewReader(input)); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	wantCalls := []inspectCall{
		{inventoryNumber: "EBIKE-001", condition: 80},
		{inventoryNumber: "SCOOT-001", condition: 40},
	}
	if len(inspector.calls) != len(wantCalls) {
		t.Fatalf("inspection calls = %+v, want %+v", inspector.calls, wantCalls)
	}
	for i, want := range wantCalls {
		if inspector.calls[i] != want {
			t.Errorf("call[%d] = %+v, want %+v", i, inspector.calls[i], want)
		}
	}

	text := out.String()
	for _, want := range []string{
		"Заявка принята. Присвоен номер: EBIKE-001 — электровелосипед",
		"Ошибка: введите целое число от 0 до 100.",
		"Заявка отклонена: техосмотр не пройден",
		"Всего вещей на учёте: 1",
		"Суммарное суточное потребление: 2.00 кВт·ч",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("output does not contain %q:\n%s", want, text)
		}
	}
}

func TestInteractiveEquipmentAndDuplicateNumber(t *testing.T) {
	app, out, _ := newApp(t)

	input := "1\n2\n3\n\n3.2\n1\n1\n1\nDUP\n1\n5\n1\n1\n1\nDUP\n\n1\n5\n0\n"
	if err := app.Run(strings.NewReader(input)); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	text := out.String()
	for _, want := range []string{
		"Добавлено на учёт. Номер: STATION-001 — зарядная станция",
		"Добавлено на учёт. Номер: DUP — самокат",
		"Ошибка: номер \"DUP\" уже используется. Повторите ввод.",
		"Добавлено на учёт. Номер: SCOOT-001 — самокат",
		"Всего вещей на учёте: 3",
		"Суммарное суточное потребление: 5.20 кВт·ч",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("output does not contain %q:\n%s", want, text)
		}
	}
}

func TestInteractiveAcceptsEquipmentApplication(t *testing.T) {
	app, out, _ := newApp(t)

	// Заявка на зарядную станцию: без техосмотра, номер присваивается сразу.
	input := "3\n3\n\n3.2\n0\n"
	if err := app.Run(strings.NewReader(input)); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	text := out.String()
	for _, want := range []string{
		"Заявка принята. Присвоен номер: STATION-001 — зарядная станция",
		"Всего вещей на учёте: 1",
		"Оборудование: 1",
		"Зарядные станции: 1",
		"Суммарное суточное потребление: 3.20 кВт·ч",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("output does not contain %q:\n%s", want, text)
		}
	}
}
