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
	fmt.Fprintf(&out, "Транспорт: %d\n", document.Counts.Transport)
	fmt.Fprintf(&out, "Самокаты: %d\n", document.Counts.Scooters)
	fmt.Fprintf(&out, "Наколенники: %d\n", document.Counts.KneePads)

	writeSection(&out, "Вся техника", document.AllItems)
	writeSection(&out, "Транспорт", document.Transport)
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
