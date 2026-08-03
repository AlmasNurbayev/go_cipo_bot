# CLAUDE.md

Хранилище данных продаж / остатков / финансов магазина Cipo. Пользовательский интерфейс — Telegram-бот. Язык кода и комментариев — русский, тексты бота — русские.

## Язык
Отвечай на русском языке

## Команды

```bash
make bot            # go run cmd/bot/main.go -configEnv ./.env
make updater1       # updater за последний 1 день (есть updater10, updater32, updater2023-01)
make migrate_up     # миграции на локальную БД (порт 5911)
make lint           # golangci-lint run ./cmd/... ./internal/...
go build ./...      # сборка всех бинарников
```

`make test` объявлен в makefile, но каталога `tests/` в репозитории нет — тестов сейчас не существует.

Prod-запуск: `docker compose up -d` (образ `almasnurbayev/go_cipo_bot:latest`). CI (`.github/workflows/docker-publish.yml`) собирает и пушит образ **только по тегу `v*`**, предварительно прогоняя golangci-lint. Контейнер при старте выполняет `MIGRATOR -typeTask up` и затем `BOT`. KOFD_UPDATER в контейнере не запускается сам — его дёргает внешний cron каждые 2 минуты.

## Архитектура

Три бинарника из одного модуля (`cmd/`):

| Бинарник | Назначение | Жизненный цикл |
|---|---|---|
| `bot` | Telegram-бот + HTTP `/healthz` + NATS consumer | долгоживущий |
| `kofd_updater` | тянет чеки из API КОФД → Postgres → публикует новые операции в NATS | одноразовый прогон по cron |
| `migrator` | golang-migrate up/down | одноразовый |

Поток данных:

```
KOFD API ──┐
CIPO backend ──> kofd_updater ──> Postgres ──> BOT ──> Telegram
                      │                         ↑
                      └──> NATS (JetStream) ────┘   subject: new_transactions
Google Sheets ────────────────────────────────> BOT (модуль finance)
```

Внешние источники: недокументированный REST API КОФД (чеки), бэкенд сайта cipo.kz (карточки товаров, картинки, остатки — приходят из 1С Розница), Google Sheets API (расходы/себестоимость для ОПиУ).

## Структура

```
cmd/{bot,kofd_updater,migrator}/main.go
internal/
  botP/                 приложение бота
    bot.go              BotApp: сборка bot.Bot, middleware, Run/Stop
    http.go             HttpApp: только /healthz для docker healthcheck
    nats.go             RunNatsConsumer — pull-подписка, рассылка уведомлений
    _kafka.go           МЁРТВЫЙ КОД: файлы с `_` Go не компилирует, Kafka заменена на NATS
    middleware/         Recover, CheckUser
    api/                HTTP-клиенты, нужные боту (gsheets, остатки Cipo)
    summary/ charts/ qnt/ finance/ other/     фичевые модули
  kofd_updater/
    api/                клиенты KOFD (token, operations, check) и Cipo (product)
    kofd_updater_services/   бизнес-логика прогона
  storage/postgres/     Storage + методы, sql в строках, scany для скана
  models/               DB-сущности и DTO
  config/               Config из env + чтение app_settings из БД
  lib/logger/           slog + tint, splitHandler
  lib/utils/            даты, числа, парсинг чеков, крипта, форматирование сообщений
migrations/             NNN_name.{up,down}.sql, golang-migrate
```

## Ключевые паттерны

**Фичевый модуль бота = тройка `init.go` / `handlers.go` / `services.go`.** `Init(b, storage, log, cfg)` регистрирует хендлеры и определяет `initKeyboard` (ReplyKeyboard по команде `/summary`, `/charts`, `/qnt`, `/finance`, `/other`). `handlers.go` — только Telegram-специфика и форматирование текста. `services.go` — получение и агрегация данных. Все модули регистрируются списком в `BotApp.Run()` (`internal/botP/bot.go:63`).

**Хендлеры — замыкания-фабрики.** Зависимости не глобальные, а захваченные:

```go
func summaryHandler(storage storageI, log1 *slog.Logger, cfg *config.Config) bot.HandlerFunc {
    return func(ctx context.Context, b *bot.Bot, update *models.Update) { ... }
}
```

**Роутинг.** Текстовые команды ловятся регекспом по префиксу-глаголу (`^итоги.*`, `^график.*`, `^финансы.*`), inline-кнопки — `MatchTypePrefix` по `callback_data` вида `summary_`, `getCheck_<id>`, `getFullTextCheck_<id>`. Кнопки клавиатуры шлют обычный текст («итоги пр. месяц»), который разбирает `utils.GetPeriodByMode`.

