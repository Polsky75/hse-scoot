package interactive

import (
	"fmt"

	"github.com/Polsky75/hse-scoot/internal/domain"
)

type numberAllocator struct {
	used map[string]struct{}
}

func newNumberAllocator() *numberAllocator {
	return &numberAllocator{used: make(map[string]struct{})}
}

// Next возвращает первый свободный автоматический номер для категории.
// Номер не занимается до успешной регистрации.
func (n *numberAllocator) Next(category domain.ItemCategory) string {
	for i := 1; ; i++ {
		number := fmt.Sprintf("%s-%03d", categoryPrefix(category), i)
		if _, exists := n.used[number]; !exists {
			return number
		}
	}
}

// EnsureAvailable проверяет, что номер ещё не занят успешной регистрацией.
func (n *numberAllocator) EnsureAvailable(number string) error {
	if _, exists := n.used[number]; exists {
		return fmt.Errorf("номер %q уже используется", number)
	}
	return nil
}

// Commit окончательно занимает номер после успешной регистрации.
func (n *numberAllocator) Commit(number string) {
	n.used[number] = struct{}{}
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
