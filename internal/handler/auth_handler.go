package handler

import (
	"blog-api/internal/model"
	"blog-api/internal/service"
	"log"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
)

type AuthHandler struct {
	userService *service.UserService
	validate    *validator.Validate
	logger      *log.Logger
}

func NewAuthHandler(userService *service.UserService, validator *validator.Validate, logger *log.Logger) *AuthHandler {
	return &AuthHandler{
		userService: userService,
		validate:    validator,
		logger:      logger,
	}
}

// Register обрабатывает запрос на регистрацию нового пользователя
// POST /api/register
// @Summary Зарегистрировать нового пользователя
// @Tags         users
// @accept json
// @produce json
// @Param request body model.UserCreateRequest true "Парамеры нового пользователя"
// @Success 201 {object} model.TokenResponse "Токен авторизации"
// @Failure      400  {string}  string "неправильные параметры"
// @Failure      405  {string}  string "неподдерживаемый метод"
// @Failure      409  {string}  string "пользователь с таким именем уже существует"
// @Failure      500  {string}  string "ошибка записи в базу данных"
// @Router /api/register [POST]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req *model.UserCreateRequest
	//декодируем JSON из запроса
	err := parseJSONRequest(r, &req)
	if err != nil {
		sendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}
	//проверяем входные данные
	err = h.validate.Struct(req)
	if err != nil {
		sendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}
	tokenResponse, err := h.userService.Register(r.Context(), req)
	//Если не удалось добавить пользователя в базу
	if err != nil {
		status := http.StatusBadRequest
		if err == service.ErrUserAlreadyExists {
			status = http.StatusConflict
		}
		if strings.Contains(err.Error(), "Failed to write user data to db") ||
			strings.Contains(err.Error(), "Failed create user") {
			status = http.StatusInternalServerError
		}
		sendErrorResponse(w, err.Error(), status)
		return
	}
	//если все прошло успешно
	sendJSONResponse(w, tokenResponse, http.StatusCreated)
}

// Login обрабатывает запрос на вход пользователя
// POST /api/login
// @Summary Авторизация пользователя
// @Tags         users
// @accept json
// @produce json
// @Param request body model.UserLoginRequest true "данные для входа"
// @Success 200 {object} model.TokenResponse "Токен авторизации"
// @Failure      400  {string}  string "неправильные параметры"
// @Failure      401  {string}  string "ошибка авторизации"
// @Failure      405  {string}  string "неподдерживаемый метод"
// @Failure      500  {string}  string "ошибка подписи токена"
// @Router /api/login [POST]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req *model.UserLoginRequest
	//декодируем JSON из запроса
	err := parseJSONRequest(r, &req)
	if err != nil {
		h.logger.Printf("json parse error: %v", err)
		sendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}
	//проверяем входные данные
	err = h.validate.Struct(req)
	if err != nil {
		sendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}
	tokenResponse, err := h.userService.Login(r.Context(), req)
	if err != nil {
		status := http.StatusBadRequest
		if err == service.ErrInvalidCredentials {
			status = http.StatusUnauthorized
		}
		if err == service.ErrTokenSignFailed {
			h.logger.Println("Token sign failed")
			status = http.StatusInternalServerError
		}
		sendErrorResponse(w, err.Error(), status)
		return
	}

	//если все прошло успешно
	sendJSONResponse(w, tokenResponse, http.StatusOK)
}

// GetProfile возвращает профиль текущего пользователя (опционально)
// Этот метод не используется в эталонной реализации
// GET /api/porfile/
// @Summary Профиль пользователя
// @Tags         users
// @accept json
// @produce json
// @Success 200 {object} model.UserResponse "объект с данными пользователя"
// @Failure      400  {string}  string "неправильные параметры"
// @Failure      401  {string}  string "ошибка авторизации"
// @Failure      404  {string}  string "пользователь не найден в БД"
// @Failure      405  {string}  string "неподдерживаемый метод"
// @Router /api/porfile [POST]
func (h *AuthHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userID, ok := getUserIDFromContext(r.Context())
	if !ok {
		sendErrorResponse(w, service.ErrUserNotFound.Error(), http.StatusUnauthorized)
		return
	}
	user, err := h.userService.GetByID(r.Context(), userID)
	if err != nil {
		sendErrorResponse(w, service.ErrUserNotFound.Error(), http.StatusNotFound)
		return
	}
	sendJSONResponse(w, user.ToResponse(), http.StatusOK)
}
