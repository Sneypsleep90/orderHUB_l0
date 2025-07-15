# 📦 orderHub_L0

**orderHub_L0** — это микросервис на Go, предназначенный для обработки заказов, поступающих из Kafka, их сохранения в PostgreSQL и предоставления через REST API. Приложение использует in-memory кэш для ускорения доступа и логирование через zap.

---

## 🚀 Функциональность

- 🔄 Получение заказов из Kafka (топик `orders`)
- 💾 Сохранение заказов в PostgreSQL (`orders`, `deliveries`, `payments`, `items`)
- 🧠 Кэширование заказов в памяти
- 🌐 HTTP API для получения заказа по `order_uid`

---

## 📁 Структура проекта

```bash
cmd/server/main.go             # Точка входа приложения
internal/model/model.go        # Структуры данных: Order, Delivery, Payment, Item
internal/db/db.go              # Подключение к PostgreSQL через sqlx
internal/db/migrations/...     # SQL-миграции
internal/service/service.go    # Загрузка заказов в кэш, бизнес-логика
internal/consumer/consumer.go  # Kafka-консьюмер
internal/cache/cache.go        # In-memory кэш
internal/handler/handler.go    # HTTP-обработчики
.env                           # Переменные окружения
Dockerfile                     # Сборка приложения
docker-compose.yml             # Инфраструктура: PostgreSQL, Kafka, приложение
go.mod / go.sum                # Зависимости Go
```

---

## ⚙️ Требования

- Go 1.23.4
- Docker Desktop (с поддержкой WSL2 на Windows)
- PostgreSQL (в Docker)
- Kafka (в Docker)
- `curl` — для тестирования API
- `psql` — для работы с базой
- `goose` — миграции (`v3.22.1`)

---

## 🛠 Установка и запуск

### 1. Клонируйте репозиторий:

```bash
git clone <URL_репозитория>
cd orderHub_L0
```

### 2. Установите зависимости:

```bash
go mod tidy
```

### 3. Запустите сервисы:

```bash
docker-compose up -d
```

Проверьте контейнеры:

```bash
docker ps
```

### 4. Примените миграции:

```bash
docker run -v "$(pwd)/internal/db/migrations:/migrations"   --network host   -e POSTGRES_URL=postgres://order_user:securepassword@localhost:5432/orders_db?sslmode=disable   pressly/goose:3.22.1 postgres up
```

---

## 🧪 Тестирование

### 1. Отправка тестового заказа в Kafka

Создайте файл `order1001test.json` со следующим содержанием:

<details>
<summary>📄 order1001test.json</summary>

```json
{
  "order_uid": "order1001test",
  "track_number": "WBILMTESTTRACK",
  "entry": "WBIL",
  "delivery": {
    "name": "Alexey Petrov",
    "phone": "+79161234567",
    "zip": "101000",
    "city": "Saint Petersburg",
    "address": "Nevsky Prospekt 20",
    "region": "Leningrad",
    "email": "alexey.petrov@example.com"
  },
  "payment": {
    "transaction": "order1001test",
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
  "customer_id": "test",
  "delivery_service": "meest",
  "shardkey": "9",
  "sm_id": 99,
  "date_created": "2025-07-12T15:00:00Z",
  "oof_shard": "1"
}
```

</details>

Загрузите и отправьте заказ:

```bash
docker cp order1001test.json orderhub_l0_kafka_1:/tmp/order1001test.json

docker exec -it orderhub_l0_kafka_1 bash

cat /tmp/order1001test.json | kafka-console-producer.sh --broker-list kafka:9092 --topic orders

exit
```

Проверьте логи:

```bash
docker logs orderhub_l0_app_1
```

Ожидаемый лог:

```json
{"level":"info","msg":"Order processed and cached","order_uid":"order1001test"}
```

---

### 2. Проверка REST API

```bash
curl http://localhost:8081/order/order1001test
```

Ответ: JSON с заказом.

---

### 3. Проверка в базе данных

```bash
psql -h localhost -U order_user -d orders_db
```

Пароль: `securepassword`

SQL-запросы:

```sql
SELECT * FROM orders WHERE order_uid = 'order1001test';
SELECT * FROM deliveries WHERE order_uid = 'order1001test';
SELECT * FROM payments WHERE order_uid = 'order1001test';
SELECT * FROM items WHERE order_uid = 'order1001test';
```

---

## 🧹 Удаление тестового заказа

```sql
DELETE FROM items WHERE order_uid = 'order1001test';
DELETE FROM payments WHERE order_uid = 'order1001test';
DELETE FROM deliveries WHERE order_uid = 'order1001test';
DELETE FROM orders WHERE order_uid = 'order1001test';
```

---

## ❗ Возможные проблемы

### 🔒 Порт 5432 занят

Найдите процесс:

```bash
netstat -ano | findstr :5432
```

Завершите:

```bash
taskkill /PID <PID> /F
```

### 📡 Ошибки Kafka

Проверьте наличие топика:

```bash
docker exec -it orderhub_l0_kafka_1 kafka-topics.sh --bootstrap-server kafka:9092 --list
```

Создайте топик, если его нет:

```bash
kafka-topics.sh --bootstrap-server kafka:9092   --create --topic orders --partitions 1 --replication-factor 1
```

---

## 🧑‍💻 Разработка в GoLand

1. Откройте проект в GoLand.
2. Установите переменную окружения: `GO_DOTENV_PATH=.env` в `Run > Edit Configurations`.
3. Запускайте `cmd/server/main.go` напрямую.

---

## 📬 Обратная связь

Если вы нашли баг или хотите предложить улучшение — создайте issue или pull request в репозитории.

---
