package domain

import (
	"errors"
	"fmt"
	"strings"
)

// ITransport описывает транспорт в инвентаре.
//
// Любой транспорт одновременно является IInventoriable.
type ITransport interface {
	IInventoriable
	// GetEaseOfUse возвращает лёгкость использования от 1 до 10.
	GetEaseOfUse() int
}

// TransportKind — вид транспорта в парке.
type TransportKind string

const (
	KindScooter TransportKind = "scooter"
	KindBicycle TransportKind = "bicycle"
	KindEBike   TransportKind = "ebike"
)

var (
	// ErrUnknownTransportKind возвращается для неподдерживаемого вида транспорта.
	ErrUnknownTransportKind = errors.New("unknown transport kind")
	// ErrBicycleConsumption возвращается, если велосипеду задан расход энергии.
	ErrBicycleConsumption = errors.New("bicycle must not consume electricity")
)

// ParseTransportKind разбирает вид транспорта из текстового представления.
func ParseTransportKind(value string) (TransportKind, error) {
	kind := TransportKind(strings.ToLower(strings.TrimSpace(value)))
	switch kind {
	case KindScooter, KindBicycle, KindEBike:
		return kind, nil
	default:
		return "", fmt.Errorf("%w: %q", ErrUnknownTransportKind, value)
	}
}

// String возвращает текстовое представление вида транспорта.
func (k TransportKind) String() string {
	return string(k)
}

// NewTransport создаёт транспорт нужного вида.
//
// Чтобы добавить новый вид, достаточно добавить константу и ветку в
// ParseTransportKind/NewTransport: остальные слои работают с интерфейсами
// и не зависят от конкретных типов транспорта.
func NewTransport(
	kind TransportKind,
	inventoryNumber string,
	dailyConsumptionKWh float64,
	easeOfUse int,
) (ITransport, error) {
	switch kind {
	case KindScooter:
		return NewScooter(inventoryNumber, dailyConsumptionKWh, easeOfUse)
	case KindBicycle:
		if dailyConsumptionKWh != 0 {
			return nil, fmt.Errorf("%w: %q", ErrBicycleConsumption, inventoryNumber)
		}
		return NewBicycle(inventoryNumber, easeOfUse)
	case KindEBike:
		return NewEBike(inventoryNumber, dailyConsumptionKWh, easeOfUse)
	default:
		return nil, fmt.Errorf("%w: %q", ErrUnknownTransportKind, kind)
	}
}
