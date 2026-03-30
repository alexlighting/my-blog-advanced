package middleware

import (
	"context"
	"log"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/google/uuid"
)

const MAX_CONTENT_LENGTH = 1000000

// LoggingMiddleware provides request logging, CORS, recovery and other utility middleware
type LoggingMiddleware struct {
	logger *log.Logger
}

// NewLoggingMiddleware creates a new logging middleware instance
func NewLoggingMiddleware(logger *log.Logger) *LoggingMiddleware {
	return &LoggingMiddleware{
		logger: logger,
	}
}

// Logger logs all HTTP requests
func (m *LoggingMiddleware) Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		nw := newResponseWriter(w)
		clientIP := getClientIP(r)
		next.ServeHTTP(nw, r)
		m.logger.Printf("%s %s %s %d %s", r.Method, r.URL.Path, clientIP, nw.statusCode, time.Since(startTime))
	})
}

// Recovery восстанавливается после паник
func (m *LoggingMiddleware) Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//при возникновении паники записываем stack trace и вызываем следующий хендлер
		defer func() {
			if err := recover(); err != nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				m.logger.Printf("Error: %v, stack trace below", err)
				m.logger.Println(string(debug.Stack()))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// CORS добавляет CORS заголовки
func (m *LoggingMiddleware) CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Request-Method", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Max-Age", "10")

		if r.Method == "OPTIONS" {
			//handle preflight in here
			w.WriteHeader(http.StatusNoContent)
		} else {
			// Временная реализация
			next.ServeHTTP(w, r)
		}
	})
}

// RequestID добавляет уникальный ID к каждому запросу
func (m *LoggingMiddleware) RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Генерируем уникальный UUID
		id := uuid.New().String()
		// Добавляем ID в контекст запроса для использования в логах
		ctx := context.WithValue(r.Context(), "id", id)
		// 3. Добавляем ID в заголовок ответа X-Request-ID
		w.Header().Set("X-Request-ID", id)
		// Вызваем следующий handler
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ContentTypeJSON устанавливает Content-Type: application/json для всех ответов
func (m *LoggingMiddleware) ContentTypeJSON(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Устанавливаем Content-Type: application/json для всех ответов
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

// SetMaxBytesReader устанавливает лимит на размер content.Body дабы осложнить DDoS атаки на сервер
func (m *LoggingMiddleware) SetMaxBytesReader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Устанавливаем Content-Type: application/json для всех ответов
		if r.ContentLength > MAX_CONTENT_LENGTH {
			m.logger.Printf("Body size limit exeed, %d", r.ContentLength)
			w.WriteHeader(http.StatusRequestEntityTooLarge)
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, MAX_CONTENT_LENGTH)
		next.ServeHTTP(w, r)
	})
}

// getClientIP извлекает IP адрес клиента
func getClientIP(r *http.Request) string {
	var userIP string
	//получаем заголовок X-Forwarded-For
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		//если в загогловке несколько IP, то берем первый (обычно это адрес клиента)
		if strings.Contains(forwarded, ",") {
			addrs := strings.Split(forwarded, ",")
			userIP = addrs[0]
		} else {
			userIP = forwarded
		}
		return userIP
	}
	//если не нашли в X-Forwarded-For то проверяем X-Real-IP
	real := r.Header.Get("X-Real-IP")
	if real != "" {
		return real
	}
	//если и его нет то берем RemoteAddr
	return r.RemoteAddr
}

// responseWriter обертка для захвата статус кода
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

// WriteHeader сохраняет статус код
func (rw *responseWriter) WriteHeader(code int) {
	if !rw.written {
		rw.statusCode = code
		rw.ResponseWriter.WriteHeader(code)
		rw.written = true
	}
}

// Write вызывает WriteHeader если еще не был вызван
func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.written {
		rw.WriteHeader(http.StatusOK)
	}
	return rw.ResponseWriter.Write(b)
}

// newResponseWriter создает новую обертку
func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
		written:        false,
	}
}
