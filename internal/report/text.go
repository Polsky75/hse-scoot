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

	writeSection(&out, document.AllItems)
	writeSection(&out, document.Transport)
	writeSection(&out, document.BeginnerTransport)

	return out.String()
}

func writeSection(out *strings.Builder, section Section) {
	fmt.Fprintf(out, "\n%s (%d):\n", section.Title, len(section.Items))
	if len(section.Items) == 0 {
		out.WriteString("  —\n")
		return
	}

	for _, item := range section.Items {
		fmt.Fprintf(out, "  %s — %s\n", item.InventoryNumber, item.Category.Label())
	}
}
