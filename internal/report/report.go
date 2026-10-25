// Package report собирает сводный отчёт по инвентарю и выводит его
// в выбранное место.
package report

import "github.com/Polsky75/hse-scoot/internal/domain"

// CategoryCount — количество вещей одной категории.
type CategoryCount struct {
	Category domain.ItemCategory
	Count    int
}

// Counts — счёт вещей на учёте.
//
// Categories хранит разбивку по всем категориям в порядке первого
// появления в инвентаре, поэтому новые виды не требуют правок отчёта.
type Counts struct {
	Total                    int
	Transport                int
	Equipment                int
	Categories               []CategoryCount
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
			Equipment:                CountEquipment(items),
			Categories:               CountCategories(items),
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

// CountEquipment возвращает количество вещей, не являющихся транспортом.
func CountEquipment(items []domain.IInventoriable) int {
	var count int
	for _, item := range items {
		if _, ok := item.(domain.ITransport); !ok {
			count++
		}
	}
	return count
}

// CountCategories возвращает количество вещей по категориям в порядке
// первого появления. Отчёт не знает конкретных видов вещей.
func CountCategories(items []domain.IInventoriable) []CategoryCount {
	counts := make(map[domain.ItemCategory]int, len(items))
	order := make([]domain.ItemCategory, 0, len(items))
	for _, item := range items {
		category := item.GetCategory()
		if _, exists := counts[category]; !exists {
			order = append(order, category)
		}
		counts[category]++
	}

	result := make([]CategoryCount, 0, len(order))
	for _, category := range order {
		result = append(result, CategoryCount{Category: category, Count: counts[category]})
	}
	return result
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
