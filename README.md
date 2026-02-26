# Warehouse Management System (WMS) на Go + PostgreSQL

Полнофункциональная система учёта товаров на складе с использованием QR-кодов, REST API и аудит-лога всех перемещений.

## 🏗️ Архитектура

```
📱 Браузер (сканер QR)
        ↓
   Go REST API (порт 8081)
        ↓
    PostgreSQL (порт 5434)
```

## 📦 Что реализовано (ЭТАП 1-2)

### ✅ ЭТАП 2: База данных (PostgreSQL)
- Таблица `locations` — локации/полки на складе
- Таблица `items` — товары с текущей локацией
- Таблица `users` — операторы/администраторы
- Таблица `item_movements` — полный аудит-лог всех перемещений товаров

### ✅ ЭТАП 1: REST API (Gin Framework)
Полностью функциональный API с 5 endpoints:

```
POST   /api/login         — Вход (username/password)
GET    /api/item/:id      — Информация о товаре + текущая локация
GET    /api/item/:id/history — История всех перемещений товара
POST   /api/move          — Переместить товар на новую локацию
GET    /health            — Проверка статуса сервера
```

## 🚀 Быстрый старт

### 1. Подготовка окружения

```bash
# Клонируем/открываем проект
cd n:\Qr\QR-genrator

# Копируем .env файл
copy .env.example .env
```

### 2. Поднимаем PostgreSQL в Docker

```bash
# Запускаем контейнеры
docker-compose up -d

# Проверяем, что БД готова (ожидается 2-5 секунд)
docker ps | findstr warehouse
```

### 3. Запускаем API сервер

```bash
# Копилируем
go build -o warehouse.exe

# Или запускаем напрямую
go run main.go
```

Сервер запустится на **http://localhost:8081**

### 4. Заполняем БД тестовыми данными (один раз)

```bash
go run main.go --seed
```

Это создаст:
- 3 локации (LOC-A1, LOC-A2, LOC-B1)
- 3 товара (Widget Pro, Gadget Plus, Component X)
- 1 пользователя (operator1 / password123)
- 6 QR кодов в папке `qrcodes/`

## 📚 API Документация

### 1. POST /api/login — Авторизация

**Запрос:**
```json
{
  "username": "operator1",
  "password": "password123"
}
```

**Ответ (200):**
```json
{
  "success": true,
  "message": "Успешная авторизация",
  "user_id": "user1",
  "username": "operator1",
  "role": "operator",
  "token": "bearer_user1"
}
```

---

### 2. GET /api/item/:id — Получить товар

**Пример:**
```
GET http://localhost:8081/api/item/item1
```

**Ответ (200):**
```json
{
  "success": true,
  "item": {
    "id": "item1",
    "name": "Widget Pro",
    "sku": "WDGT-001",
    "description": "High-performance widget",
    "quantity": 50,
    "part_number": "PN-2024-001",
    "batch_number": "BATCH-2024-01",
    "location_id": "location2",
    "location": {
      "id": "location2",
      "code": "LOC-A2",
      "description": "Shelf A - Row 2"
    },
    "created_at": "2026-02-25T18:22:04.542364+03:00"
  }
}
```

---

### 3. GET /api/item/:id/history — История перемещений

**Пример:**
```
GET http://localhost:8081/api/item/item1/history
```

**Ответ (200):**
```json
{
  "success": true,
  "item_id": "item1",
  "total": 1,
  "movements": [
    {
      "id": 1,
      "item_id": "item1",
      "from_location_id": "location1",
      "from_location": { "id": "location1", "code": "LOC-A1" },
      "to_location_id": "location2",
      "to_location": { "id": "location2", "code": "LOC-A2" },
      "user_id": "user1",
      "user": { "id": "user1", "username": "operator1" },
      "notes": "Перемещение товара на полку A2",
      "moved_at": "2026-02-25T18:33:44.214845+03:00"
    }
  ]
}
```

---

### 4. POST /api/move — Переместить товар

**Запрос:**
```json
{
  "item_id": "item1",
  "to_location_id": "location3",
  "user_id": "user1",
  "notes": "Перемещение на склад B"
}
```

**Ответ (200):**
```json
{
  "success": true,
  "message": "Товар успешно перемещён",
  "movement": {
    "id": 2,
    "item_id": "item1",
    "from_location_id": "location2",
    "to_location_id": "location3",
    "user_id": "user1",
    "notes": "Перемещение на склад B",
    "moved_at": "2026-02-25T18:35:10.123456+03:00"
  }
}
```

---

### 5. GET /health — Проверка статуса

**Пример:**
```
GET http://localhost:8081/health
```

**Ответ (200):**
```json
{
  "status": "ok",
  "message": "🚀 Warehouse API is running"
}
```

## 📊 Схема БД

### Таблица: locations
```
id (PK)          - уникальный ID локации
code (UNIQUE)    - код (LOC-A1, LOC-A2 и т.д.)
description      - описание
row              - ряд (A, B, C...)
section          - секция (1, 2, 3...)
shelf            - полка (1, 2, 3...)
created_at       - дата создания
updated_at       - дата обновления
```

