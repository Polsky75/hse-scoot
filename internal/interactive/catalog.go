package interactive

import "github.com/Polsky75/hse-scoot/internal/domain"

// itemKind — пункт нумерованного списка видов.
type itemKind struct {
	category   domain.ItemCategory
	transport  domain.TransportKind
	equipment  domain.EquipmentKind
	usesEnergy bool
}

var transportCatalog = []itemKind{
	{category: domain.CategoryScooter, transport: domain.KindScooter, usesEnergy: true},
	{category: domain.CategoryBicycle, transport: domain.KindBicycle},
	{category: domain.CategoryEBike, transport: domain.KindEBike, usesEnergy: true},
}

var equipmentCatalog = []itemKind{
	{category: domain.CategoryHelmet, equipment: domain.KindHelmet},
	{category: domain.CategoryKneePads, equipment: domain.KindKneePads},
	{category: domain.CategoryChargingStation, equipment: domain.KindChargingStation, usesEnergy: true},
}

func (k itemKind) title() string {
	return k.category.Label()
}

func (k itemKind) isTransport() bool {
	return k.transport != ""
}
