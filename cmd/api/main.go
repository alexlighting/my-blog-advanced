package main

import (
	"blog-api/internal/handler"
	"blog-api/internal/middleware"
	"blog-api/internal/repository"
	"blog-api/internal/service"
	"blog-api/pkg/auth"
	"blog-api/pkg/database"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"
	"unicode"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"

	// "github.com/redis/go-redis/v9"
	"github.com/joho/godotenv"
)

const NUM_WORKERS = 5
const DRAFT_PUBLISH_TIMEOUT = 10

// @title blog-api
// @version 1.0
// @description Back-end для блог-платформы с поддержкой постов и комментариев к ним.
// @contact.name alex.derakchev
// @contact.email alex.derakchev@gmail.com
func main() {
	//регистрируем обработчик системных сигналов
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	file, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	logger := log.New(file, "", log.Ldate|log.Ltime|log.Lshortfile)
	logger.Println("Start logging system")

	// Загружаем конфигурацию из .env файла
	if err := godotenv.Load(); err != nil {
		logger.Printf("Warning: .env file not found, using environment variables")
	}

	// Загружаем конфигурацию из переменных окружения
	cfg := loadConfig()

	// Подключаемся к базе данных
	dbConfig := database.Config{
		Host:         cfg.DBHost,
		Port:         cfg.DBPort,
		User:         cfg.DBUser,
		Password:     cfg.DBPassword,
		DBName:       cfg.DBName,
		SSLMode:      cfg.DBSSLMode,
		MaxOpenConns: cfg.MaxOpenConns,
		MaxIdleConns: cfg.MaxIdleConns,
	}
	db, err := database.NewPostgresDB(dbConfig, logger)
	if err != nil {
		logger.Printf("Error create db: %v", err)
		panic("Can't create DB")
	}
	defer db.Close()
	err = database.Migrate(db)
	if err != nil {
		logger.Printf("Error migrate db: %v", err)
		panic("Can't migrate DB")
	}

	//подключаемся к кеш Redis
	redisConfig := database.RedisConfig{
		Addr: fmt.Sprintf("%s:%d", cfg.RedisHost, cfg.RedisPort),
		// User:        cfg.RedisUser,
		Password:    cfg.RedisPassword,
		DB:          0,
		MaxRetries:  cfg.RedisMaxRetries,
		DialTimeout: time.Duration(cfg.RedisDialTimeout) * time.Second,
		Timeout:     time.Duration(cfg.RedisTimeout) * time.Second,
	}
	redis, err := database.NewRedisClient(ctx, redisConfig)
	if err != nil {
		panic(err)
	}
	fmt.Println("Redis ready to work")
	// redis.Ping(ctx)
	//TODO дописать работу с Redis

	//создаем валидатор
	validate := validator.New()

	//Custom field validator for password strength
	validate.RegisterValidation("strongpassword", func(fl validator.FieldLevel) bool {
		var (
			str        = fl.Field().String()
			hasUpper   = false
			hasLower   = false
			hasNumber  = false
			hasSpecial = false
		)
		for _, char := range str {
			switch {
			case unicode.IsUpper(char):
				hasUpper = true
			case unicode.IsLower(char):
				hasLower = true
			case unicode.IsNumber(char):
				hasNumber = true
			case unicode.IsPunct(char) || unicode.IsSymbol(char):
				hasSpecial = true
			}
		}
		return hasUpper && hasLower && hasNumber && hasSpecial
	})

	// TODO: Инициализировать JWT менеджер
	// - Создать jwtManager через auth.NewJWTManager
	jwtManager := auth.NewJWTManager(cfg.JWTSecret, cfg.JWTExpiryHours)
	// TODO: Создать слои приложения
	// 1. Репозитории (передать db)
	userRepo := repository.NewUserRepo(db, logger)
	postRepo := repository.NewPostRepo(db, logger)
	commentRepo := repository.NewCommentRepo(db, logger)
	// 2. Сервисы (передать репозитории и jwtManager)
	userService := service.NewUserService(userRepo, jwtManager, logger)
	postService := service.NewPostService(postRepo, userRepo)
	commenService := service.NewCommentService(commentRepo, postRepo, userRepo, logger)
	// 3. Хендлеры (передать сервисы)
	authHandler := handler.NewAuthHandler(userService, validate, logger)
	postHandler := handler.NewPostHandler(postService, validate, redis)
	commentHandler := handler.NewCommentHandler(commenService, validate)
	// 4. Middleware (передать необходимые зависимости)
	authMiddleware := middleware.NewAuthMiddleware(jwtManager)

	loggingMiddleware := middleware.NewLoggingMiddleware(logger)
	// Настраиваем маршруты
	router := chi.NewRouter()

	// TODO: Настроить middleware
	// - Добавить глобальные middleware (logging, recovery, CORS)
	router.Use(loggingMiddleware.Logger)
	router.Use(loggingMiddleware.Recovery)
	router.Use(loggingMiddleware.CORS)
	router.Use(loggingMiddleware.RequestID)
	router.Use(loggingMiddleware.ContentTypeJSON)
	// router.Use(loggingMiddleware.SetMaxBytesReader)

	// TODO: Настроить маршруты
	// Публичные эндпоинты:
	// - POST /api/register
	router.Post("/api/register", authHandler.Register)
	// - POST /api/login
	router.Post("/api/login", authHandler.Login)
	// - GET /api/posts
	router.Get("/api/posts", authMiddleware.OptionalAuth(postHandler.GetAll))
	// - GET /api/posts/{id}
	router.Get("/api/posts/{id}", authMiddleware.OptionalAuth(postHandler.GetByID))
	// - GET /api/posts/author/{id}
	router.Get("/api/posts/author/{id}", authMiddleware.OptionalAuth(postHandler.GetByAuthor))
	// - GET /api/posts/{id}/comments
	router.Get("/api/posts/{id}/comments", authMiddleware.OptionalAuth(commentHandler.GetByPost))
	//
	// Защищенные эндпоинты (требуют JWT):

	// - POST /api/posts
	router.Post("/api/posts", authMiddleware.RequireAuth(postHandler.Create))
	// - PUT /api/posts/{id}
	router.Put("/api/posts/{id}", authMiddleware.RequireAuth(postHandler.Update))
	// - DELETE /api/posts/{id}
	router.Delete("/api/posts/{id}", authMiddleware.RequireAuth(postHandler.Delete))
	// - POST /api/posts/{id}/comments
	router.Post("/api/posts/{id}/comments", authMiddleware.RequireAuth(commentHandler.Create))
	router.Put("/api/comments/{id}", authMiddleware.RequireAuth(commentHandler.Update))
	router.Get("/api/porfile/", authMiddleware.RequireAuth(authHandler.GetProfile))

	// Health check эндпоинт
	router.Get("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","service":"blog-api"}`))
	})

	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.ServerHost, cfg.ServerPort),
		Handler: router,
	}
	logger.Printf("🚀 Server starting on port %d", cfg.ServerPort)

	log.Printf("🚀 Server starting on port %d", cfg.ServerPort)
	log.Printf("👤 Register: POST http://localhost:%d/api/register", cfg.ServerPort)
	log.Printf("🔐 Login: POST http://localhost:%d/api/login", cfg.ServerPort)
	log.Printf("📝 Post: POST http://localhost:%d//api/posts (requires token)", cfg.ServerPort)
	log.Printf("❤️  Health: GET http://localhost:%d/api/health", cfg.ServerPort)

	//запускаем http сервер в отдельной горутине чтобы main не блокировалась и была возможность
	//отработать gracefull shutdown
	go func() {
		srv.ListenAndServe()
	}()

	//Создаем тикер для периодической публикации постов у которых подошел срок
	ticker := time.NewTicker(DRAFT_PUBLISH_TIMEOUT * time.Second)
	defer ticker.Stop()
	//запускем его периодической фоновой задачей
	shedulerContext, shedulerCancel := context.WithCancel(context.Background())
	go func() {
		for {
			<-ticker.C
			sheduler(shedulerContext, postRepo, logger)
		}
	}()

	<-ctx.Done()
	//пришел сигнал на завершение работы приложения
	logger.Println("Received termination signal, shutting down...")
	logger.Println("Shutting down http server...")
	//создаем контекст с таймаутом на закрытие соединений
	ctx_shutdown, cancelFn := context.WithTimeout(context.Background(), time.Duration(cfg.ShutdownTimeout)*time.Second)
	defer cancelFn()
	//останавливаем фоновые задачи
	shedulerCancel()
	//и вызываем Shutdown у HTTP сервера
	if err = srv.Shutdown(ctx_shutdown); err != nil {
		logger.Printf("shutdown: %v", err)
	}
	//после остановки http останавливаем базу данных
	logger.Println("Shutting down DB connection...")
	err = database.Close(db)
	if err != nil {
		logger.Printf("Error close db connection: %v", err)
		panic("Can't close DB connection correctly")
	}
	logger.Println("Shutting down done, exit.")
}

// Config представляет конфигурацию приложения
type Config struct {
	// Server
	ServerHost string
	ServerPort int

	// Database
	DBHost       string
	DBPort       int
	DBUser       string
	DBPassword   string
	DBName       string
	DBSSLMode    string
	MaxOpenConns int
	MaxIdleConns int

	// JWT
	JWTSecret      string
	JWTExpiryHours int

	// Cache
	CacheTTLMinutes int

	// Shutdown, second
	ShutdownTimeout int

	// Redis
	RedisHost        string
	RedisPort        int
	RedisUser        string
	RedisPassword    string
	RedisMaxRetries  int
	RedisDialTimeout int
	RedisTimeout     int
}

// loadConfig загружает конфигурацию из переменных окружения
func loadConfig() *Config {
	conf := Config{}
	conf.ServerHost = getEnv("SERVER_HOST", "localhost")
	conf.ServerPort = getEnvAsInt("SERVER_PORT", 8080)
	conf.DBHost = getEnv("DB_HOST", "localhost")
	conf.DBPort = getEnvAsInt("DB_PORT", 5432)
	conf.DBUser = getEnv("DB_USER", "bloguser")
	conf.DBPassword = getEnv("DB_PASSWORD", "blogpassword")
	conf.DBName = getEnv("DB_NAME", "blogdb")
	conf.DBSSLMode = getEnv("DB_SSLMODE", "disable")
	conf.MaxOpenConns = getEnvAsInt("DB_MAXOPENCONNS", 5)
	conf.MaxIdleConns = getEnvAsInt("DB_MAXIDLECONNS", 5)
	conf.JWTSecret = getEnv("JWT_SECRET", "o3h0JpmxOLoMgK1IMEGiX9QDjcX1P4WPkpdMGK95ZgY")
	conf.JWTExpiryHours = getEnvAsInt("JWT_EXPIRY_HOURS", 24)
	conf.CacheTTLMinutes = getEnvAsInt("CACHE_TTL_MINUTES", 5)
	conf.ShutdownTimeout = getEnvAsInt("ShutdownTimeout", 5)
	conf.RedisHost = getEnv("REDIS_HOST", "localhost")
	conf.RedisPort = getEnvAsInt("REDIS_PORT", 6380)
	conf.RedisUser = getEnv("REDIS_USER", "RedisUser")
	conf.RedisPassword = getEnv("REDIS_PASSWORD", "mypassword")
	conf.RedisMaxRetries = getEnvAsInt("REDIS_MAX_RETRIES", 5)
	conf.RedisDialTimeout = getEnvAsInt("REDIS_DIAL_TIMEOUT", 10)
	conf.RedisTimeout = getEnvAsInt("REDIS_TIMEOUT", 5)
	return &conf
}

// getEnv получает значение переменной окружения или возвращает значение по умолчанию
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt получает значение переменной окружения как int или возвращает значение по умолчанию
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

// sheduler запускает обработку отложенной публикации постов
func sheduler(ctx context.Context, postRepo *repository.PostRepo, logger *log.Logger) {
	log.Println("Start sheduler")
	jobs := make(chan int, NUM_WORKERS)
	var wg sync.WaitGroup
	//запускаем пул воркеров
	for i := 1; i <= NUM_WORKERS; i++ {
		wg.Add(1)
		go worker(jobs, ctx, postRepo, logger, &wg)
	}
	//получаем список черновиков
	draftPosts, err := postRepo.GetDraftToUpdate(ctx)
	if err != nil {
		log.Printf("%s\n", err.Error())
	}
	//заполняем ими канал задач
	for _, j := range draftPosts {
		jobs <- j
	}
	close(jobs)
	//ждем завершения воркеров
	go func() {
		wg.Wait()
	}()
	log.Println("Sheduler done")
}

// worker выполняет отложенную публикацию поста
func worker(jobs <-chan int, ctx context.Context, postRepo *repository.PostRepo, logger *log.Logger, wg *sync.WaitGroup) {
	defer wg.Done()
	for job := range jobs {
		select {
		default:
			err := postRepo.PublishById(ctx, job)
			if err != nil {
				logger.Printf("Can't publish draft post with id=%d, error: %s\n", job, err.Error())
			} else {
				logger.Printf("Draft post with id=%d, published successfully\n", job)
			}
		case <-ctx.Done():
			log.Println("Stop signal received, colsing...")
			return
		}
	}
}
