package fleet

import "github.com/Polsky75/hse-scoot/internal/domain"

// BeginnerMinEaseOfUse — минимальная лёгкость использования, при которой
// транспорт считается подходящим новичкам (строго больше 6).
const BeginnerMinEaseOfUse = 7

// BeginnerDevices выбирает транспорт, подходящий новичкам, и считает
// суммарный суточный расход энергии выбранных устройств.
//
// Лёгкость использования читается через ITransport, поэтому снаряжение
// (шлемы, наколенники, зарядные станции) в выборку не попадает.
// Потребители опрашиваются только через IConsumptor: велосипеды и другие
// неэлектрические устройства в сумму не входят.
func BeginnerDevices(items []domain.IInventoriable) ([]domain.IInventoriable, float64) {
	selected := make([]domain.IInventoriable, 0, len(items))
	var total float64

	for _, item := range items {
		transport, ok := item.(domain.ITransport)
		if !ok || transport.GetEaseOfUse() < BeginnerMinEaseOfUse {
			continue
		}
		selected = append(selected, item)

		consumer, ok := item.(domain.IConsumptor)
		if !ok {
			continue
		}
		total += consumer.GetDailyConsumptionKWh()
	}

	return selected, total
}

// BeginnerDevices возвращает транспорт текущего инвентаря, подходящий
// новичкам, и суммарный суточный расход энергии этих устройств.
func (s *Service) BeginnerDevices() ([]domain.IInventoriable, float64) {
	return BeginnerDevices(s.items)
}
