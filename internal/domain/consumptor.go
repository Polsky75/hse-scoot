// Package domain содержит бизнес-сущности и их контракты.
package domain

// IConsumptor описывает устройство, которое потребляет электроэнергию.
type IConsumptor interface {
	// GetDailyConsumptionKWh возвращает суточный расход энергии в кВт·ч.
	GetDailyConsumptionKWh() float64
}
