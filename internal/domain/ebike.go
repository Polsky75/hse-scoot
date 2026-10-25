package domain

// EBike — электровелосипед кикшеринг-сервиса.
//
// Числится на балансе и потребляет электроэнергию.
type EBike struct {
	electricDevice
}

// NewEBike создаёт электровелосипед и проверяет его инварианты.
func NewEBike(inventoryNumber string, dailyConsumptionKWh float64, easeOfUse int) (*EBike, error) {
	device, err := newElectricDevice(
		inventoryNumber,
		CategoryEBike,
		dailyConsumptionKWh,
		easeOfUse,
	)
	if err != nil {
		return nil, err
	}
	return &EBike{electricDevice: device}, nil
}

var (
	_ IConsumptor    = (*EBike)(nil)
	_ ITransport     = (*EBike)(nil)
	_ IInventoriable = (*EBike)(nil)
)
