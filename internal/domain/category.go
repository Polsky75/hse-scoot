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
