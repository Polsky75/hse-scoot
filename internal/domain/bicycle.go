package domain

// Bicycle — велосипед без электропривода.
//
// Числится на балансе, но не является потребителем электроэнергии.
type Bicycle struct {
	transportDevice
}

// NewBicycle создаёт велосипед и проверяет его инварианты.
func NewBicycle(inventoryNumber string, easeOfUse int) (*Bicycle, error) {
	device, err := newTransportDevice(inventoryNumber, CategoryBicycle, easeOfUse)
	if err != nil {
		return nil, err
	}
	return &Bicycle{transportDevice: device}, nil
}

var (
	_ ITransport     = (*Bicycle)(nil)
	_ IInventoriable = (*Bicycle)(nil)
)
