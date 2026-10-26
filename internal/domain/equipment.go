package domain

import (
	"errors"
	"fmt"
	"strings"
)

// EquipmentKind — вид снаряжения в инвентаре.
type EquipmentKind string

const (
	KindHelmet          EquipmentKind = "helmet"
	KindKneePads        EquipmentKind = "kneepads"
	KindChargingStation EquipmentKind = "charging_station"
)

var (
	// ErrUnknownEquipmentKind возвращается для неподдерживаемого вида снаряжения.
	ErrUnknownEquipmentKind = errors.New("unknown equipment kind")
	// ErrEquipmentConsumption возвращается, если расход задан не зарядной станции.
	ErrEquipmentConsumption = errors.New("only charging station may consume electricity")
)

// ParseEquipmentKind разбирает вид снаряжения из текстового представления.
func ParseEquipmentKind(value string) (EquipmentKind, error) {
	kind := EquipmentKind(strings.ToLower(strings.TrimSpace(value)))
	switch kind {
	case KindHelmet, KindKneePads, KindChargingStation:
		return kind, nil
	default:
		return "", fmt.Errorf("%w: %q", ErrUnknownEquipmentKind, value)
	}
}

// NewEquipment создаёт вещь инвентаря, не являющуюся транспортом.
//
// Расход энергии допустим только для зарядной станции.
func NewEquipment(
	kind EquipmentKind,
	inventoryNumber string,
	dailyConsumptionKWh float64,
) (IInventoriable, error) {
	switch kind {
	case KindHelmet:
		if dailyConsumptionKWh != 0 {
			return nil, fmt.Errorf("%w: %q", ErrEquipmentConsumption, inventoryNumber)
		}
		return NewHelmet(inventoryNumber)
	case KindKneePads:
		if dailyConsumptionKWh != 0 {
			return nil, fmt.Errorf("%w: %q", ErrEquipmentConsumption, inventoryNumber)
		}
		return NewKneePads(inventoryNumber)
	case KindChargingStation:
		return NewChargingStation(inventoryNumber, dailyConsumptionKWh)
	default:
		return nil, fmt.Errorf("%w: %q", ErrUnknownEquipmentKind, kind)
	}
}

// equipmentItem — общая часть снаряжения; в отличие от транспорта
// не имеет лёгкости использования.
type equipmentItem struct {
	inventoryItem
}

func newEquipmentItem(inventoryNumber string, category ItemCategory) (equipmentItem, error) {
	item, err := newInventoryItem(inventoryNumber, category)
	if err != nil {
		return equipmentItem{}, err
	}
	return equipmentItem{inventoryItem: item}, nil
}

// Helmet — шлем.
type Helmet struct {
	equipmentItem
}

// NewHelmet создаёт шлем.
func NewHelmet(inventoryNumber string) (*Helmet, error) {
	item, err := newEquipmentItem(inventoryNumber, CategoryHelmet)
	if err != nil {
		return nil, err
	}
	return &Helmet{equipmentItem: item}, nil
}

var _ IInventoriable = (*Helmet)(nil)

// KneePads — наколенники.
type KneePads struct {
	equipmentItem
}

// NewKneePads создаёт наколенники.
func NewKneePads(inventoryNumber string) (*KneePads, error) {
	item, err := newEquipmentItem(inventoryNumber, CategoryKneePads)
	if err != nil {
		return nil, err
	}
	return &KneePads{equipmentItem: item}, nil
}

var _ IInventoriable = (*KneePads)(nil)

// ChargingStation — зарядная станция.
//
// Числится на балансе и потребляет электроэнергию.
type ChargingStation struct {
	equipmentItem
	dailyConsumptionKWh float64
}

// NewChargingStation создаёт зарядную станцию.
func NewChargingStation(inventoryNumber string, dailyConsumptionKWh float64) (*ChargingStation, error) {
	item, err := newEquipmentItem(inventoryNumber, CategoryChargingStation)
	if err != nil {
		return nil, err
	}
	if err := validateConsumption(dailyConsumptionKWh); err != nil {
		return nil, err
	}
	return &ChargingStation{
		equipmentItem:       item,
		dailyConsumptionKWh: dailyConsumptionKWh,
	}, nil
}

func (c *ChargingStation) GetDailyConsumptionKWh() float64 {
	return c.dailyConsumptionKWh
}

var (
	_ IInventoriable = (*ChargingStation)(nil)
	_ IConsumptor    = (*ChargingStation)(nil)
)
