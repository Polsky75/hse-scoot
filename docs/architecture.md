# Архитектура hse-scoot

## Обзор

Приложение — сценарий обслуживания инвентаря: загрузка списка вещей,
восстановление начального транспорта, добавление снаряжения, техосмотр
нового транспорта, постановка его на учёт, подсчёт вещей и вывод отчёта
в консоль или файл. Зависимости направлены внутрь: внешние слои знают
о домене, домен не знает о файлах и консоли.

```text
main
 └── internal/di                 composition root
      ├── internal/console        запуск сценария и служебный статус
      ├── internal/report         счёт, документ и способы вывода
      │    └── report.Output      ConsoleOutput / FileOutput
      └── internal/fleet          сценарий Sync и состояние инвентаря
           ├── internal/source    порт загрузки инвентаря
           │    └── textfile      адаптер текстового файла
           ├── internal/inspection
           └── internal/domain    контракты и иерархия вещей
```

| Пакет | Ответственность |
|---|---|
| `internal/domain` | Контракты `IConsumptor`, `IInventoriable`, `ITransport`; иерархия вещей; фабрики `NewTransport` и `NewEquipment`; инвентарные номера и категории. |
| `internal/inspection` | Техосмотр: оценка состояния и допуск транспорта к постановке на учёт. |
| `internal/source` | Порт `Source`, входные `DeviceSpec`, `EquipmentSpec` и `Dataset`. Не знает о файлах. |
| `internal/source/textfile` | Единственный адаптер, который открывает текстовый файл и валидирует формат. |
| `internal/fleet` | Прикладной сценарий `Sync`, хранение вещей, подбор транспорта для новичков. |
| `internal/report` | Счёт вещей, сборка `Document` и вывод отчёта через `Output`. |
| `internal/console` | Запуск сценария и печать служебного статуса регистрации. |
| `internal/di` | Создание и связывание конкретных реализаций. |
| `main` | Разбор флагов `-devices`, `-report`, `-report-path`, запуск контейнера. |

## Интерфейсы

```go
type IConsumptor interface {
	GetDailyConsumptionKWh() float64
}

type IInventoriable interface {
	GetInventoryNumber() string
	GetCategory() ItemCategory
}

type ITransport interface {
	IInventoriable
	GetEaseOfUse() int // лёгкость использования, 1..10
}
```

- Все вещи в инвентаре хранятся как `IInventoriable`.
- У каждой вещи есть инвентарный номер и категория.
- Весь транспорт одновременно реализует `ITransport` и `IInventoriable`.
- Потребители электроэнергии дополнительно реализуют `IConsumptor`;
  велосипед и снаряжение — нет.
- Шлемы, наколенники и зарядные станции реализуют только
  `IInventoriable`: у них нет ни лёгкости использования, ни расхода.

Соответствие проверяется на этапе компиляции через
`var _ ITransport = (*Scooter)(nil)` и аналогичные объявления.

## Иерархия вещей

```text
inventoryItem                  # inventoryNumber + category
├── transportDevice            # + easeOfUse, реализует ITransport
│   ├── electricDevice         # + dailyConsumptionKWh
│   │   ├── Scooter
│   │   └── EBike
│   └── Bicycle
└── equipmentItem              # только IInventoriable
    ├── Helmet
    ├── KneePads
    └── ChargingStation
```

Базовые типы неэкспортируемы: снаружи можно создать вещь только через
конструктор, который проверяет инварианты. Наследование в Go выражено
встраиванием, поэтому геттеры `GetInventoryNumber`, `GetCategory`,
`GetEaseOfUse` и `GetDailyConsumptionKWh` поднимаются по цепочке.

Виды задаются типами `TransportKind` и `EquipmentKind`, а категория —
типом `ItemCategory`:

```go
const (
	CategoryScooter         ItemCategory = "scooter"
	CategoryBicycle         ItemCategory = "bicycle"
	CategoryEBike           ItemCategory = "ebike"
	CategoryHelmet          ItemCategory = "helmet"
	CategoryKneePads        ItemCategory = "kneepads"
	CategoryChargingStation ItemCategory = "charging_station"
)
```

