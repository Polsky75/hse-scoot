package domain

// IInventoriable описывает любую вещь в инвентаре сервиса.
//
// Транспорт и снаряжение хранятся в инвентаре именно по этому контракту.
type IInventoriable interface {
	// GetInventoryNumber возвращает инвентарный номер вещи.
	GetInventoryNumber() string
	// GetCategory возвращает категорию вещи.
	GetCategory() ItemCategory
}
