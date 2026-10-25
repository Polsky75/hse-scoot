package domain

// ItemCategory — категория вещи в инвентаре.
type ItemCategory string

const (
	CategoryScooter         ItemCategory = "scooter"
	CategoryBicycle         ItemCategory = "bicycle"
	CategoryEBike           ItemCategory = "ebike"
	CategoryHelmet          ItemCategory = "helmet"
	CategoryKneePads        ItemCategory = "kneepads"
	CategoryChargingStation ItemCategory = "charging_station"
)

// String возвращает код категории.
func (c ItemCategory) String() string {
	return string(c)
}

// LabelPlural возвращает название категории для статистики.
func (c ItemCategory) LabelPlural() string {
	switch c {
	case CategoryScooter:
		return "Самокаты"
	case CategoryBicycle:
		return "Велосипеды"
	case CategoryEBike:
		return "Электровелосипеды"
	case CategoryHelmet:
		return "Шлемы"
	case CategoryKneePads:
		return "Наколенники"
	case CategoryChargingStation:
		return "Зарядные станции"
	default:
		return string(c)
	}
}

// Label возвращает человекочитаемое название категории.
func (c ItemCategory) Label() string {
	switch c {
	case CategoryScooter:
		return "самокат"
	case CategoryBicycle:
		return "велосипед"
	case CategoryEBike:
		return "электровелосипед"
	case CategoryHelmet:
		return "шлем"
	case CategoryKneePads:
		return "наколенники"
	case CategoryChargingStation:
		return "зарядная станция"
	default:
		return string(c)
	}
}
