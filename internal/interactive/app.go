// Package interactive реализует консольный ввод инвентаря без файла.
package interactive

import (
	"bufio"
	"errors"
	"fmt"
	"io"

	"github.com/Polsky75/hse-scoot/internal/domain"
	"github.com/Polsky75/hse-scoot/internal/fleet"
	"github.com/Polsky75/hse-scoot/internal/report"
)

// App принимает данные из консоли и сразу регистрирует введённые вещи.
type App struct {
	fleet   *fleet.Service
	output  report.Output
	out     io.Writer
	numbers *numberAllocator
}

// New создаёт интерактивное приложение.
func New(fleetService *fleet.Service, output report.Output, out io.Writer) *App {
	return &App{
		fleet:   fleetService,
		output:  output,
		out:     out,
		numbers: newNumberAllocator(),
	}
}

// Run читает команды до выбора «Завершить» или конца ввода.
func (a *App) Run(in io.Reader) error {
	scanner := bufio.NewScanner(in)
	fmt.Fprintln(a.out, "Интерактивный режим hse-scoot")

	for {
		fmt.Fprintln(a.out, "\nДействие:")
		fmt.Fprintln(a.out, "1. Добавить уже стоящую на учёте вещь")
		fmt.Fprintln(a.out, "2. Поставить транспорт на учёт")
		fmt.Fprintln(a.out, "3. Поставить оборудование на учёт")
		fmt.Fprintln(a.out, "0. Завершить и вывести отчёт")

		choice, err := readInt(scanner, a.out, "Выбор: ", 0, 3)
		if errors.Is(err, io.EOF) {
			return a.writeReport()
		}
		if err != nil {
			return err
		}

		switch choice {
		case 0:
			return a.writeReport()
		case 1:
			err = a.addRegistered(scanner)
		case 2:
			err = a.addTransportApplication(scanner)
		case 3:
			err = a.addEquipmentApplication(scanner)
		}
		if errors.Is(err, io.EOF) {
			return a.writeReport()
		}
		if err != nil {
			return err
		}
	}
}

func (a *App) addRegistered(scanner *bufio.Scanner) error {
	kinds, err := a.chooseGroup(scanner)
	if err != nil {
		return err
	}
	kind, err := a.chooseKind(scanner, kinds)
	if err != nil {
		return err
	}
	number, err := a.chooseNumber(scanner, kind.category)
	if err != nil {
		return err
	}

	item, _, err := a.buildItem(scanner, kind, number, false)
	if err != nil {
		return a.reportBuildError(err)
	}

	a.fleet.AddRegistered(item)
	a.numbers.Commit(item.GetInventoryNumber())
	fmt.Fprintf(
		a.out,
		"Добавлено на учёт. Номер: %s — %s\n",
		item.GetInventoryNumber(),
		item.GetCategory().Label(),
	)
	return nil
}

func (a *App) addTransportApplication(scanner *bufio.Scanner) error {
	kind, err := a.chooseKind(scanner, transportCatalog)
	if err != nil {
		return err
	}
	number, err := a.chooseNumber(scanner, kind.category)
	if err != nil {
		return err
	}

	item, condition, err := a.buildItem(scanner, kind, number, true)
	if err != nil {
		return a.reportBuildError(err)
	}
	transport, ok := item.(domain.ITransport)
	if !ok {
		fmt.Fprintln(a.out, "Ошибка: созданная вещь не является транспортом.")
		return nil
	}
	if err := a.fleet.RegisterTransport(transport, condition); err != nil {
		fmt.Fprintf(a.out, "Заявка отклонена: %v. Номер %s не присвоен.\n", err, number)
		return nil
	}
	a.numbers.Commit(item.GetInventoryNumber())

	fmt.Fprintf(
		a.out,
		"Заявка принята. Присвоен номер: %s — %s\n",
		item.GetInventoryNumber(),
		item.GetCategory().Label(),
	)
	return nil
}

