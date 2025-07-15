# orderHub_L0

**orderHub_L0** — это микросервис на Go для обработки заказов, поступающих из Kafka, их сохранения в PostgreSQL и предоставления через REST API. Приложение использует кэш в памяти для быстрого доступа к заказам и логирование с помощью zap.

---

## 🔧 Функциональность

- Получение заказов из Kafka (топик `orders`)
- Сохранение заказов в PostgreSQL в таблицы `orders`, `deliveries`, `payments`, `items`
- Кэширование заказов в памяти
- HTTP API для получения заказа по `order_uid`

---

## 🗂 Структура проекта

```plaintext
cmd/server/main.go                # Точка входа приложения
internal/model/model.go           # Структуры данных (Order, Delivery, Payment, Item)
internal/db/db.go                 # Подключение к PostgreSQL через sqlx
internal/db/migrations/...        # Миграции для создания таблиц
internal/service/service.go       # Загрузка заказов в кэш, обработка логики
internal/consumer/consumer.go     # Kafka-консьюмер и обработка заказов
internal/cache/cache.go           # In-memory кэш
internal/handler/handler.go       # HTTP обработчики
.env                              # Переменные окружения
docker-compose.yml                # Запуск PostgreSQL, Kafka и приложения
Dockerfile                        # Сборка приложения
go.mod / go.sum                   # Зависимости Go
📦 Требования
Go 1.23.4

Docker Desktop с WSL2 (для Windows)

PostgreSQL (через Docker)

Kafka (через Docker)

curl — для тестов API

psql — для работы с базой

goose — для миграций (v3.22.1)

🚀 Установка и запуск
1. Клонировать репозиторий
bash
Копировать
Редактировать
git clone <URL_репозитория>
cd orderHub_L0
2. Установить зависимости
bash
Копировать
Редактировать
go mod tidy
3. Запустить сервисы через Docker
bash
Копировать
Редактировать
docker-compose up -d
Проверьте, что контейнеры работают:

bash
Копировать
Редактировать
docker ps
4. Применить миграции
bash
Копировать
Редактировать
docker run -v "$(pwd)/internal/db/migrations:/migrations" --network host \
  -e POSTGRES_URL=postgres://order_user:securepassword@localhost:5432/orders_db?sslmode=disable \
  pressly/goose:3.22.1 postgres up
🧪 Тестирование
1. Отправка тестового заказа в Kafka
Создайте файл order1001test.json со следующим содержимым:

json
Копировать
Редактировать
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
Загрузите файл в контейнер Kafka и отправьте сообщение:

bash
Копировать
Редактировать
docker cp order1001test.json orderhub_l0_kafka_1:/tmp/order1001test.json

docker exec -it orderhub_l0_kafka_1 bash
cat /tmp/order1001test.json | kafka-console-producer.sh --broker-list kafka:9092 --topic orders
exit
Проверьте логи:

bash
Копировать
Редактировать
docker logs orderhub_l0_app_1
Ожидаемый вывод:

json
Копировать
Редактировать
{"level":"info","msg":"Order processed and cached","order_uid":"order1001test"}
2. Проверка REST API
bash
Копировать
Редактировать
curl http://localhost:8081/order/order1001test
Ожидается JSON с заказом.

3. Проверка в базе данных
bash
Копировать
Редактировать
psql -h localhost -U order_user -d orders_db
sql
Копировать
Редактировать
SELECT * FROM orders WHERE order_uid = 'order1001test';
SELECT * FROM deliveries WHERE order_uid = 'order1001test';
SELECT * FROM payments WHERE order_uid = 'order1001test';
SELECT * FROM items WHERE order_uid = 'order1001test';
Пароль пользователя: securepassword

⚙️ Проблемы и их решения
Порты заняты
Проверьте и завершите процессы:

bash
Копировать
Редактировать
netstat -ano | findstr :5432
taskkill /PID <PID> /F
Ошибки Kafka
Убедитесь, что топик orders существует:

bash
Копировать
Редактировать
docker exec -it orderhub_l0_kafka_1 bash
kafka-topics.sh --bootstrap-server kafka:9092 --list
Создайте топик, если он отсутствует:

bash
Копировать
Редактировать
kafka-topics.sh --bootstrap-server kafka:9092 --create --topic orders --partitions 1 --replication-factor 1
Повторная очистка БД (удалить тестовый заказ)
sql
Копировать
Редактировать
DELETE FROM items WHERE order_uid = 'order1001test';
DELETE FROM payments WHERE order_uid = 'order1001test';
DELETE FROM deliveries WHERE order_uid = 'order1001test';
DELETE FROM orders WHERE order_uid = 'order1001test';
🧑‍💻 Разработка с GoLand
Открой проект в GoLand.

Установи переменную окружения GO_DOTENV_PATH=.env в Run > Edit Configurations.

Запускай cmd/server/main.go напрямую.

📹 Демонстрация
Для финальной проверки:

Показать структуру проекта

Запустить Docker-сервисы

Применить миграции

Отправить JSON в Kafka

Проверить логи и API

Проверить запись в базе данных

