// Package fleet содержит прикладной сервис инвентаря и транспорта.
package fleet

import (
	"fmt"

	"github.com/Polsky75/hse-scoot/internal/domain"
	"github.com/Polsky75/hse-scoot/internal/inspection"
	"github.com/Polsky75/hse-scoot/internal/source"
)

// RegistrationResult — итог постановки одного транспорта на учёт.
type RegistrationResult struct {
	Kind            domain.TransportKind
	InventoryNumber string
	Err             error
}

// Accepted сообщает, прошло ли устройство техосмотр и было ли поставлено на учёт.
func (r RegistrationResult) Accepted() bool {
	return r.Err == nil
}

// Service загружает инвентарь, восстанавливает уже стоящий на учёте транспорт,
// добавляет снаряжение и регистрирует новые устройства.
//
// Новый транспорт проходит техосмотр до добавления в парк; начальный парк
// уже стоит на учёте и восстанавливается без повторного осмотра.
type Service struct {
	source    source.Source
	inspector *inspection.Inspector
	items     []domain.IInventoriable
	consumers []domain.IConsumptor
}

// NewService создаёт сервис инвентаря.
func NewService(src source.Source, inspector *inspection.Inspector) *Service {
	return &Service{
		source:    src,
		inspector: inspector,
	}
}

// Sync загружает инвентарь и выполняет постановку нового транспорта на учёт.
func (s *Service) Sync() ([]RegistrationResult, error) {
	dataset, err := s.source.Load()
	if err != nil {
		return nil, fmt.Errorf("load inventory: %w", err)
	}

	// Начальный парк уже на балансе, поэтому техосмотр для него не повторяется.
	for _, spec := range dataset.Registered {
		device, err := domain.NewTransport(
			spec.Kind,
			spec.InventoryNumber,
			spec.DailyConsumptionKWh,
			spec.EaseOfUse,
		)
		if err != nil {
			return nil, fmt.Errorf("restore registered transport %q: %w", spec.InventoryNumber, err)
		}
		s.add(device)
	}

	// Шлемы, наколенники и зарядные станции — вещи инвентаря, не транспорт.
	for _, spec := range dataset.Equipment {
		item, err := domain.NewEquipment(spec.Kind, spec.InventoryNumber)
		if err != nil {
			return nil, fmt.Errorf("restore equipment %q: %w", spec.InventoryNumber, err)
		}
		s.add(item)
	}

	results := make([]RegistrationResult, 0, len(dataset.Incoming))
	for _, spec := range dataset.Incoming {
		results = append(results, s.register(spec))
	}
	return results, nil
}

// Items возвращает копию списка вещей на учёте.
func (s *Service) Items() []domain.IInventoriable {
	items := make([]domain.IInventoriable, len(s.items))
	copy(items, s.items)
	return items
}

func (s *Service) register(spec source.DeviceSpec) RegistrationResult {
	result := RegistrationResult{
		Kind:            spec.Kind,
		InventoryNumber: spec.InventoryNumber,
	}

	device, err := domain.NewTransport(
		spec.Kind,
		spec.InventoryNumber,
		spec.DailyConsumptionKWh,
		spec.EaseOfUse,
	)
	if err != nil {
		result.Err = err
		return result
	}
	if err := s.inspector.Inspect(device, spec.Condition); err != nil {
		result.Err = err
		return result
	}

	s.add(device)
	return result
}

func (s *Service) add(item domain.IInventoriable) {
	s.items = append(s.items, item)
	if consumer, ok := item.(domain.IConsumptor); ok {
		s.consumers = append(s.consumers, consumer)
	}
}
