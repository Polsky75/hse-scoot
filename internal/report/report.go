// Package report собирает сводный отчёт по инвентарю и выводит его
// в выбранное место.
package report

import "github.com/Polsky75/hse-scoot/internal/domain"

// Counts — счёт вещей на учёте.
type Counts struct {
	Total     int
	Transport int
	Scooters  int
	KneePads  int
}

// Item — строка отчёта по одной вещи.
type Item struct {
	InventoryNumber string
	Category        domain.ItemCategory
}

// Document — готовый отчёт по инвентарю.
type Document struct {
	Counts            Counts
	AllItems          []Item
	Transport         []Item
	BeginnerTransport []Item
}

// Build формирует отчёт по всем вещам и отдельно по подходящему новичкам
// транспорту.
func Build(items, beginnerItems []domain.IInventoriable) Document {
	document := Document{
		Counts: Counts{
			Total:     CountAll(items),
			Transport: CountTransport(items),
			Scooters:  CountCategory(items, domain.CategoryScooter),
			KneePads:  CountCategory(items, domain.CategoryKneePads),
		},
	}

	for _, item := range items {
		view := newItem(item)
		document.AllItems = append(document.AllItems, view)
		if _, ok := item.(domain.ITransport); ok {
			document.Transport = append(document.Transport, view)
		}
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
