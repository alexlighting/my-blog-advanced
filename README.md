# Blog API - Шаблон проектной работы

## Описание проекта

Вам необходимо реализовать REST API для блога с функциональностью:
- Аутентификация пользователей (JWT)
- CRUD операции для постов
- Комментарии к постам
- Авторизация (только автор может редактировать/удалять свои посты и комментарии)

## Структура проекта

```
blog-api/
├── cmd/api/              # Точка входа приложения
│   └── main.go
├── docs/                 # Документация Swagger
├── internal/             # Внутренние пакеты приложения
│   ├── model/           # Модели данных
│   ├── handler/         # HTTP хендлеры
│   ├── service/         # Бизнес-логика
│   ├── repository/      # Работа с БД
│   └── middleware/      # HTTP middleware
├── pkg/                 # Переиспользуемые пакеты
│   ├── auth/           # JWT и пароли
│   └── database/       # Подключение к БД
├── migrations/         # SQL миграции
├── docker-compose.yml  # PostgreSQL и Adminer
├── .env.example        # Пример конфигурации
├── go.mod
└── README.md
```

## Начало работы

### 1. Подготовка окружения

```bash
# Клонировать шаблон
cp -r template my-blog-api
cd my-blog-api

# Установить зависимости
go mod download

# Создать файл конфигурации
cp .env.example .env

# Запустить PostgreSQL
docker-compose up -d

# Проверить что БД работает (опционально)
docker-compose logs postgres
```

### 2. Порядок реализации

#### Этап 1: Базовая инфраструктура
1. **pkg/database/postgres.go**
   - Реализовать подключение к БД
   - Реализовать функцию миграций

2. **pkg/auth/password.go**
   - Реализовать хеширование паролей (bcrypt)
   - Реализовать проверку пароля

3. **pkg/auth/jwt.go**
   - Реализовать генерацию JWT токенов
   - Реализовать валидацию токенов

#### Этап 2: Репозитории
1. **internal/repository/user_repo.go**
   - Завершить реализацию всех методов
   - SQL запросы уже подготовлены

2. **internal/repository/post_repo.go**
   - Завершить реализацию CRUD операций
   - Добавить методы пагинации

3. **internal/repository/comment_repo.go**
   - Завершить реализацию работы с комментариями

#### Этап 3: Бизнес-логика
1. **internal/service/user_service.go**
   - Регистрация с валидацией
   - Вход с проверкой пароля
   - Генерация JWT токена

2. **internal/service/post_service.go**
   - CRUD операции с проверкой прав
   - Пагинация и фильтрация

3. **internal/service/comment_service.go**
   - Создание комментариев
   - Проверка прав на редактирование

#### Этап 4: HTTP слой
1. **internal/handler/auth_handler.go**
   - Обработка регистрации и входа
   - Возврат JWT токена

2. **internal/handler/post_handler.go**
   - REST эндпоинты для постов
   - Обработка ошибок

3. **internal/handler/comment_handler.go**
   - Эндпоинты для комментариев

#### Этап 5: Middleware
1. **internal/middleware/auth.go**
   - JWT проверка
   - Добавление user_id в контекст

2. **internal/middleware/logging.go**
   - Логирование запросов
   - Recovery от паник
   - CORS заголовки

#### Этап 6: Сборка приложения
1. **cmd/api/main.go**
   - Инициализация всех компонентов
   - Настройка маршрутов
   - Запуск сервера

## API Эндпоинты

### Публичные (без аутентификации)
- `POST /api/register` - регистрация пользователя
- `POST /api/login` - вход пользователя
- `GET /api/posts` - список постов
- `GET /api/posts/{id}` - получить пост
- `GET /api/posts/{id}/comments` - комментарии к посту
- `GET /api/posts/author/{id}` - получить все посты автора
- `GET /api/health` - проверить состояние сервера


### Защищенные (требуют JWT токен)
- `POST /api/posts` - создать пост
- `PUT /api/posts/{id}` - обновить пост (только автор)
- `DELETE /api/posts/{id}` - удалить пост (только автор)
- `POST /api/posts/{id}/comments` - создать комментарий к посту
- `PUT /api/comments/{id}` - изменить комментарий к посту
- `GET /api/porfile/` - получить профиль пользователя

## Требования к реализации

### Обязательные требования
- ✅ Все основные эндпоинты работают
- ✅ JWT аутентификация реализована
- ✅ Проверка прав доступа работает
- ✅ Валидация входных данных
- ✅ Обработка ошибок
- ✅ Пагинация для списков

### Дополнительные требования (для высокой оценки)
- 📊 Кеширование часто запрашиваемых данных
- 🔍 Поиск и фильтрация постов
- 📝 Подробное логирование
- ⚡ Оптимизированные SQL запросы
- 🧪 Юнит-тесты для критической логики
- 📚 API документация (Swagger/OpenAPI)

## Полезные команды

```bash
# Запуск приложения
go run cmd/api/main.go

# Запуск с hot-reload (установить air)
air

# Тестирование
go test ./...

# Проверка на ошибки
go vet ./...
golangci-lint run

# Форматирование кода
go fmt ./...

# База данных
docker-compose up -d    # Запустить
docker-compose down      # Остановить
docker-compose logs -f   # Логи

# Примеры запросов
# Регистрация
curl -X POST http://localhost:8080/api/register \
  -H "Content-Type: application/json" \
  -d '{"username" :"alex", "email" :"alex@somemail.info", "password" : "NDy1s$32"}'

# Вход
curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"email":"alex@somemail.info","password":"NDy1s$32"}'

# Создание поста (с токеном)
curl -X POST http://localhost:8080/api/posts \
  -H "Authorization: Bearer _YOUR_TOKEN_" \
  -H "Content-Type: application/json" \
  -d '{"title":"My Post","content":"Post content","publish_at" : "2026-02-16T18:23:35.2343209+03:00"}'
```

## Где искать подсказки

1. **TODO комментарии** - в каждом файле есть указания что нужно реализовать
2. **Интерфейсы** - в `repository/interfaces.go` описаны все методы
3. **Модели** - в `model/` определены все структуры данных
4. **SQL запросы** - базовые запросы уже есть в репозиториях
5. **Примеры из solution** - можете подсмотреть в готовое решение при затруднениях

## Частые ошибки

1. **Не забудьте обработку ошибок** - всегда проверяйте err != nil
2. **SQL injection** - используйте placeholder'ы ($1, $2) в SQL запросах
3. **Контекст** - передавайте context во все методы работы с БД
4. **Закрытие ресурсов** - используйте defer для rows.Close()
5. **Права доступа** - проверяйте что пользователь может редактировать только свои данные

## Критерии оценки

### Минимум для зачета (60%)
- Работают эндпоинты регистрации и входа
- Можно создавать и получать посты
- JWT токены генерируются и проверяются

### Хорошо (80%)
- Все CRUD операции работают
- Реализована проверка прав доступа
- Корректная обработка ошибок
- Пагинация работает

### Отлично (100%)
- Код хорошо структурирован
- Добавлены дополнительные функции
- Есть тесты
- Оптимизированы запросы к БД
- Документирован API

## Полезные ссылки

- [Go database/sql tutorial](http://go-database-sql.org/)
- [JWT in Go](https://github.com/golang-jwt/jwt)
- [Chi router](https://github.com/go-chi/chi)
- [bcrypt in Go](https://pkg.go.dev/golang.org/x/crypto/bcrypt)
- [PostgreSQL documentation](https://www.postgresql.org/docs/)