`ParseTransportKind`/`ParseEquipmentKind` валидируют строки из источника,
`NewTransport`/`NewEquipment` создают нужную сущность. Велосипед
с ненулевым расходом отклоняется (`ErrBicycleConsumption`). Лёгкость
использования проверяется в диапазоне 1..10 (`ErrInvalidEaseOfUse`),
инвентарный номер — на непустоту (`ErrEmptyInventoryNumber`).

## Транспорт для новичков

Функция `fleet.BeginnerDevices` отбирает только `ITransport` с
`GetEaseOfUse() >= 7` и считает расход отобранных устройств. Снаряжение
не реализует `ITransport`, поэтому в выборку не попадает. К потребителям
функция обращается только через `IConsumptor`:

```go
for _, item := range items {
	transport, ok := item.(domain.ITransport)
	if !ok || transport.GetEaseOfUse() < BeginnerMinEaseOfUse {
		continue
	}
	selected = append(selected, item)
	if consumer, ok := item.(domain.IConsumptor); ok {
		total += consumer.GetDailyConsumptionKWh()
	}
}
```

Метод `fleet.Service.BeginnerDevices` применяет функцию к текущему
инвентарю. Порог задан константой `BeginnerMinEaseOfUse = 7` (строго
больше 6).

## Счёт и отчёт

`internal/report` получает срез `[]domain.IInventoriable` и строит
`Document`:

- `Counts` — всего вещей, транспорт, самокаты, наколенники;
- `AllItems` — вся техника (транспорт и снаряжение);
- `Transport` — только `ITransport`;
- `BeginnerTransport` — транспорт, отобранный для новичков.

```go
func Build(items, beginnerItems []domain.IInventoriable) Document
func CountAll(items []domain.IInventoriable) int
func CountTransport(items []domain.IInventoriable) int
func CountCategory(items []domain.IInventoriable, category domain.ItemCategory) int
```

`RenderText` формирует текст, в котором каждая строка содержит
инвентарный номер и человекочитаемую категорию:

```text
Всего вещей на учёте: 8
Транспорт: 5
Самокаты: 2
Наколенники: 1

Вся техника (8):
  SCOOT-001 — самокат
  ...
```

Вывод отчёта отделён интерфейсом:

```go
type Output interface {
	Write(Document) error
}
```

`ConsoleOutput` пишет в `io.Writer`, `FileOutput` — в файл.
`report.NewOutput(target, path, console)` — фабрика способов вывода.
Новый способ (JSON, HTTP, почта) добавляется новой реализацией `Output`
и веткой фабрики; `Document`, счёт и домен не меняются.

## Поток данных

1. `main` выбирает `report.Output` по флагам и передаёт его в `di.New`.
2. Контейнер создаёт `textfile.Source`, `inspection.Inspector`,
   `fleet.Service` и `console.App`.
3. `console.App.Run` вызывает `fleet.Service.Sync`.
4. `Sync` загружает `source.Dataset` через порт `source.Source`.
5. Транспорт из `[registered]` восстанавливается как уже стоящий на учёте —
   повторный техосмотр для него не выполняется.
6. Снаряжение из `[equipment]` добавляется в инвентарь напрямую.
7. Для каждого устройства из `[incoming]`:
   - `domain.NewTransport` создаёт сущность;
   - `inspection.Inspector.Inspect` проверяет оценку состояния;
   - прошедшее устройство добавляется в инвентарь, отклонённое — только
     в результат с причиной.
8. `console.App` печатает служебный статус регистрации, затем строит
   `report.Document` и передаёт его в выбранный `Output`.

Техосмотр по умолчанию требует оценку не ниже 70 из 100. Порог задаётся
при сборке контейнера и может быть изменён без правок домена.

## SOLID

- **SRP** — сущности хранят инварианты, инспектор проверяет состояние,
  сервис управляет сценарием, `report` считает и форматирует, контейнер
  связывает.
- **OCP** — новый вид транспорта/снаряжения добавляется в домен, новый
  формат данных — реализацией `source.Source`, новый способ вывода —
  реализацией `report.Output`.
- **LSP** — любая реализация контрактов подставляется в сервис и отчёт
  без изменения их поведения.
