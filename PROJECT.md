# hse-scoot

## Назначение

Консольное приложение сервиса кикшеринга hse-scoot. Загружает инвентарь,
восстанавливает начальный транспорт, проводит техосмотр новых устройств
и ставит их на учёт. Считает вещи и суммарное суточное потребление
энергии, выводит отчёт со списками всего оборудования, транспорта
и транспорта для новичков — в консоль или файл. Данные можно загрузить
из файла или ввести интерактивно без файла.

В инвентаре есть транспорт (самокаты, велосипеды, электровелосипеды)
и снаряжение (шлемы, наколенники, зарядные станции). Зарядная станция
может потреблять энергию и учитывается в суммарном расходе. У каждой
вещи есть инвентарный номер и категория. Новые виды добавляются в доменном слое.
Список приходит из текстового файла через порт `source.Source`, поэтому
источник можно заменить на JSON или БД без правок прикладного слоя.

## Стек и требования

- Go 1.23+.
- Только стандартная библиотека, внешних зависимостей нет.
- Современные идиомы Go по [JetBrains Modern Go Guidelines](https://github.com/JetBrains/go-modern-guidelines).
- Покрытие тестами больше 60% (текущее — 86.7%).
- Запуск: `go run .`; тесты: `go test ./...`.

## Быстрый старт

```sh
# из корня проекта, файловый режим
go run .

# интерактивный ввод без файла
go run . -interactive

# отчёт в файл
go run . -report file -report-path report.txt

# другой файл инвентаря
go run . -devices data/devices.txt

# проверка
go test ./...
go vet ./...

# подсчёт покрытия
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out | tail -1
```

## Структура

```text
.
├── main.go                       # флаги -devices, -report, -report-path, -interactive и DI
├── data/
│   └── devices.txt               # транспорт и снаряжение
├── internal/
│   ├── domain/                   # контракты, иерархия вещей, фабрики
│   │   ├── consumptor.go         # IConsumptor
│   │   ├── inventoriable.go      # IInventoriable: номер + категория
│   │   ├── category.go           # ItemCategory и русские названия
│   │   ├── transport.go          # ITransport, TransportKind, NewTransport
│   │   ├── equipment.go          # EquipmentKind, Helmet, KneePads, ChargingStation (IConsumptor)
│   │   ├── device.go             # inventoryItem → transportDevice → electricDevice
│   │   ├── scooter.go            # Scooter
│   │   ├── bicycle.go            # Bicycle
│   │   └── ebike.go              # EBike
│   ├── inspection/               # техосмотр до постановки на учёт
│   ├── source/                   # порт Source, Dataset, DeviceSpec/EquipmentSpec
│   │   └── textfile/             # адаптер текстового файла
│   ├── fleet/                    # Sync, Items и подбор для новичков
│   │   ├── service.go            # сценарий инвентаря и порт Inspector
│   │   └── beginner.go           # BeginnerDevices через ITransport/IConsumptor
│   ├── report/                   # счёт, документ отчёта и способы вывода
│   │   ├── report.go             # Counts, Document, Build, Count*
│   │   ├── text.go               # текстовое представление
│   │   └── output.go             # Output, ConsoleOutput, FileOutput
│   ├── console/                  # запуск файлового сценария и статус
│   ├── interactive/              # раздельный ввод транспорта/оборудования и номера
│   └── di/                       # ручной DI-контейнер (composition root)
├── PROJECT.md
└── ai-usage.md
```
