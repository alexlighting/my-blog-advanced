-- Миграция для создания начальной схемы базы данных
-- TODO: Реализуйте создание таблиц для блог-платформы

-- Таблица пользователей
-- TODO: Создайте таблицу users со следующими полями:
-- - id (serial, primary key)
-- - username (varchar(50), unique, not null)
-- - email (varchar(255), unique, not null)
-- - password (varchar(255), not null) - для хешированного пароля
-- - created_at (timestamp, not null)
-- - updated_at (timestamp, not null)

-- Пример структуры:
-- CREATE TABLE IF NOT EXISTS users (
--     ...ваши поля здесь...
-- );
		`-- Создание таблицы пользователей
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL, 
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);`
`-- Добавим комментарии к таблице для документации
COMMENT ON TABLE users IS 'Таблица пользователей блога';
COMMENT ON COLUMN users.email IS 'Email пользователя (уникальный)';
COMMENT ON COLUMN users.username IS 'Имя пользователя (уникальное)';
COMMENT ON COLUMN users.password IS 'Хеш пароля (bcrypt)';
COMMENT ON COLUMN users.created_at IS 'Дата и время регистрации';
COMMENT ON COLUMN users.updated_at IS 'Дата и время последнего обновления';`
-- Таблица постов
-- TODO: Создайте таблицу posts со следующими полями:
-- - id (serial, primary key)
-- - title (varchar(200), not null)
-- - content (text, not null)
-- - author_id (integer, foreign key на users.id)
-- - created_at (timestamp, not null)
-- - updated_at (timestamp, not null)
		`-- Создание таблицы постов
CREATE TABLE IF NOT EXISTS posts (
    id SERIAL PRIMARY KEY,
    title VARCHAR(200) NOT NULL,
    content TEXT NOT NULL,
	author_id INTEGER,
    CONSTRAINT fk_user FOREIGN KEY (author_id) REFERENCES users (id) ON DELETE CASCADE,
    draft BOOLEAN NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL, 
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    publish_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL 
);`
`-- Добавим комментарии к таблице для документации
COMMENT ON TABLE posts IS 'Таблица статей блога';
COMMENT ON COLUMN posts.title IS 'Заголовок статьи';
COMMENT ON COLUMN posts.content IS 'Содержимое статьи';
COMMENT ON COLUMN posts.draft IS 'Статья не опубликована (черновик)';
COMMENT ON COLUMN posts.created_at IS 'Дата и время создания';
COMMENT ON COLUMN posts.updated_at IS 'Дата и время последнего обновления';
COMMENT ON COLUMN posts.publish_at IS 'Дата и время публикации';`

-- Таблица комментариев
-- TODO: Создайте таблицу comments со следующими полями:
-- - id (serial, primary key)
-- - content (text, not null)
-- - post_id (integer, foreign key на posts.id)
-- - author_id (integer, foreign key на users.id)
-- - created_at (timestamp, not null)
`-- Создание таблицы комментариев
CREATE TABLE IF NOT EXISTS comments (
    id SERIAL PRIMARY KEY,
    content TEXT NOT NULL,
	post_id INTEGER,
	author_id INTEGER,
    CONSTRAINT fk_post FOREIGN KEY (post_id) REFERENCES posts (id) ON DELETE CASCADE,
    CONSTRAINT fk_comment_user FOREIGN KEY (author_id) REFERENCES users (id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL, 
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);`
`-- Добавим комментарии к таблице для документации
COMMENT ON TABLE comments IS 'Таблица комментариев';
COMMENT ON COLUMN comments.content IS 'Содержимое комментария ';
COMMENT ON COLUMN comments.created_at IS 'Дата и время регистрации';
COMMENT ON COLUMN comments.updated_at IS 'Дата и время последнего обновления';`
-- Индексы
-- TODO: Создайте индексы для оптимизации запросов:
-- - Индекс на posts.author_id для быстрого поиска постов пользователя
-- - Индекс на comments.post_id для быстрого поиска комментариев к посту
-- - Индекс на posts.created_at для сортировки по дате
		`-- Индексы для оптимизации поиска
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at);`
		`-- Индексы для оптимизации поиска
CREATE INDEX IF NOT EXISTS idx_posts_created_at ON posts(created_at DESC);		
CREATE INDEX IF NOT EXISTS idx_posts_author_id ON posts(author_id);`
        `-- Индексы для оптимизации поиска
CREATE INDEX IF NOT EXISTS idx_comments_created_at ON comments(created_at DESC);		
CREATE INDEX IF NOT EXISTS idx_comments_post_id ON comments(post_id);`
-- Подсказки:
-- 1. Используйте IF NOT EXISTS для избежания ошибок при повторном запуске
-- 2. Для foreign key используйте ON DELETE CASCADE для автоматического удаления связанных записей
-- 3. Для timestamp полей можно использовать DEFAULT CURRENT_TIMESTAMP
-- 4. Не забудьте про ограничения (constraints) для валидации данных
