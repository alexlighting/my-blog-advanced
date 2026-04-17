# Blog API - блог-платформа

## Описание проекта

REST API платформа для блога со следующей функциональностью:
* Аутентификация пользователей (JWT)
* CRUD операции для постов
* Комментарии к постам
* Авторизация (только автор может редактировать/удалять свои посты и комментарии)
* Планировщик отложенных публикаций
* graceful shutdown: при остановке сервиса фоновые задачи должны корректно завершаться.
* Кеширование постов на 2 минуты.

### Описание планировщика
Если пользователь указал время публикации в будущем, пост создаётся в статусе черновик. Фоновая горутина-планировщик (ticker + контекст) каждые N секунд:
* проверяет посты, у которых настало время публикации;
* автоматически публикует такие посты;
Обработка публикаций должна происходить конкурентно (воркер или worker pool), логирование процесса публикации обязательно.

## Структура проекта
В соответствии с принципами Clear Architecture структуру проекта представим в следующем виде.

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
│   ├── auth/            # JWT и пароли
│   └── database/        # Подключение к БД
│      ├── postgres.go   # Коненктор к БД postgres
│      └── redis.go      # Инициация кеш Redis
├── migrations/          # SQL миграции
├── docker-compose.yml   # PostgreSQL и Adminer, Redis
├── .env.example         # Пример конфигурации
├── go.mod
└── README.md
```


## Особенности реализации

#### Базовая инфраструктура
1. **pkg/database/postgres.go**
   - подключение к БД
   - миграция структуры БД

2. **pkg/auth/password.go**
   - хеширование паролей (bcrypt)
   - проверка пароля

3. **pkg/auth/jwt.go**
   - генерацию JWT токенов
   - валидацию токенов

#### Репозитории
1. **internal/repository/user_repo.go**
   - методы репозитория пользователя
   
2. **internal/repository/post_repo.go**
   - методы репозиторя постов (CRUD операции и проч.)

3. **internal/repository/comment_repo.go**
   - методы репозитория комментариев (CRUD операции и проч.)

#### Бизнес-логика
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

#### HTTP слой
1. **internal/handler/auth_handler.go**
   - Обработка регистрации и входа
   - Возврат JWT токена

2. **internal/handler/post_handler.go**
   - REST эндпоинты для постов
   - Обработка ошибок

3. **internal/handler/comment_handler.go**
   - Эндпоинты для комментариев

#### Middleware
1. **internal/middleware/auth.go**
   - JWT проверка
   - Добавление user_id в контекст

2. **internal/middleware/logging.go**
   - Логирование запросов
   - Recovery от паник
   - CORS заголовки
   - генерация и добавление RequestID в контекст
   - ограничение макссимального размера запроса
   - получение user IP

#### Сборка приложения
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

- ✅ JWT аутентификация
- ✅ Проверка прав доступа
- ✅ Валидация входных данных
- ✅ Обработка ошибок
- ✅ Пагинация для списков
- 📝 Подробное логирование
- ⚡ Оптимизированные SQL запросы
- 📚 API документация (Swagger/OpenAPI)
- 🕑 планировщик публикаций

## Полезные команды

```bash

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

