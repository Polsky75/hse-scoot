// Package source описывает порт загрузки содержимого инвентаря.
//
// Прикладной слой зависит только от интерфейса Source и не знает,
// откуда берутся данные: текстовый файл, JSON, БД или сеть.
package source

import "github.com/Polsky75/hse-scoot/internal/domain"

// DeviceSpec — описание транспорта во входных данных.
type DeviceSpec struct {
	Kind                domain.TransportKind
	InventoryNumber     string
	DailyConsumptionKWh float64
	Condition           int
	EaseOfUse           int
}

// EquipmentSpec — описание снаряжения во входных данных.
type EquipmentSpec struct {
	Kind                domain.EquipmentKind
	InventoryNumber     string
	DailyConsumptionKWh float64
}

// Dataset — начальный парк, новые устройства и снаряжение.
type Dataset struct {
	// Registered — транспорт, уже стоящий на балансе.
	// Он восстанавливается без повторного техосмотра.
	Registered []DeviceSpec
	// Incoming — новый транспорт: перед постановкой на учёт
	// каждое устройство проходит техосмотр.
	Incoming []DeviceSpec
	// Equipment — шлемы, наколенники и зарядные станции.
	// Это вещи инвентаря, а не транспорт.
	Equipment []EquipmentSpec
}

// Source загружает содержимое инвентаря.
type Source interface {
	Load() (Dataset, error)
}