**Локальные интерфейсы у потребителя.** Каждый пакет объявляет свой минимальный `storageI` (или `storageOperations`) с нужными ему методами, а `*storage.Storage` подставляется как реализация. Новый запрос к БД → метод на `*Storage` + строка в локальном интерфейсе того пакета, который его вызывает. Часть модулей (charts, finance) для простоты принимает `*storage.Storage` напрямую — расширять интерфейс предпочтительнее.

**Storage знает про внешнюю транзакцию.** У `Storage` есть публичное поле `Tx *pgx.Tx`; каждый метод выбирает: если `s.Tx != nil` — работать в транзакции, иначе через пул. Транзакцию открывает вызывающий (см. `cmd/kofd_updater/main.go`), сам storage её не коммитит. Запросы — сырой SQL в строках, скан через `pgxscan.Select/Get`, `sql.ErrNoRows` трактуется как пустой результат, а не ошибка.

**Логирование.** В начале каждой функции: `op := "пакет.Функция"; log := log1.With(slog.String("op", op))`. Параметр называется `log1`, чтобы затенённый `log` был обогащённым. Ошибки логируются на месте **и** возвращаются наверх. `logger.InitLogger` даёт splitHandler: всё в stdout (tint в dev, JSON в prod), а `Error` и выше дублируется в JSON-файл `LOG_ERROR_PATH`.

**Config.** `config.Mustload(path)` — godotenv (если передан `-configEnv`) + `caarlos0/env`. Секреты помечены `json:"-"`, потому что конфиг печатается в лог при старте. Изменяемые бизнес-настройки лежат не в env, а в таблице `app_settings` (JSON-значения) и читаются хелперами `config.GetSettingsString/Float64/USDRates/GSheetsSources` — так задаются статьи ОПиУ, курсы доллара, ссылки на Google Sheets.

**Модели.** `guregu/null/v5` для nullable-колонок, теги `json` + `db`, поля в стиле `Sum_operation`, `Kassa_id`. `ChequeJSONList` реализует `sql.Scanner` для JSONB-колонки с распарсенным составом чека.

**HTTP-клиенты внешних API.** Один файл — одна функция: `url.Parse` → `http.NewRequest` → `client.Do` → проверка `StatusCode` → `io.ReadAll` → `json.Unmarshal`, ошибки логируются и возвращаются. Таймаут выставлен не везде: `kofdGetCheck.go` и `cipoGetProduct.go` — 5 секунд, остальные создают `&http.Client{}` без таймаута; при правках стоит добавлять таймаут.

**Параллелизм.** Только в `GetOperationsFromApi`: `errgroup.Group` с `SetLimit(10)` тянет чеки, каждая горутина пишет исключительно в свой индекс `listEntity[index]` — мьютексов нет, это условие нужно сохранять. В `cmd/bot/main.go` три горутины (bot, http, nats consumer) и остановка по `signal.NotifyContext`.

## Особенности, о которых стоит помнить

- `middleware.CheckUser` пропускает апдейт дальше без проверки, если `update.Message == nil`, — то есть **callback-запросы от inline-кнопок не авторизуются**.
- `middleware.Recover` логирует панику и делает `os.Exit(1)`: процесс падает, docker поднимает заново.
- Роль `kaspi_manager` (миграция 011) — пользователь получает только NATS-уведомления о продажах товаров с `kaspi_in_sale`, все команды бота ему заблокированы в CheckUser.
- Курсор рассылки — колонка `users.transaction_cursor`; `DetectNewOperations` сдвигает его всегда, даже если сообщение пользователю по фильтру роли не отправляется.
- Порт Postgres в `.env` задаётся как `127.0.0.1:5911`, чтобы наружу он не торчал — доступ только через SSH-туннель.
- NATS JetStream ограничен 50 Мб памяти и 50 Мб диска (`nats_conf/nats-server.conf`); стрим создаётся consumer'ом на лету, если его нет.
- Telegram режет сообщения на 4096 символов — длинные списки чеков режутся на части (см. `summary`), максимум 50 чеков.
- Графики — `go-analyze/charts` (форк go-chart), рендер в PNG-байты, отправка как фото.
- Список задач и история сделанного ведутся в `readme.md` в разделе TODO с датами и пометками `[v]`.
