package handler

import (
	"blog-api/internal/model"
	"blog-api/internal/service"
	"context"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

type CommentsResponse struct {
	Comments []*model.Comment `json:"comments"`
	Total    int              `json:"total"`
	Limit    int              `json:"limit"`
	Offset   int              `json:"offset"`
	PostID   int              `json:"post_id"`
}

type CommentHandler struct {
	commentService *service.CommentService
	validate       *validator.Validate
}

func NewCommentHandler(commentService *service.CommentService, validator *validator.Validate) *CommentHandler {
	return &CommentHandler{
		commentService: commentService,
		validate:       validator,
	}
}

// Create обрабатывает создание нового комментария
// POST /api/posts/{id}/comments
// Требует аутентификации
// @Summary Создание нового комментария
// @Tags         comments
// @accept json
// @produce json
// @Param request body model.PostCreateRequest true "данные нового поста"
// @Success 201 {object} model.Post "объект с данными созданного поста"
// @Failure      400  {string}  string "неправильные параметры"
// @Failure      401  {string}  string "попытка создать пост не авторизовавшись"
// @Failure      404  {string}  string "не найден пост"
// @Failure      405  {string}  string "неподдерживаемый метод"
// @Failure      500  {string}  string "ошибка записи в базу данных"
// @Router /api/posts/{id}/comments [POST]
func (h *CommentHandler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	//получаем userID из контекста
	userID, ok := getUserIDFromContext(r.Context())
	if !ok {
		sendErrorResponse(w, service.ErrInvalidCredentials.Error(), http.StatusUnauthorized)
		return
	}
	//парсим входыне данные в model.CommentCreateRequest
	var req *model.CommentCreateRequest
	err := parseJSONRequest(r, &req)
	if err != nil {
		sendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}
	//проверяем полученную структуру на корректность
	err = h.validate.Struct(req)
	if err != nil {
		sendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}
	//получаем id поста из пути
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		sendErrorResponse(w, "Invalid comment ID", http.StatusBadRequest)
		return
	}
	postID := int(id)
	//создаем комментарий
	comment, err := h.commentService.Create(r.Context(), userID, postID, req)
	if err != nil {
		status := http.StatusInternalServerError
		if err == service.ErrPostNotExists {
			status = http.StatusNotFound
		}
		sendErrorResponse(w, err.Error(), status)
		return
	}
	sendJSONResponse(w, comment, http.StatusCreated)
}

// GetByID возвращает комментарий по ID
// GET /api/comments/{id}
// Не требует аутентификации
// @Summary Получение комментария по ID
// @Tags         comments
// @produce json
// @Param id path integer true "ID искомого комментария"
// @Success 200 {object} model.Comment "объект с данными комментария"
// @Failure      400  {string}  string "неправильные параметры"
// @Failure      404  {string}  string "комемнтарий с таким ID не найден"
// @Failure      405  {string}  string "неподдерживаемый метод"
// @Failure      500  {string}  string "ошибка БД"
// @Router /api/comments/{id} [GET]
func (h *CommentHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	// Провяем метод запроса (должен быть GET)

	if r.Method != http.MethodGet {
		sendErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Извлекаем ID из URL
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		sendErrorResponse(w, "Invalid comment ID", http.StatusBadRequest)
		return
	}
	postID := int(id)
	// Получаем комментарий через сервис
	comment, err := h.commentService.GetByID(r.Context(), postID)
	if err != nil {
		if err == service.ErrCommentNotFound {
			sendErrorResponse(w, "Comment not found", http.StatusNotFound)
			return
		}
		if err == service.ErrDBError {
			sendErrorResponse(w, "Failed to get comment from db", http.StatusInternalServerError)
			return
		}
		return
	}
	sendJSONResponse(w, comment, http.StatusOK)
}

