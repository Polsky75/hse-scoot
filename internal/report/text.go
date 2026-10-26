package report

import (
	"fmt"
	"strings"
)

// RenderText формирует текстовое представление отчёта.
func RenderText(document Document) string {
	var out strings.Builder

	out.WriteString("hse-scoot — отчёт по инвентарю\n\n")
	fmt.Fprintf(&out, "Всего вещей на учёте: %d\n", document.Counts.Total)
	fmt.Fprintf(&out, "Оборудование: %d\n", document.Counts.Equipment)
	fmt.Fprintf(&out, "Транспорт для профессионалов: %d\n", document.Counts.ProfessionalTransport)
	fmt.Fprintf(&out, "Транспорт для новичков: %d\n", document.Counts.BeginnerTransport)
	for _, count := range document.Counts.Categories {
		if count.Count == 0 {
			continue
		}
		fmt.Fprintf(&out, "%s: %d\n", count.Category.LabelPlural(), count.Count)
	}
	fmt.Fprintf(&out, "Суммарное суточное потребление: %.2f кВт·ч\n", document.Counts.TotalDailyConsumptionKWh)

	writeSection(&out, "Оборудование", document.Equipment)
	writeSection(&out, "Транспорт для профессионалов", document.ProfessionalTransport)
	writeSection(&out, "Транспорт для новичков", document.BeginnerTransport)

	return out.String()
}

func writeSection(out *strings.Builder, title string, items []Item) {
	fmt.Fprintf(out, "\n%s (%d):\n", title, len(items))
	if len(items) == 0 {
		out.WriteString("  —\n")
		return
	}

	for _, item := range items {
		fmt.Fprintf(out, "  %s — %s\n", item.InventoryNumber, item.Category.Label())
	}
}
