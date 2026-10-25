package interactive

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Polsky75/hse-scoot/internal/domain"
)

type numberAllocator struct {
	used     map[string]struct{}
	counters map[domain.ItemCategory]int
}

func newNumberAllocator() *numberAllocator {
	return &numberAllocator{
		used:     make(map[string]struct{}),
		counters: make(map[domain.ItemCategory]int),
	}
}

// Next возвращает следующий свободный автоматический номер для категории.
func (n *numberAllocator) Next(category domain.ItemCategory) string {
	for {
		n.counters[category]++
		number := fmt.Sprintf("%s-%03d", categoryPrefix(category), n.counters[category])
		if _, exists := n.used[number]; exists {
			continue
		}
		n.used[number] = struct{}{}
		return number
	}
}

// Reserve занимает номер, введённый пользователем.
func (n *numberAllocator) Reserve(number string) error {
	if strings.ContainsAny(number, " \t") {
		return errors.New("номер не должен содержать пробелов")
	}
	if _, exists := n.used[number]; exists {
		return fmt.Errorf("номер %q уже используется", number)
	}
	n.used[number] = struct{}{}
	return nil
}

func categoryPrefix(category domain.ItemCategory) string {
	switch category {
	case domain.CategoryScooter:
		return "SCOOT"
	case domain.CategoryBicycle:
		return "BIKE"
	case domain.CategoryEBike:
		return "EBIKE"
	case domain.CategoryHelmet:
		return "HELMET"
	case domain.CategoryKneePads:
		return "KNEE"
	case domain.CategoryChargingStation:
		return "STATION"
	default:
		return "INV"
	}
}
