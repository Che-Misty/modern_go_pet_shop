# go-pet-shop

Учебный backend интернет-магазина на Go: товары, пользователи, заказы,
оформление заказа в транзакции и аналитика.

**Стек:** Go, chi (роутер), pgx/pgxpool (драйвер), PostgreSQL 16,
Docker Compose, mockery + testify (тесты).

## Запуск

Нужны Docker и Docker Compose.

1. Скопируйте переменные окружения и при желании поменяйте пароль:

```bash
   cp .env.example .env
```

2. Поднимите проект одной командой:

```bash
   docker compose up --build
```

   Compose запускает PostgreSQL, дожидается его готовности, применяет
   миграции и стартует приложение на `http://localhost:8080`.

3. Проверка: `GET http://localhost:8080/health`.

Остановить проект и удалить данные (база снова стартует с нуля):

```bash
docker compose down -v
```

### Миграции вручную

```bash
go run ./cmd/migrator --migrations-path ./migrations
```

### Запуск приложения без Docker

Поднимите только базу и запустите приложение локально
(в `.env` хост базы `localhost`):

```bash
docker compose up postgres
go run ./cmd/app
```

## Версии

Каждая версия живёт в своей ветке и включает всё из предыдущей.

### v1: пользователи и товары

Созданы интерфейсы хранилища для пользователей (`User`) и товаров
(`Products`) и их реализация на PostgreSQL.

- **Пользователи:** создание, поиск по email, список всех.
- **Товары:** создание, обновление, удаление, поиск по ID, список всех.

Хендлеры: `CreateUser`, `GetUserByEmail`, `GetAllUsers`, `CreateProduct`,
`UpdateProduct`, `DeleteProduct`, `GetProductByID`, `GetAllProducts`.

### v2: заказы и позиции заказа

Добавлен интерфейс `Orders` и таблицы `orders` и `order_items`.

- **Заказы:** создание заказа, получение по ID, список заказов
  пользователя по email.
- **Позиции:** добавление позиции к заказу, получение позиций заказа.

Хендлеры: `CreateOrder`, `AddOrderItem`, `GetOrderByID`,
`GetOrdersByUserEmail`. Ручка `GET /orders/{id}` вызывает два метода
хранилища (`GetOrderByID` и `GetOrderItemsByOrderID`) и отдаёт заказ
вместе с позициями.

### v3: оформление заказа в транзакции

Добавлен метод `PlaceOrder`. В одной транзакции он:

1. уменьшает `stock` товаров (`UPDATE ... WHERE stock >= $1`),
2. создаёт заказ,
3. добавляет позиции заказа,
4. создаёт запись в `transactions`.

При любой ошибке выполняется откат, частичных записей не остаётся.
Нехватка товара возвращает ошибку `ErrInsufficientStock`,
несуществующий товар возвращает `ErrNotFound`.

Хендлер: `PlaceOrder` (`POST /checkout`).

### v4: история и аналитика

Добавлены два метода хранилища:

- `GetUserOrderHistory`: заказы пользователя с позициями и статусом
  транзакции (`JOIN` и `LEFT JOIN`), результат собирается в структуру
  `OrderDetail`.
- `GetPopularProducts`: топ товаров по сумме продаж (`JOIN`,
  `GROUP BY`, `SUM`), структура `PopularProduct`.

Хендлеры: `GetUserOrderHistory` (`GET /users/history?email=...`),
`GetPopularProducts` (`GET /products/popular`).

## Тесты

Юнит-тесты хендлеров написаны на моках хранилища (mockery + testify).

```bash
go test ./...
```

Если интерфейс хранилища изменился, моки нужно сгенерировать заново:

```bash
go run github.com/vektra/mockery/v3
```