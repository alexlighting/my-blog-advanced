package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

// Config содержит параметры подключения к PostgreSQL
type Config struct {
	Host         string
	Port         int
	User         string
	Password     string
	DBName       string
	SSLMode      string
	MaxOpenConns int
	MaxIdleConns int
}

var logger *log.Logger

// NewPostgresDB создает новое подключение к PostgreSQL
func NewPostgresDB(cfg Config, lg *log.Logger) (*sql.DB, error) {
	var err error
	var db *sql.DB
	logger = lg
	//создаем объект базы данных
	db, err = sql.Open("postgres", GetDSN(cfg))
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}
	//проверяем доступность базы
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}
	//устанавливаем ограничения на подключения
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)

	return db, nil
}

// Migrate выполняет миграции базы данных
func Migrate(db *sql.DB) error {
	// TODO: Реализовать применение миграций
	// Шаги:
	// 1. Создать таблицу users если не существует
	// 2. Создать таблицу posts если не существует
	// 3. Создать таблицу comments если не существует
	// 4. Создать необходимые индексы
	// 5. Вернуть ошибку если что-то пошло не так
	//
	// Структура таблиц:
	// - users: id, username, email, password, created_at, updated_at
	// - posts: id, title, content, author_id, created_at, updated_at
	// - comments: id, content, post_id, author_id, created_at, updated_at

	queries := []string{

		`-- Создание таблицы пользователей
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL, 
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);`,

		`-- Индексы для оптимизации поиска
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at);`,

		`-- Добавим комментарии к таблице для документации
COMMENT ON TABLE users IS 'Таблица пользователей блога';
COMMENT ON COLUMN users.email IS 'Email пользователя (уникальный)';
COMMENT ON COLUMN users.username IS 'Имя пользователя (уникальное)';
COMMENT ON COLUMN users.password IS 'Хеш пароля (bcrypt)';
COMMENT ON COLUMN users.created_at IS 'Дата и время регистрации';
COMMENT ON COLUMN users.updated_at IS 'Дата и время последнего обновления';`,

		`-- Создание таблицы постов
CREATE TABLE IF NOT EXISTS posts (
    id SERIAL PRIMARY KEY,
    title VARCHAR(200) NOT NULL,
    content TEXT NOT NULL,
	author_id INTEGER,
	draft BOOLEAN NOT NULL,
    CONSTRAINT fk_user FOREIGN KEY (author_id) REFERENCES users (id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
	publish_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);`,
		`-- Индексы для оптимизации поиска
CREATE INDEX IF NOT EXISTS idx_posts_created_at ON posts(created_at DESC);		
CREATE INDEX IF NOT EXISTS idx_posts_author_id ON posts(author_id);`,

		`-- Добавим комментарии к таблице для документации
COMMENT ON TABLE posts IS 'Таблица статей блога';
COMMENT ON COLUMN posts.title IS 'Заголовок статьи';
COMMENT ON COLUMN posts.content IS 'Содержимое статьи ';
COMMENT ON COLUMN posts.draft IS 'Статья не опубликована (черновик)';
COMMENT ON COLUMN posts.created_at IS 'Дата и время регистрации';
COMMENT ON COLUMN posts.updated_at IS 'Дата и время последнего обновления';
COMMENT ON COLUMN posts.publish_at IS 'Дата и время публикации';`,

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
);`,
		`-- Индексы для оптимизации поиска
CREATE INDEX IF NOT EXISTS idx_comments_created_at ON comments(created_at DESC);		
CREATE INDEX IF NOT EXISTS idx_comments_post_id ON comments(post_id);`,

		`-- Добавим комментарии к таблице для документации
COMMENT ON TABLE comments IS 'Таблица комментариев';
COMMENT ON COLUMN comments.content IS 'Содержимое комментария ';
COMMENT ON COLUMN comments.created_at IS 'Дата и время регистрации';
COMMENT ON COLUMN comments.created_at IS 'Дата и время последнего обновления';`,
	}
	//запускаем транзакцию, чтобы созданная структура базы была полной
	// и не получилось ситуации когда в результате сбоя часть таблиц
	// создана, а часть нет. У нас будет или полная структура базы или ничего
	tx, err := db.Begin()
	if err != nil {
		logger.Printf("Can't start transaction, error: %v \n", err)
		return err
	}
	//в цикле выполняем все запросы из среза queries
	for _, query := range queries {
		_, err := tx.Exec(query)
		if err != nil {
			// в случае ошибки откатываем транзакцию
			logger.Printf("Got error %v when executing %s\n", err, query)
			tx.Rollback()
			return err
		}
	}
	//если все хорошо - коммитим транзакцию
	err = tx.Commit()
	if err != nil {
		logger.Printf("Can't commit transaction, error: %v \n", err)
		return err
	}

	return nil
}

// CheckConnection проверяет соединение с базой данных
func CheckConnection(db *sql.DB) error {
	// TODO: Реализовать проверку соединения
	// Использовать db.Ping() для проверки
	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %v", err)
	}
	return nil
}

// GetDSN формирует строку подключения к PostgreSQL
func GetDSN(cfg Config) string {
	// TODO: Сформировать DSN строку
	// Формат: "host=%s port=%d user=%s password=%s dbname=%s sslmode=%s"

	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.DBName,
		cfg.SSLMode,
	)
}

// Close закрывает соединение с базой данных
func Close(db *sql.DB) error {
	// TODO: Корректно закрыть соединение
	err := db.Close()
	if err != nil {
		return err
	}
	return nil
}

// TestConnection выполняет тестовый запрос к БД (опциональное задание)
func TestConnection(db *sql.DB) error {
	// TODO: Выполнить простой запрос для проверки работы БД
	// Например: SELECT 1
	_, err := db.Exec("SELECT 1")
	if err != nil {
		logger.Printf("Got error %v when executing SELECT 1\n", err)
		return err
	}
	return nil
}
