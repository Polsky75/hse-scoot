package domain

import (
	"errors"
	"fmt"
	"strings"
)

const (
	// MinEaseOfUse — минимальная лёгкость использования транспорта.
	MinEaseOfUse = 1
	// MaxEaseOfUse — максимальная лёгкость использования транспорта.
	MaxEaseOfUse = 10
)

var (
	// ErrEmptyInventoryNumber возвращается, если инвентарный номер пуст.
	ErrEmptyInventoryNumber = errors.New("inventory number must not be empty")
	// ErrEmptyCategory возвращается, если категория вещи не задана.
	ErrEmptyCategory = errors.New("item category must not be empty")
	// ErrNegativeConsumption возвращается при отрицательном суточном расходе.
	ErrNegativeConsumption = errors.New("daily consumption must not be negative")
	// ErrInvalidEaseOfUse возвращается для лёгкости использования вне 1..10.
	ErrInvalidEaseOfUse = errors.New("ease of use must be between 1 and 10")
)

// inventoryItem — общая часть всех вещей в инвентаре.
type inventoryItem struct {
	inventoryNumber string
	category        ItemCategory
}

func newInventoryItem(inventoryNumber string, category ItemCategory) (inventoryItem, error) {
	inventoryNumber = strings.TrimSpace(inventoryNumber)
	if inventoryNumber == "" {
		return inventoryItem{}, ErrEmptyInventoryNumber
	}
	if category == "" {
		return inventoryItem{}, ErrEmptyCategory
	}
	return inventoryItem{
		inventoryNumber: inventoryNumber,
		category:        category,
	}, nil
}

func (i inventoryItem) GetInventoryNumber() string {
	return i.inventoryNumber
}

func (i inventoryItem) GetCategory() ItemCategory {
	return i.category
}

// transportDevice — общая часть транспорта: вещь на балансе и лёгкость.
type transportDevice struct {
	inventoryItem
	easeOfUse int
}

func newTransportDevice(inventoryNumber string, category ItemCategory, easeOfUse int) (transportDevice, error) {
	item, err := newInventoryItem(inventoryNumber, category)
	if err != nil {
		return transportDevice{}, err
	}
	if easeOfUse < MinEaseOfUse || easeOfUse > MaxEaseOfUse {
		return transportDevice{}, fmt.Errorf("%w: %d", ErrInvalidEaseOfUse, easeOfUse)
	}
	return transportDevice{
		inventoryItem: item,
		easeOfUse:     easeOfUse,
	}, nil
}

func (d transportDevice) GetEaseOfUse() int {
	return d.easeOfUse
}

// electricDevice добавляет к транспорту суточный расход электроэнергии.
type electricDevice struct {
	transportDevice
	dailyConsumptionKWh float64
}

func newElectricDevice(
	inventoryNumber string,
	category ItemCategory,
	dailyConsumptionKWh float64,
	easeOfUse int,
) (electricDevice, error) {
	transport, err := newTransportDevice(inventoryNumber, category, easeOfUse)
	if err != nil {
		return electricDevice{}, err
	}
	if dailyConsumptionKWh < 0 {
		return electricDevice{}, fmt.Errorf("%w: %g", ErrNegativeConsumption, dailyConsumptionKWh)
	}

	return electricDevice{
		transportDevice:     transport,
		dailyConsumptionKWh: dailyConsumptionKWh,
	}, nil
}

func (d electricDevice) GetDailyConsumptionKWh() float64 {
	return d.dailyConsumptionKWh
}
