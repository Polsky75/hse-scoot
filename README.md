# hse-scoot

Консольное приложение сервиса кикшеринга: загружает инвентарь, проводит
техосмотр нового транспорта и ставит его на учёт, затем считает вещи
на учёте и выводит отчёт в консоль или файл.

В инвентаре есть транспорт (самокаты, велосипеды, электровелосипеды)
и снаряжение (шлемы, наколенники, зарядные станции). У каждой вещи есть
инвентарный номер и категория. Весь транспорт реализует одновременно
`ITransport` и `IInventoriable`, снаряжение — только `IInventoriable`.

## Запуск

```sh
# отчёт в консоль
go run .

# отчёт в файл
go run . -report file -report-path report.txt

# другой файл инвентаря
go run . -devices data/devices.txt
```

Флаги:

- `-devices` — путь к текстовому файлу инвентаря (по умолчанию
  `data/devices.txt`);
- `-report` — способ вывода: `console` или `file` (по умолчанию `console`);
- `-report-path` — путь к файлу при `-report=file`.

## Отчёт

Отчёт содержит счёт вещей и три списка:

- вся техника;
- весь транспорт;
- транспорт, подходящий новичкам.

У каждой строки указаны инвентарный номер и категория. Счёт включает
все вещи, транспорт, самокаты и наколенники.

```text
hse-scoot — отчёт по инвентарю

Всего вещей на учёте: 8
Транспорт: 5
Самокаты: 2
Наколенники: 1

Вся техника (8):
  SCOOT-001 — самокат
  BIKE-001 — велосипед
  EBIKE-001 — электровелосипед
  HELMET-001 — шлем
  KNEE-001 — наколенники
  STATION-001 — зарядная станция
  SCOOT-004 — самокат
  BIKE-002 — велосипед

Транспорт (5):
  ...

Транспорт для новичков (3):
  SCOOT-001 — самокат
  BIKE-001 — велосипед
  SCOOT-004 — самокат
```

Перед отчётом в консоль выводится служебный статус постановки на учёт,
включая отклонённые техосмотром устройства.

## Интерфейсы и иерархия

```go
type IInventoriable interface {
	GetInventoryNumber() string
	GetCategory() ItemCategory
}

type ITransport interface {
	IInventoriable
	GetEaseOfUse() int
}
```

Иерархия домена:

```text
inventoryItem                  # number + category
├── transportDevice            # + easeOfUse, ITransport
│   ├── electricDevice         # + dailyConsumptionKWh
│   │   ├── Scooter
│   │   └── EBike
│   └── Bicycle
└── equipmentItem
    ├── Helmet
    ├── KneePads
    └── ChargingStation
```

Все вещи хранятся в сервисе как `[]IInventoriable`. Транспорт
дополнительно приводится к `ITransport`; потребители энергии опрашиваются
через `IConsumptor`.

## Формат входного файла

Транспорт:

```text
kind inventory_number daily_consumption_kwh inspection_condition ease_of_use
```

- `kind` — `scooter`, `bicycle` или `ebike`;
- `inventory_number` — уникальный инвентарный номер;
- `daily_consumption_kwh` — суточный расход, неотрицательное число;
- `inspection_condition` — оценка техсостояния от 0 до 100;
- `ease_of_use` — лёгкость использования от 1 до 10.

Снаряжение:

```text
kind inventory_number
```

- `kind` — `helmet`, `kneepads` или `charging_station`;
- `inventory_number` — уникальный инвентарный номер.

Секция `[registered]` — транспорт, уже стоящий на балансе; оценка состояния
в ней указана для полноты формата и повторно не проверяется.
Секция `[incoming]` — транспорт, который проходит техосмотр и ставится
на учёт; минимальный проходной балл — 70.
Секция `[equipment]` — шлемы, наколенники и зарядные станции, которые
добавляются в инвентарь напрямую. Пустые строки и комментарии с `#`
игнорируются. Номера уникальны между всеми секциями. Пример —
[data/devices.txt](data/devices.txt).

Файл читается только адаптером `internal/source/textfile`; остальной код
зависит от интерфейса `source.Source`, поэтому формат можно заменить
(например, на JSON) без изменений домена и сценария.

## Способы вывода отчёта

Способ вывода описан интерфейсом:

```go
type Output interface {
	Write(Document) error
}
```

Сейчас есть две реализации:

- `ConsoleOutput` — печатает в `io.Writer`;
- `FileOutput` — сохраняет текстовый отчёт в файл.

Фабрика `report.NewOutput(target, path, console)` выбирает реализацию.
Новый способ (JSON, HTTP, почта) добавляется новой реализацией `Output`
и веткой в фабрике; сборка отчёта и домен при этом не меняются.

## Тесты

```sh
go test ./...
go vet ./...
```

Покрытие считается так:

```sh
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out | tail -1
```

Текущее покрытие — **88.9%** (требование: больше 60%).

## Техосмотр в тестах

Прикладной слой зависит от узкого порта `fleet.Inspector`, а не от
конкретного `inspection.Inspector`. В тестах `fleet` и `console`
используется `fakeInspector`:

- он записывает, какие инвентарные номера и оценки состояния получил;
- умеет вернуть управляемую ошибку для выбранного номера;
- позволяет проверить, что осматривается только `[incoming]`-транспорт,
  а начальный парк и снаряжение — нет.

Сама пороговая логика проверяется отдельными unit-тестами
`internal/inspection` на настоящих доменных объектах.

## Современный Go

Проект следует [JetBrains Modern Go Guidelines](https://github.com/JetBrains/go-modern-guidelines)
для возможностей, доступных в Go 1.23 из `go.mod`:

- `slices.Clone` вместо ручного копирования срезов;
- `strings.CutPrefix`/`strings.CutSuffix` вместо проверки и среза вручную;
- `errors.Is`, `any`, современная семантика замыканий в цикле;
- `run() error` в `main` вместо повторяющегося `os.Exit` в каждой ветке.

`go fix -diff ./...` не находит дополнительных автоматических модернизаций.

## Документация

- [PROJECT.md](PROJECT.md) — назначение, структура, статус и roadmap.
- [docs/architecture.md](docs/architecture.md) — слои, SOLID и DI-контейнер.
- [docs/decisions/](docs/decisions/) — архитектурные решения (ADR).
- [PROMPTS.md](PROMPTS.md) и [ai-usage.md](ai-usage.md) — журналы промтов.