- **ISP** — `ITransport` отделён от `IInventoriable`; снаряжение не обязано
  иметь лёгкость использования, а велосипед — расход энергии.
- **DIP** — `fleet.Service` зависит от порта `source.Source`, `console.App`
  — от `report.Output`; конкретные реализации создаются в `di`.

## Внедрение зависимостей

Используется ручной DI-контейнер (`internal/di`), он же composition root.
Контейнер собирает источник, инспектора, сервис, способ вывода и консоль.

```go
output, err := report.NewOutput("file", "report.txt", os.Stdout)
container, err := di.New(os.Stdout, "data/devices.txt", output)
if err := container.Run(); err != nil { /* ... */ }
```

`di.NewWithSource(out, src, minCondition, output)` — точка расширения:
принимает любой `source.Source`, порог техосмотра и любой `report.Output`.
JSON-адаптер источника или новый способ вывода подключается здесь одной
строкой.

## Источник данных

`source.Source` — порт с единственным методом `Load() (Dataset, error)`.
`Dataset` состоит из трёх списков: `Registered` (транспорт на балансе),
`Incoming` (транспорт на постановку) и `Equipment` (снаряжение).
Текстовый адаптер:

- читает секции `[registered]`, `[incoming]` и `[equipment]`;
- для транспорта проверяет вид, расход, оценку состояния и лёгкость 1..10;
- для снаряжения проверяет вид и инвентарный номер;
- отклоняет дубликаты инвентарных номеров между всеми секциями
  и неизвестные секции;
- не пропускает вещи без секции и битые строки.

Файл не читается нигде, кроме `internal/source/textfile`. Прикладной слой
работает с портом и не импортирует `os` для загрузки данных.

## Ошибки и вывод

- Ошибка загрузки источника или восстановления начального инвентаря
  прерывает сценарий: без корректного списка продолжать нельзя.
- Ошибка отдельного нового транспорта (невалидные данные или проваленный
  техосмотр) не прерывает обработку остальных: она попадает в
  `RegistrationResult`.
- Ошибки оборачиваются через `%w`; вызывающий код может проверить причину
  через `errors.Is`.
- `console.App` зависит от `report.Output` и `io.Writer`, поэтому в тестах
  используются буферы, а в production — `os.Stdout` и файл.

## Расширение

1. **Новый вид транспорта.** Добавить сущность в `internal/domain`,
   константу и ветки в `ParseTransportKind`/`NewTransport`. Сущность
   обязана реализовать `ITransport` (а значит, и `IInventoriable`);
   если она потребляет энергию — ещё и `IConsumptor`.
2. **Новый вид снаряжения.** Добавить сущность, константу и ветку в
   `ParseEquipmentKind`/`NewEquipment`; она реализует только
   `IInventoriable`. Адаптер источника расширяется новым значением kind.
3. **Новый способ вывода отчёта.** Реализовать `report.Output` и добавить
   ветку в `report.NewOutput`; счёт, `Document` и домен не меняются.
4. **JSON или БД вместо текстового файла.** Реализовать `source.Source`
   отдельным пакетом и передать его в `di.NewWithSource`; домен и сценарий
   останутся прежними.
5. **Другая политика техосмотра.** Расширить `inspection.Inspector`;
   контракт `Inspect` и место вызова в `fleet` сохранятся.
6. **Хранение инвентаря.** Когда появится реальное хранилище, добавить
   репозиторий и подключить его в `di`; сейчас состояние живёт в памяти
   на время одного запуска.

## Тестирование

- `internal/domain` — парсинг видов, фабрики, валидация инвентарных
  номеров и категорий, разделение `ITransport`/`IInventoriable`,
  отсутствие `IConsumptor` у велосипеда и снаряжения.
- `internal/inspection` — порог, отказ, невалидные оценки.
- `internal/source/textfile` — секции, комментарии, все ветки ошибок,
  дубликаты между транспортом и снаряжением.
- `internal/fleet` — восстановление инвентаря, регистрация, отказ,
  подбор для новичков, ошибки источника.
- `internal/report` — счёт, секции, текст, консольный и файловый вывод.
- `internal/console` — служебный статус и передача отчёта в Output.
- `internal/di` — полная сборка с текстовым файлом, файловым выводом
  и подмена источника.