// GetByPost возвращает комментарии к посту
// GET /api/posts/{id}/comments?limit=20&offset=0
// Не требует аутентификации
// @Summary Получение комментария по ID
// @Tags         comments
// @produce json
// @Param id path integer true "ID искомого поста"
// @Param limit path integer true "количество комментариев на страницу"
// @Param offset path integer true "смещение от начала списка"
// @Success 200 {object} CommentsResponse "объект со списком комментариев, общим количеством комментариев лимитом и смещением"
// @Failure      400  {string}  string "неправильные параметры"
// @Failure      404  {string}  string "комемнтарий с таким ID не найден"
// @Failure      405  {string}  string "неподдерживаемый метод"
// @Failure      500  {string}  string "ошибка БД"
// @Router /api/posts/{id}/comments{limit}{offset} [GET]
func (h *CommentHandler) GetByPost(w http.ResponseWriter, r *http.Request) {
	// Провяем метод запроса (должен быть GET)
	if r.Method != http.MethodGet {
		sendErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	//получаем id поста из пути
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		sendErrorResponse(w, "Invalid post ID", http.StatusBadRequest)
		return
	}
	postID := int(id)
	//получаем limit из пути
	limit, err := strconv.ParseInt(r.URL.Query().Get("limit"), 10, 64)
	if err != nil {
		sendErrorResponse(w, "limit must be positive integer", http.StatusBadRequest)
		return
	}
	//получаем offset из пути
	offset, err := strconv.ParseInt(r.URL.Query().Get("offset"), 10, 64)
	if err != nil {
		sendErrorResponse(w, "offset must be positive integer", http.StatusBadRequest)
		return
	}

	comments, total, err := h.commentService.GetByPost(r.Context(), postID, int(limit), int(offset))
	if err != nil {
		if err == service.ErrPostNotExists {
			sendErrorResponse(w, "Post not found", http.StatusNotFound)
		} else {
			sendErrorResponse(w, "Failed to get comments", http.StatusInternalServerError)
		}
		return
	}

	resp := CommentsResponse{
		Comments: comments,
		Total:    total,
		Limit:    int(limit),
		Offset:   int(offset),
		PostID:   postID,
	}

	sendJSONResponse(w, resp, http.StatusOK)
}

// Update обновляет комментарий
// PUT /api/comments/{id}
// Требует аутентификации, может обновить только автор
// @Summary Изменение комментария
// @Tags         comments
// @accept json
// @produce json
// @Param request body model.CommentUpdateRequest true "ID изменяемого комментария"
// @Param id path integer true "ID изменяемого комментария"
// @Success 200 {object} model.Comment "объект с данными измененного комментария"
// @Failure      400  {string}  string "неправильные параметры"
// @Failure      403  {string}  string "попытка изменить чужой комментарий"
// @Failure      404  {string}  string "комментарий с таким ID не найден"
// @Failure      405  {string}  string "неподдерживаемый метод"
// @Failure      500  {string}  string "ошибка БД"
// @Router /api/comments/{id} [PUT]
func (h *CommentHandler) Update(w http.ResponseWriter, r *http.Request) {
	// Провяем метод запроса (должен быть Put)
	if r.Method != http.MethodPut {
		sendErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	//получаем userID из контекста
	userID, ok := getUserIDFromContext(r.Context())
	if !ok {
		sendErrorResponse(w, service.ErrInvalidCredentials.Error(), http.StatusUnauthorized)
		return
	}
	//получаем id комментария из пути
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		sendErrorResponse(w, "Invalid comment ID", http.StatusBadRequest)
		return
	}
	commentID := int(id)
	//парсим входной JSON в model.CommentUpdateRequest
	var req model.CommentUpdateRequest
	err = parseJSONRequest(r, &req)
	if err != nil {
		sendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}
	//проверяем полученную структуру на корректность
	err = h.validate.Struct(req)
	if err != nil {
		sendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}
	//обновляем комментарий
	comment, err := h.commentService.Update(r.Context(), commentID, userID, &req)
	if err != nil {
		switch err {
		case service.ErrForbidden:
			sendErrorResponse(w, "You can only update your own comments", http.StatusForbidden)
		case service.ErrCommentNotFound:
			sendErrorResponse(w, "Comment not found", http.StatusNotFound)
		default:
			sendErrorResponse(w, "Failed to update comment", http.StatusInternalServerError)
		}
		return
	}
	sendJSONResponse(w, comment, http.StatusOK)
}

// getUserIDFromContext извлекает ID пользователя из контекста
func getUserIDFromContext(ctx context.Context) (int, bool) {
	userID, ok := ctx.Value("userID").(int)
	// fmt.Printf("userID: %d, ok: %t\n", userID, ok)
	if !ok {
		return 0, ok
	}
	return userID, ok
}
