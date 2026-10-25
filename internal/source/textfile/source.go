// Package textfile реализует source.Source для построчного текстового формата.
package textfile

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/Polsky75/hse-scoot/internal/domain"
	"github.com/Polsky75/hse-scoot/internal/source"
)

const (
	sectionRegistered = "registered"
	sectionIncoming   = "incoming"
	sectionEquipment  = "equipment"
)

// Source читает инвентарь из текстового файла.
type Source struct {
	path string
}

// New создаёт источник, читающий указанный файл.
func New(path string) *Source {
	return &Source{path: path}
}

// Load читает и валидирует файл.
//
// Транспорт: kind inventory_number daily_consumption_kwh condition ease_of_use
// в секциях [registered] и [incoming].
// Снаряжение: kind inventory_number в секции [equipment].
// Пустые строки и строки, начинающиеся с #, игнорируются.
func (s *Source) Load() (source.Dataset, error) {
	file, err := os.Open(s.path)
	if err != nil {
		return source.Dataset{}, fmt.Errorf("open inventory file %q: %w", s.path, err)
	}
	defer file.Close()

	var dataset source.Dataset
	var section string
	seen := make(map[string]int)
	scanner := bufio.NewScanner(file)
	lineNumber := 0

	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(strings.TrimPrefix(scanner.Text(), "\ufeff"))
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if inner, ok := strings.CutPrefix(line, "["); ok {
			rawSection, closed := strings.CutSuffix(inner, "]")
			if !closed {
				return source.Dataset{}, fmt.Errorf("%s:%d: malformed section header %q", s.path, lineNumber, line)
			}
			section = strings.TrimSpace(rawSection)
			switch section {
			case sectionRegistered, sectionIncoming, sectionEquipment:
			default:
				return source.Dataset{}, fmt.Errorf("%s:%d: unknown section %q", s.path, lineNumber, section)
			}
			continue
		}

		if section == "" {
			return source.Dataset{}, fmt.Errorf("%s:%d: item %q is outside a section", s.path, lineNumber, line)
		}

		if section == sectionEquipment {
			spec, err := parseEquipmentSpec(line)
			if err != nil {
				return source.Dataset{}, fmt.Errorf("%s:%d: %w", s.path, lineNumber, err)
			}
			if err := registerInventoryNumber(seen, spec.InventoryNumber, lineNumber, s.path); err != nil {
				return source.Dataset{}, err
			}
			dataset.Equipment = append(dataset.Equipment, spec)
			continue
		}

		spec, err := parseDeviceSpec(line)
		if err != nil {
			return source.Dataset{}, fmt.Errorf("%s:%d: %w", s.path, lineNumber, err)
		}
		if err := registerInventoryNumber(seen, spec.InventoryNumber, lineNumber, s.path); err != nil {
			return source.Dataset{}, err
		}

		if section == sectionRegistered {
			dataset.Registered = append(dataset.Registered, spec)
		} else {
			dataset.Incoming = append(dataset.Incoming, spec)
		}
	}

	if err := scanner.Err(); err != nil {
		return source.Dataset{}, fmt.Errorf("read inventory file %q: %w", s.path, err)
	}
	return dataset, nil
}

func registerInventoryNumber(seen map[string]int, inventoryNumber string, lineNumber int, path string) error {
	if firstLine, ok := seen[inventoryNumber]; ok {
		return fmt.Errorf(
			"%s:%d: duplicate inventory number %q (first seen at line %d)",
			path, lineNumber, inventoryNumber, firstLine,
		)
	}
	seen[inventoryNumber] = lineNumber
	return nil
}

func parseEquipmentSpec(line string) (source.EquipmentSpec, error) {
	fields := strings.Fields(line)
	if len(fields) < 2 {
		return source.EquipmentSpec{}, fmt.Errorf(
			"expected at least 2 fields (kind, inventory_number), got %d", len(fields),
		)
	}

	kind, err := domain.ParseEquipmentKind(fields[0])
	if err != nil {
		return source.EquipmentSpec{}, err
	}

	spec := source.EquipmentSpec{
		Kind:            kind,
		InventoryNumber: fields[1],
	}
	if kind != domain.KindChargingStation {
		if len(fields) != 2 {
			return source.EquipmentSpec{}, fmt.Errorf(
				"expected 2 fields (kind, inventory_number) for %s, got %d",
				kind, len(fields),
			)
		}
		return spec, nil
	}

	if len(fields) != 3 {
		return source.EquipmentSpec{}, fmt.Errorf(
			"expected 3 fields (kind, inventory_number, daily_consumption_kwh) for %s, got %d",
			kind, len(fields),
		)
	}
	consumption, err := strconv.ParseFloat(fields[2], 64)
	if err != nil {
		return source.EquipmentSpec{}, fmt.Errorf("invalid daily consumption %q: %w", fields[2], err)
	}
	if math.IsNaN(consumption) || math.IsInf(consumption, 0) || consumption < 0 {
		return source.EquipmentSpec{}, fmt.Errorf(
			"daily consumption must be a non-negative finite number, got %q", fields[2],
		)
	}
	spec.DailyConsumptionKWh = consumption
	return spec, nil
}

func parseDeviceSpec(line string) (source.DeviceSpec, error) {
	fields := strings.Fields(line)
	if len(fields) != 5 {
		return source.DeviceSpec{}, fmt.Errorf(
			"expected 5 fields (kind, inventory_number, daily_consumption_kwh, condition, ease_of_use), got %d",
			len(fields),
		)
	}

	kind, err := domain.ParseTransportKind(fields[0])
	if err != nil {
		return source.DeviceSpec{}, err
	}

	consumption, err := strconv.ParseFloat(fields[2], 64)
	if err != nil {
		return source.DeviceSpec{}, fmt.Errorf("invalid daily consumption %q: %w", fields[2], err)
	}
	if math.IsNaN(consumption) || math.IsInf(consumption, 0) || consumption < 0 {
		return source.DeviceSpec{}, fmt.Errorf(
			"daily consumption must be a non-negative finite number, got %q", fields[2],
		)
	}

	condition, err := strconv.Atoi(fields[3])
	if err != nil {
		return source.DeviceSpec{}, fmt.Errorf("invalid condition %q: %w", fields[3], err)
	}
	if condition < 0 || condition > 100 {
		return source.DeviceSpec{}, fmt.Errorf("condition must be between 0 and 100, got %d", condition)
	}

	easeOfUse, err := strconv.Atoi(fields[4])
	if err != nil {
		return source.DeviceSpec{}, fmt.Errorf("invalid ease of use %q: %w", fields[4], err)
	}
	if easeOfUse < domain.MinEaseOfUse || easeOfUse > domain.MaxEaseOfUse {
		return source.DeviceSpec{}, fmt.Errorf(
			"ease of use must be between %d and %d, got %d",
			domain.MinEaseOfUse, domain.MaxEaseOfUse, easeOfUse,
		)
	}

	return source.DeviceSpec{
		Kind:                kind,
		InventoryNumber:     fields[1],
		DailyConsumptionKWh: consumption,
		Condition:           condition,
		EaseOfUse:           easeOfUse,
	}, nil
}