### Таблица: items
```
id (PK)          - уникальный ID товара
name             - наименование товара
sku (UNIQUE)     - артикул товара
description      - описание
quantity         - количество на складе
part_number      - номер детали
batch_number     - номер партии
location_id (FK) - текущая локация
created_at       - дата создания
updated_at       - дата обновления
```

### Таблица: users
```
id (PK)          - уникальный ID пользователя
username (UNIQUE) - имя пользователя
email (UNIQUE)   - электронная почта
password_hash    - хеш пароля (SHA256)
role             - роль (admin, operator)
created_at       - дата создания
updated_at       - дата обновления
```

### Таблица: item_movements
```
id (PK)          - уникальный ID движения
item_id (FK)     - ID товара
from_location_id (FK) - откуда переместили
to_location_id (FK) - куда переместили
user_id (FK)     - кто переместил
notes            - примечания
moved_at         - время движения
created_at       - время записи в БД
```

## 🔧 Команды управления

### Запуск только с API сервером (по умолчанию)
```bash
go run main.go
```

### Заполнение БД тестовыми данными
```bash
go run main.go --seed
```

### Генерирование QR кодов
```bash
go run main.go --genqr
```

### Полная инициализация (seed + QR + API)
```bash
go run main.go --seed --genqr
```

## 🐳 Docker команды

```bash
# Запуск контейнеров
docker-compose up -d

# Остановка контейнеров
docker-compose down

# Просмотр логов БД
docker-compose logs -f postgres

# Доступ к консоли PostgreSQL
docker exec -it warehouse_db psql -U warehouse -d warehouse_db

# pgAdmin доступен по адресу: http://localhost:5050
# Email: admin@warehouse.local
# Password: admin123
```

## 📱 SQL запросы для тестирования

```sql
-- Просмотр всех товаров с локациями
SELECT i.id, i.name, i.sku, i.quantity, l.code 
FROM items i 
LEFT JOIN locations l ON i.location_id = l.id;

-- История перемещений конкретного товара
SELECT m.*, l1.code as from_loc, l2.code as to_loc, u.username
FROM item_movements m
LEFT JOIN locations l1 ON m.from_location_id = l1.id
LEFT JOIN locations l2 ON m.to_location_id = l2.id
LEFT JOIN users u ON m.user_id = u.id
WHERE m.item_id = 'item1'
ORDER BY m.moved_at DESC;

-- Количество товаров на каждой локации
SELECT l.code, COUNT(i.id) as item_count, SUM(i.quantity) as total_qty
FROM locations l
LEFT JOIN items i ON l.id = i.location_id
GROUP BY l.id, l.code;
```

## ⚙️ Конфигурация (.env)

```
# Database Configuration
DATABASE_HOST=localhost
DATABASE_PORT=5434
DATABASE_USER=warehouse
DATABASE_PASSWORD=secret123
DATABASE_NAME=warehouse_db
DATABASE_SSLMODE=disable

# Server Configuration
SERVER_PORT=8081
SERVER_ENV=development
```

## 📋 Зависимости

- **Go**: 1.25.7+
- **PostgreSQL**: 15 (в Docker)
- **Gin-gonic**: REST API фреймворк
- **GORM**: ORM для работы с БД
- **pgx**: PostgreSQL драйвер

## 🛣️ Roadmap

- [ ] **ЭТАП 3**: HTML сканер + WebRTC (браузер → камера → QR)
- [ ] **ЭТАП 4**: HTTPS через nginx + SSL сертификаты
- [ ] Middleware для авторизации (JWT токены)
- [ ] Валидация входных данных
- [ ] Пагинация для истории
- [ ] Отчёты по перемещениям товаров
- [ ] REST API документация (Swagger)

## 🔐 Безопасность (TODO)

⚠️ **Внимание**: Это MVP, для production'а нужно:

1. Использовать **bcrypt** вместо SHA256 для хеширования паролей
2. Реализовать **JWT токены** для авторизации
3. Добавить **CORS** политики
4. Использовать **HTTPS** вместо HTTP
5. Добавить **rate limiting**
6. Валидация и sanitization входных данных
7. Логирование всех действий (особенно важно для аудита)

## 📝 Примеры использования

### Пример 1: Полный цикл перемещения товара

```bash
# 1. Авторизуемся
curl -X POST http://localhost:8081/api/login \
  -H "Content-Type: application/json" \
  -d '{"username":"operator1","password":"password123"}'

# 2. Смотрим где сейчас товар
curl http://localhost:8081/api/item/item1

# 3. Перемещаем товар на новую локацию
curl -X POST http://localhost:8081/api/move \
  -H "Content-Type: application/json" \
  -d '{
    "item_id":"item1",
    "to_location_id":"location3",
    "user_id":"user1",
    "notes":"Перемещение в зону A"
  }'

# 4. Смотрим историю всех перемещений
curl http://localhost:8081/api/item/item1/history
```

## 📞 Поддержка

QR formato:
- **Товары**: `ITEM:item_id` (например: `ITEM:item123`)
- **Локации**: `LOC:location_id` (например: `LOC:location7`)

Это позволяет просто сканировать QR и отправлять ID, а не весь JSON.

---

**Статус**: Beta (ЭТАП 1-2 завершены) ✅  
**Следующий шаг**: ЭТАП 3 - HTML сканер для браузера 📱