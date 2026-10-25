# hse-scoot

## Назначение

Консольное приложение сервиса кикшеринга hse-scoot. Загружает инвентарь,
восстанавливает начальный транспорт, проводит техосмотр новых устройств
и ставит их на учёт. Считает вещи на учёте и выводит отчёт со списками
всей техники, транспорта и транспорта для новичков — в консоль или файл.

В инвентаре есть транспорт (самокаты, велосипеды, электровелосипеды)
и снаряжение (шлемы, наколенники, зарядные станции). У каждой вещи есть
инвентарный номер и категория. Новые виды добавляются в доменном слое.
Список приходит из текстового файла через порт `source.Source`, поэтому
источник можно заменить на JSON или БД без правок прикладного слоя.

## Стек и требования

- Go 1.23+.
- Только стандартная библиотека, внешних зависимостей нет.
- Запуск: `go run .`; тесты: `go test ./...`.

## Быстрый старт

```sh
# из корня проекта, отчёт в консоль
go run .

# отчёт в файл
go run . -report file -report-path report.txt

# другой файл инвентаря
go run . -devices data/devices.txt

# проверка
go test ./...
go vet ./...
```

## Структура

```text
.
├── main.go                       # флаги -devices, -report, -report-path и DI
├── data/
│   └── devices.txt               # транспорт и снаряжение
├── internal/
│   ├── domain/                   # контракты, иерархия вещей, фабрики
│   │   ├── consumptor.go         # IConsumptor
│   │   ├── inventoriable.go      # IInventoriable: номер + категория
│   │   ├── category.go           # ItemCategory и русские названия
│   │   ├── transport.go          # ITransport, TransportKind, NewTransport
│   │   ├── equipment.go          # EquipmentKind, Helmet, KneePads, ChargingStation
│   │   ├── device.go             # inventoryItem → transportDevice → electricDevice
│   │   ├── scooter.go            # Scooter
│   │   ├── bicycle.go            # Bicycle
│   │   └── ebike.go              # EBike
│   ├── inspection/               # техосмотр до постановки на учёт
│   ├── source/                   # порт Source, Dataset, DeviceSpec/EquipmentSpec
│   │   └── textfile/             # адаптер текстового файла
│   ├── fleet/                    # Sync, Items и подбор для новичков
│   │   ├── service.go            # сценарий инвентаря
│   │   └── beginner.go           # BeginnerDevices через ITransport/IConsumptor
│   ├── report/                   # счёт, документ отчёта и способы вывода
│   │   ├── report.go             # Counts, Document, Build, Count*
│   │   ├── text.go               # текстовое представление
│   │   └── output.go             # Output, ConsoleOutput, FileOutput
│   ├── console/                  # запуск сценария и служебный статус
│   └── di/                       # ручной DI-контейнер (composition root)
├── docs/
│   ├── architecture.md
│   └── decisions/
│       ├── 001-manual-di.md
│       ├── 002-transport-kinds.md
│       ├── 003-device-source-and-inspection.md
│       ├── 004-beginner-devices.md
│       ├── 005-itransport-and-inventory.md
│       └── 006-accounting-and-report.md
├── PROJECT.md
├── PROMPTS.md
└── ai-usage.md
```

## Документация

- [Архитектура](docs/architecture.md) — слои, SOLID, поток данных и расширение.
- [ADR 001](docs/decisions/001-manual-di.md) — ручной DI-контейнер без библиотек.
- [ADR 002](docs/decisions/002-transport-kinds.md) — виды транспорта и фабрика.
- [ADR 003](docs/decisions/003-device-source-and-inspection.md) — порт источника и техосмотр.
- [ADR 004](docs/decisions/004-beginner-devices.md) — лёгкость использования и новички (заменён ADR 005).
- [ADR 005](docs/decisions/005-itransport-and-inventory.md) — отдельный `ITransport`, инвентарь и иерархия вещей.
- [ADR 006](docs/decisions/006-accounting-and-report.md) — инвентарные номера, категории, счёт и вывод отчёта.
- [Журнал промтов](PROMPTS.md) — дословная история запросов.
- [AI usage](ai-usage.md) — дубликат журнала промтов по требованию заказчика.

## Статус и roadmap

- [x] Доменные интерфейсы `IConsumptor`, `IInventoriable`, `ITransport`.
- [x] Три вида транспорта: `Scooter`, `Bicycle`, `EBike`.
- [x] Транспорт одновременно реализует `ITransport` и `IInventoriable`.
- [x] Снаряжение: `Helmet`, `KneePads`, `ChargingStation` как `IInventoriable`.
- [x] Иерархия вещей: `inventoryItem` → `transportDevice`/`equipmentItem`.
- [x] Фабрики `domain.NewTransport` и `domain.NewEquipment`.
- [x] Инвентарный номер и категория у каждой вещи.
- [x] Техосмотр перед постановкой транспорта на учёт.
- [x] Порт `source.Source` и текстовый адаптер.
- [x] Начальный транспорт, очередь новых устройств и снаряжение в одном файле.
- [x] Лёгкость использования 1..10 и геттер в `ITransport`.
- [x] Подбор транспорта для новичков и расход через `IConsumptor`.
- [x] Счёт всех вещей, транспорта, самокатов и наколенников.
- [x] Отчёт со списками техники, транспорта и транспорта для новичков.
- [x] Вывод отчёта в консоль или файл через интерфейс `report.Output`.
- [x] Ручной DI-контейнер и тесты.
- [ ] Другие способы вывода отчёта (JSON, HTTP).
- [ ] JSON-адаптер `source.Source`.
- [ ] Хранение инвентаря (файл или БД) и загрузка конфигурации.
- [ ] Команды CLI (`fleet`, `device`) вместо одного отчёта.

## История изменений

| Дата | Что изменилось | Ссылка |
|---|---|---|
| 2026-10-25 | Создан каркас приложения: domain-интерфейсы, `Scooter`, сводка, консоль и DI | [ADR 001](docs/decisions/001-manual-di.md) |
| 2026-10-25 | Добавлены три вида транспорта, техосмотр и порт текстового источника | [ADR 002](docs/decisions/002-transport-kinds.md), [ADR 003](docs/decisions/003-device-source-and-inspection.md) |
| 2026-10-25 | Добавлена лёгкость использования и подбор устройств для новичков | [ADR 004](docs/decisions/004-beginner-devices.md) |
| 2026-10-25 | Выделен `ITransport`; добавлены снаряжение и иерархия вещей инвентаря | [ADR 005](docs/decisions/005-itransport-and-inventory.md) |
| 2026-10-25 | Добавлены инвентарные номера, категории, счёт и отчёт в консоль/файл | [ADR 006](docs/decisions/006-accounting-and-report.md) |
