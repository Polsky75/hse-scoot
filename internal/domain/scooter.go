package domain

// Scooter — самокат кикшеринг-сервиса.
//
// Самокат числится на балансе и потребляет электроэнергию.
type Scooter struct {
	electricDevice
}

// NewScooter создаёт самокат и проверяет его инварианты.
func NewScooter(inventoryNumber string, dailyConsumptionKWh float64, easeOfUse int) (*Scooter, error) {
	device, err := newElectricDevice(
		inventoryNumber,
		CategoryScooter,
		dailyConsumptionKWh,
		easeOfUse,
	)
	if err != nil {
		return nil, err
	}
	return &Scooter{electricDevice: device}, nil
}

var (
	_ IConsumptor    = (*Scooter)(nil)
	_ ITransport     = (*Scooter)(nil)
	_ IInventoriable = (*Scooter)(nil)
)
