# Order Service (тестовый микросервис)

Автор: разработчик проекта

---

## Описание

Это небольшой демонстрационный микросервис на Go, который получает данные о заказах из очереди (Kafka), сохраняет их в базу данных PostgreSQL и кэширует в памяти для быстрого доступа. Также сервис предоставляет HTTP API для получения информации о заказе по `order_uid` и простую web-страницу для поиска заказа по ID.

Цель проекта — показать взаимодействие компонентов: брокер сообщений (Kafka) → потребитель (Go) → БД (Postgres) + in-memory cache → HTTP API / frontend.

---

## Что внутри

* `cmd/main.go` — точка входа приложения (запускает HTTP-сервер, инициализирует хранилище, кеш и Kafka consumer).
* `internal/kafka/consumer.go` — логика подписки на тему Kafka, парсинга сообщений и их обработки.
* `internal/storage` — модуль работы с Postgres и кешем (в памяти).
* `internal/models` — структуры данных (Order, Delivery, Payment, Item).
* `configs/docker/init.sql` — SQL-скрипт создания таблиц и заполнения тестовыми данными.
* `configs/.env.template` — шаблон переменных окружения.
* `frontend/index.html` — простая страница для поиска заказа по `order_uid`.
* `docker-compose.yaml` — содержит сервисы для быстрого старта: Postgres, Kafka (и Zookeeper), сам сервис.

---

## Предварительные требования

* Docker & Docker Compose (рекомендуется для локальной развёртки)
* Go 1.25.0 (если вы планируете запускать сервис локально без Docker)

---

## Быстрый старт (рекомендованный — Docker Compose)

1. Скопируйте шаблон `.env.template` в `.env` и при необходимости скорректируйте переменные:

2. Запустите сервисы через `docker-compose` (в корне проекта рядом с `docker-compose.yaml`):

```bash
docker-compose up --build
```

`docker-compose` автоматически:

* запустит Postgres и применит `configs/docker/init.sql` (создание таблиц + тестовые данные),
* запустит Kafka (и Zookeeper),
* запустит контейнер с вашим Go-сервисом (или вы можете запускать сервис локально).

3. Откройте фронтенд в браузере (если frontend статическая страница сервируется через контейнер — путь `/frontend/index.html` или откройте файл `frontend/index.html` локально) и введите `order_uid` (например, `order_001`), либо используйте API напрямую.

---

## HTTP API

**GET** `/order/{order_uid}`

Возвращает JSON с данными заказа. Пример:

```bash
curl -v http://localhost:8080/order/order_001
```

Если заказ не найден в кеше — сервис попытается получить его из базы данных и вернёт соответствующий HTTP-код (404 если не найден).

---

## Web-frontend

Фронтенд — простая статическая страница `frontend/index.html`.
Она делает `fetch('/order/{order_uid}')` и отображает ответ в формате JSON.
Можете открыть файл локально в браузере или настроить Nginx/статическую раздачу для удобства.

---

## Kafka — отправка тестовых сообщений

Проект не содержит готового producer-а, но вы можете отправлять тестовые сообщения так:

1. Если используете Kafka из `docker-compose`, выполните в отдельном окне терминала (пример для образа `confluentinc/cp-kafka`):

```bash
# если у вас есть kafka-console-producer внутри контейнера
docker exec -it <kafka_container> bash
kafka-console-producer --broker-list localhost:9092 --topic orders
```

2. Отправьте JSON-сообщение (пример ниже). Обратите внимание, что структура должна соответствовать моделям в `internal/models`.

```json
{
   "order_uid": "b563feb7b2b84b6test",
   "track_number": "WBILMTESTTRACK",
   "entry": "WBIL",
   "delivery": {
      "name": "Test Testov",
      "phone": "+9720000000",
      "zip": "2639809",
      "city": "Kiryat Mozkin",
      "address": "Ploshad Mira 15",
      "region": "Kraiot",
      "email": "test@gmail.com"
   },
   "payment": {
      "transaction": "b563feb7b2b84b6test",
      "request_id": "",
      "currency": "USD",
      "provider": "wbpay",
      "amount": 1817,
      "payment_dt": 1637907727,
      "bank": "alpha",
      "delivery_cost": 1500,
      "goods_total": 317,
      "custom_fee": 0
   },
   "items": [
      {
         "chrt_id": 9934930,
         "track_number": "WBILMTESTTRACK",
         "price": 453,
         "rid": "ab4219087a764ae0btest",
         "name": "Mascaras",
         "sale": 30,
         "size": "0",
         "total_price": 317,
         "nm_id": 2389212,
         "brand": "Vivienne Sabo",
         "status": 202
      }
   ],
   "locale": "en",
   "internal_signature": "",
   "customer_id": "test",
   "delivery_service": "meest",
   "shardkey": "9",
   "sm_id": 99,
   "date_created": "2021-11-26T06:22:19Z",
   "oof_shard": "1"
}
```

После отправки consumer, подписанный на тему (topic) `orders` должен прочитать сообщение, распарсить JSON, сохранить его в Postgres и положить в кэш.

(Если используется другой `topic` — проверьте конфигурацию в коде/переменные окружения).

---

## Модель данных

Структура заказа описана в `internal/models/order.go`. Ключевые поля:

* `order_uid` — уникальный идентификатор заказа (string)
* `track_number`, `entry`
* `delivery` — объект Delivery (name, phone, zip, city, address, region, email)
* `payment` — объект Payment (transaction, request\_id, currency, amount и т.д.)
* `items` — массив товаров (Item: chrt\_id, price, name, nm\_id и др.)
* `locale`, `customer_id`, `delivery_service`, `date_created` и т.д.

SQL-структура таблиц находится в `configs/docker/1_init_tables.sql`.

---

## Кеширование

Сервис хранит последние полученные данные заказов в памяти (map). При старте сервис подтягивает актуальные данные из базы и заполняет кеш.

Кеш ускоряет повторные запросы по одному и тому же `order_uid` — данные отдаются из памяти без обращения к БД.

---

## Надёжность и обработка ошибок

* При получении сообщений из Kafka выполняется валидация/парсинг JSON; некорректные сообщения логируются и игнорируются.
* Сохранение в БД выполняется через транзакции (см. `internal/storage`), чтобы избежать частичного записанного состояния.
* Consumer использует механизмы подтверждения (commit) в Kafka (зависит от реализации `internal/kafka/consumer.go` и используемой библиотеки `sarama`).

---

## Полезные команды и примеры

* Собрать binary и запустить локально (но там нужно сильно заморочиться, честно говоря, локально даже не пробовал):

```bash
go build -o order-service ./cmd
./order-service
```

* Пример запроса API:

```bash
curl http://localhost:8080/order/order_001
```

* Если сервис в `docker-compose` проброшен на другой порт (например 8081), используйте соответствующий порт:

```bash
curl http://localhost:8081/order/order_001
```