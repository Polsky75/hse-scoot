// Package inspection содержит процедуру технического осмотра устройств.
package inspection

import (
	"errors"
	"fmt"

	"github.com/Polsky75/hse-scoot/internal/domain"
)

// DefaultMinCondition — минимальная оценка состояния для допуска на учёт.
const DefaultMinCondition = 70

var (
	// ErrFailed возвращается, если устройство не прошло техосмотр.
	ErrFailed = errors.New("техосмотр не пройден")
	// ErrInvalidCondition возвращается для оценки вне диапазона 0..100.
	ErrInvalidCondition = errors.New("condition must be between 0 and 100")
)

// Inspector проверяет техническое состояние устройства.
type Inspector struct {
	minCondition int
}

// NewInspector создаёт инспектора с минимальной допустимой оценкой.
func NewInspector(minCondition int) (*Inspector, error) {
	if minCondition < 0 || minCondition > 100 {
		return nil, fmt.Errorf("%w: %d", ErrInvalidCondition, minCondition)
	}
	return &Inspector{minCondition: minCondition}, nil
}

// Inspect проверяет устройство и возвращает ошибку, если осмотр не пройден.
func (i *Inspector) Inspect(device domain.IInventoriable, condition int) error {
	if condition < 0 || condition > 100 {
		return fmt.Errorf("%w: %d", ErrInvalidCondition, condition)
	}
	if condition < i.minCondition {
		return fmt.Errorf(
			"%w: устройство %s, состояние %d из 100, требуется не менее %d",
			ErrFailed, device.GetInventoryNumber(), condition, i.minCondition,
		)
	}
	return nil
}
