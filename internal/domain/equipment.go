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

// ErrUnknownEquipmentKind возвращается для неподдерживаемого вида снаряжения.
var ErrUnknownEquipmentKind = errors.New("unknown equipment kind")

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

// String возвращает текстовое представление вида снаряжения.
func (k EquipmentKind) String() string {
	return string(k)
}

// NewEquipment создаёт вещь инвентаря, не являющуюся транспортом.
func NewEquipment(kind EquipmentKind, inventoryNumber string) (IInventoriable, error) {
	switch kind {
	case KindHelmet:
		return NewHelmet(inventoryNumber)
	case KindKneePads:
		return NewKneePads(inventoryNumber)
	case KindChargingStation:
		return NewChargingStation(inventoryNumber)
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
type ChargingStation struct {
	equipmentItem
}

// NewChargingStation создаёт зарядную станцию.
func NewChargingStation(inventoryNumber string) (*ChargingStation, error) {
	item, err := newEquipmentItem(inventoryNumber, CategoryChargingStation)
	if err != nil {
		return nil, err
	}
	return &ChargingStation{equipmentItem: item}, nil
}

var _ IInventoriable = (*ChargingStation)(nil)
