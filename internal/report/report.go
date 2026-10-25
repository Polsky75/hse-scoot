// Package report собирает сводный отчёт по инвентарю и выводит его
// в выбранное место.
package report

import "github.com/Polsky75/hse-scoot/internal/domain"

// Counts — счёт вещей на учёте.
type Counts struct {
	Total                    int
	Transport                int
	Scooters                 int
	KneePads                 int
	TotalDailyConsumptionKWh float64
}

// Item — строка отчёта по одной вещи.
type Item struct {
	InventoryNumber string
	Category        domain.ItemCategory
}

// Document — готовый отчёт по инвентарю.
type Document struct {
	Counts            Counts
	Equipment         []Item
	Transport         []Item
	BeginnerTransport []Item
}

// Build формирует отчёт: оборудование (всё, что не транспорт), транспорт
// и отдельно транспорт, подходящий новичкам.
func Build(items, beginnerItems []domain.IInventoriable) Document {
	document := Document{
		Counts: Counts{
			Total:                    CountAll(items),
			Transport:                CountTransport(items),
			Scooters:                 CountCategory(items, domain.CategoryScooter),
			KneePads:                 CountCategory(items, domain.CategoryKneePads),
			TotalDailyConsumptionKWh: CountDailyConsumptionKWh(items),
		},
	}

	for _, item := range items {
		view := newItem(item)
		if _, ok := item.(domain.ITransport); ok {
			document.Transport = append(document.Transport, view)
			continue
		}
		document.Equipment = append(document.Equipment, view)
	}

	for _, item := range beginnerItems {
		if _, ok := item.(domain.ITransport); !ok {
			continue
		}
		document.BeginnerTransport = append(document.BeginnerTransport, newItem(item))
	}

	return document
}

// CountAll возвращает количество всех вещей.
func CountAll(items []domain.IInventoriable) int {
	return len(items)
}

// CountTransport возвращает количество транспорта.
func CountTransport(items []domain.IInventoriable) int {
	var count int
	for _, item := range items {
		if _, ok := item.(domain.ITransport); ok {
			count++
		}
	}
	return count
}

// CountDailyConsumptionKWh возвращает суммарный суточный расход всех
// потребителей энергии в инвентаре.
func CountDailyConsumptionKWh(items []domain.IInventoriable) float64 {
	var total float64
	for _, item := range items {
		consumer, ok := item.(domain.IConsumptor)
		if !ok {
			continue
		}
		total += consumer.GetDailyConsumptionKWh()
	}
	return total
}

// CountCategory возвращает количество вещей указанной категории.
func CountCategory(items []domain.IInventoriable, category domain.ItemCategory) int {
	var count int
	for _, item := range items {
		if item.GetCategory() == category {
			count++
		}
	}
	return count
}

func newItem(item domain.IInventoriable) Item {
	return Item{
		InventoryNumber: item.GetInventoryNumber(),
		Category:        item.GetCategory(),
	}
}