func (a *App) addEquipmentApplication(scanner *bufio.Scanner) error {
	kind, err := a.chooseKind(scanner, equipmentCatalog)
	if err != nil {
		return err
	}
	number, err := a.chooseNumber(scanner, kind.category)
	if err != nil {
		return err
	}

	item, _, err := a.buildItem(scanner, kind, number, false)
	if err != nil {
		return a.reportBuildError(err)
	}

	// Оборудование не проходит техосмотр и принимается сразу.
	a.fleet.AddRegistered(item)
	a.numbers.Commit(item.GetInventoryNumber())
	fmt.Fprintf(
		a.out,
		"Заявка принята. Присвоен номер: %s — %s\n",
		item.GetInventoryNumber(),
		item.GetCategory().Label(),
	)
	return nil
}

func (a *App) chooseGroup(scanner *bufio.Scanner) ([]itemKind, error) {
	fmt.Fprintln(a.out, "\nЧто добавить:")
	fmt.Fprintln(a.out, "1. Транспорт")
	fmt.Fprintln(a.out, "2. Оборудование")

	choice, err := readInt(scanner, a.out, "Группа: ", 1, 2)
	if err != nil {
		return nil, err
	}
	if choice == 1 {
		return transportCatalog, nil
	}
	return equipmentCatalog, nil
}

func (a *App) chooseKind(scanner *bufio.Scanner, kinds []itemKind) (itemKind, error) {
	title := "Вид транспорта:"
	if !kinds[0].isTransport() {
		title = "Вид оборудования:"
	}
	fmt.Fprintf(a.out, "\n%s\n", title)
	for i, kind := range kinds {
		fmt.Fprintf(a.out, "%d. %s\n", i+1, kind.title())
	}

	choice, err := readInt(scanner, a.out, "Вид: ", 1, len(kinds))
	if err != nil {
		return itemKind{}, err
	}
	return kinds[choice-1], nil
}

func (a *App) chooseNumber(scanner *bufio.Scanner, category domain.ItemCategory) (string, error) {
	for {
		number, err := readString(
			scanner,
			a.out,
			"Инвентарный номер (Enter — присвоить автоматически): ",
		)
		if err != nil {
			return "", err
		}
		if number == "" {
			return a.numbers.Next(category), nil
		}
		if err := domain.ValidateInventoryNumber(number); err != nil {
			fmt.Fprintln(
				a.out,
				"Ошибка: номер может состоять только из заглавных латинских букв, цифр и тире. Повторите ввод.",
			)
			continue
		}
		if err := a.numbers.EnsureAvailable(number); err != nil {
			fmt.Fprintf(a.out, "Ошибка: %v. Повторите ввод.\n", err)
			continue
		}
		return number, nil
	}
}

func (a *App) buildItem(
	scanner *bufio.Scanner,
	kind itemKind,
	number string,
	withInspection bool,
) (domain.IInventoriable, int, error) {
	consumption, err := a.readConsumption(scanner, kind)
	if err != nil {
		return nil, 0, err
	}

	if !kind.isTransport() {
		item, err := domain.NewEquipment(kind.equipment, number, consumption)
		return item, 0, err
	}

	ease, err := readInt(scanner, a.out, "Лёгкость использования (1..10): ", 1, 10)
	if err != nil {
		return nil, 0, err
	}

	condition := 0
	if withInspection {
		condition, err = readInt(scanner, a.out, "Оценка техосмотра (0..100): ", 0, 100)
		if err != nil {
			return nil, 0, err
		}
	}

	transport, err := domain.NewTransport(kind.transport, number, consumption, ease)
	return transport, condition, err
}

func (a *App) readConsumption(scanner *bufio.Scanner, kind itemKind) (float64, error) {
	if !kind.usesEnergy {
		return 0, nil
	}
	return readFloat(scanner, a.out, "Суточный расход, кВт·ч: ")
}

func (a *App) reportBuildError(err error) error {
	if errors.Is(err, io.EOF) {
		return err
	}
	fmt.Fprintf(a.out, "Ошибка: %v.\n", err)
	return nil
}

func (a *App) writeReport() error {
	beginnerItems, _ := a.fleet.BeginnerDevices()
	document := report.Build(a.fleet.Items(), beginnerItems)
	if err := a.output.Write(document); err != nil {
		return fmt.Errorf("write report: %w", err)
	}
	return nil
}
