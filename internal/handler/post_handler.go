package handler

import (
	"blog-api/internal/model"
	"blog-api/internal/service"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/redis/go-redis/v9"
)

type PostHandler struct {
	postService *service.PostService
	validate    *validator.Validate
	redis       *redis.Client
}

type PostsResponse struct {
	Posts    []*model.Post `json:"posts"`
	Total    int           `json:"total"`
	Limit    int           `json:"limit"`
	Offset   int           `json:"offset"`
	AuthorID int           `json:"author_id"`
}

func NewPostHandler(postService *service.PostService, validator *validator.Validate, redis *redis.Client) *PostHandler {
	return &PostHandler{
		postService: postService,
		validate:    validator,
		redis:       redis,
	}
}

// Create обрабатывает создание нового поста
// POST /api/posts
// Требует аутентификации
// @Summary Создание нового поста
// @Tags         posts
// @accept json
// @produce json
// @Param request body model.PostCreateRequest true "данные нового поста"
// @Success 201 {object} model.Post "объект с данными созданного поста"
// @Failure      400  {string}  string "неправильные параметры"
// @Failure      401  {string}  string "попытка создать пост не авторизовавшись"
// @Failure      405  {string}  string "неподдерживаемый метод"
// @Failure      500  {string}  string "ошибка записи в базу данных/кэш"
// @Router /api/posts [POST]
func (h *PostHandler) Create(w http.ResponseWriter, r *http.Request) {
	// Проверяем метод запроса (должен быть POST)
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	//получаем userID из контекста
	userID, ok := getUserIDFromContext(r.Context())
	if !ok {
		sendErrorResponse(w, service.ErrUserNotFound.Error(), http.StatusUnauthorized)
		return
	}
	// Декодируем JSON в PostCreateRequest
	var postCreateReq *model.PostCreateRequest
	err := parseJSONRequest(r, &postCreateReq)
	if err != nil {
		sendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	//проверяем полученную структуру на корректность
	err = h.validate.Struct(postCreateReq)
	if err != nil {
		sendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Создаем пост через postService.Create
	post, err := h.postService.Create(r.Context(), userID, postCreateReq)
	if err != nil {
		status := http.StatusBadRequest
		if strings.Contains(err.Error(), "Failed to write post data to db") ||
			strings.Contains(err.Error(), "Failed create post") {
			status = http.StatusInternalServerError
		}
		sendErrorResponse(w, err.Error(), status)
		return
	}
	err = post.ToRedisSet(r.Context(), h.redis, strconv.Itoa(post.ID))
	if err != nil {
		//беда случилась, редис отвалился
		sendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Printf("Положили в кэш пост с id=%d\n", post.ID)
	sendJSONResponse(w, post, http.StatusCreated)
}

// GetByID возвращает пост по ID
// GET /api/posts/{id}
// Не требует аутентификации
// @Summary Получение поста по ID
// @Tags         posts
// @produce json
// @Param id path integer true "ID искомого поста"
// @Success 200 {object} model.Post "объект с данными искомого поста"
// @Failure      400  {string}  string "неправильные параметры"
// @Failure      404  {string}  string "пост с таким ID не найден"
// @Failure      405  {string}  string "неподдерживаемый метод"
// @Failure      500  {string}  string "ошибка записи в кэш"
// @Router /api/posts/{id} [GET]
func (h *PostHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	// Проверяем метод запроса (должен быть GET)
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	//получаем id из пути.
	// Когда я обнаружил что chi  умеет это делать у меня уже была готовая
	//работающая функция которая его вытаскивала, переделывать не стал
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		sendErrorResponse(w, "The input value given is not a valid integer", http.StatusBadRequest)
		return
	}
	// Получаем пост через postService.GetByID
	postID := int(id)
	//пробуем забать пост из кэша
	post := new(model.Post)
	err = h.redis.HGetAll(r.Context(), strconv.Itoa(postID)).Scan(post)
	//если пост есть в кэше, то возвращаем его
	if err == nil && (*post != model.Post{}) {
		fmt.Printf("Отдали пост с id=%d из кэша\n", postID)
		sendJSONResponse(w, post, http.StatusOK)
		return
	}

	//если в кэше такого поста нет то берем из базы
	post, err = h.postService.GetByID(r.Context(), postID)
	if err != nil {
		if err == service.ErrPostNotFound {
			sendErrorResponse(w, err.Error(), http.StatusNotFound)
		} else {
			sendErrorResponse(w, err.Error(), http.StatusBadRequest)
		}
		return
	}
	err = post.ToRedisSet(r.Context(), h.redis, strconv.Itoa(postID))
	if err != nil {
		//беда случилась, редис отвалился
		sendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Printf("Положили в кэш пост с id=%d\n", postID)
	sendJSONResponse(w, post, http.StatusOK)
}

// GetAll возвращает список постов с пагинацией
// GET /api/posts?limit=10&offset=0
// Не требует аутентификации
// @Summary Получение списка постов с лимитом и пагинацией
// @Tags         posts
// @produce json
// @Param limit path integer true "количество постов на страницу"
// @Param offset path integer true "смещение от начала списка"
// @Success 200 {object} PostsResponse "объект со списком постов, общим количеством постов лимитом и смещением"
// @Failure      400  {string}  string "неправильные параметры"
// @Failure      405  {string}  string "неподдерживаемый метод"
// @Router /api/posts{limit}{offset} [GET]
func (h *PostHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	// Проверяем метод запроса (должен быть GET)
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// 2. Извлекаем параметры пагинации
	limit, err := strconv.ParseInt(r.URL.Query().Get("limit"), 10, 64)
	if err != nil {
		sendErrorResponse(w, "limit must be positive integer", http.StatusBadRequest)
		return
	}
	offset, err := strconv.ParseInt(r.URL.Query().Get("offset"), 10, 64)
	if err != nil {
		sendErrorResponse(w, "offset must be positive integer", http.StatusBadRequest)
		return
	}
	// Получаем посты через postService.GetAll
	posts, total, err := h.postService.GetAll(r.Context(), int(limit), int(offset))
	if err != nil {
		sendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}
	// Возвращаем список постов как JSON (200 OK)
	sendJSONResponse(w, PostsResponse{Posts: posts, Total: total, Limit: int(limit), Offset: int(offset)}, http.StatusOK)
}

// Update обновляет пост
// PUT /api/posts/{id}
// Требует аутентификации, может обновить только автор
// @Summary Изменение поста
// @Tags         posts
// @accept json
// @produce json
// @Param request body model.PostUpdateRequest true "ID искомого поста"
// @Param id path integer true "ID изменяемого поста"
// @Success 200 {object} model.Post "объект с данными измененного поста"
// @Failure      400  {string}  string "неправильные параметры"
// @Failure      403  {string}  string "попытка изменить чужой пост"
// @Failure      404  {string}  string "пост с таким ID не найден"
// @Failure      405  {string}  string "неподдерживаемый метод"
// @Failure      500  {string}  string "ошибка записи в кэш"
// @Router /api/posts/{id} [PUT]
func (h *PostHandler) Update(w http.ResponseWriter, r *http.Request) {
	// Проверяем метод запроса (должен быть PUT)
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// Получаем userID из контекста
	userID, ok := getUserIDFromContext(r.Context())
	if !ok {
		sendErrorResponse(w, service.ErrUserNotFound.Error(), http.StatusUnauthorized)
		return
	}
	// Извлекаем ID поста из URL
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		sendErrorResponse(w, "The input value given is not a valid integer", http.StatusBadRequest)
		return
	}
	postID := int(id)
	// postID := chi.URLParam(r, "id")
	// Декодируем JSON тело в PostUpdateRequest
	var postUpdateRequest *model.PostUpdateRequest
	err = parseJSONRequest(r, &postUpdateRequest)
	if err != nil {
		sendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}
	//проверяем полученную структуру на корректность
	err = h.validate.Struct(postUpdateRequest)
	if err != nil {
		sendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}
	// Обновляем через postService.Update
	post, err := h.postService.Update(r.Context(), postID, userID, postUpdateRequest)
	if err != nil {
		status := http.StatusBadRequest
		if err == service.ErrPostNotFound {
			status = http.StatusNotFound
		}
		if err == service.ErrForbidden {
			status = http.StatusForbidden
		}
		sendErrorResponse(w, err.Error(), status)
		return
	}
	//обновляем кэш
	err = post.ToRedisSet(r.Context(), h.redis, strconv.Itoa(postID))
	if err != nil {
		//беда случилась, редис отвалился
		sendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Printf("Обновили в кэш пост с id=%d\n", postID)

	// Возвращаем обновленный пост как JSON (200 OK)
	sendJSONResponse(w, post, http.StatusOK)
}

// Delete удаляет пост
// DELETE /api/posts/{id}
// Требует аутентификации, может удалить только автор
// @Summary Удаление поста
// @Tags         posts
// @Param id path integer true "ID удаляемого поста"
// @Success 204
// @Failure      400  {string}  string "неправильные параметры"
// @Failure      403  {string}  string "попытка удалить чужой пост"
// @Failure      404  {string}  string "пост с таким ID не найден"
// @Failure      405  {string}  string "неподдерживаемый метод"
// @Router /api/posts/{id} [DELETE]
func (h *PostHandler) Delete(w http.ResponseWriter, r *http.Request) {
	// Проверяем метод запроса (должен быть DELETE)
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// Получаем userID из контекста
	userID, ok := getUserIDFromContext(r.Context())
	if !ok {
		sendErrorResponse(w, service.ErrUserNotFound.Error(), http.StatusUnauthorized)
		return
	}
	// 3. Извлекаем ID поста из URL
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		sendErrorResponse(w, "The input value given is not a valid integer", http.StatusBadRequest)
		return
	}
	postID := int(id)
	// Удаляем через postService.Delete
	err = h.postService.Delete(r.Context(), postID, userID)
	if err != nil {
		status := http.StatusBadRequest
		if err == service.ErrPostNotFound {
			status = http.StatusNotFound
		}
		if err == service.ErrForbidden {
			status = http.StatusForbidden
		}
		sendErrorResponse(w, err.Error(), status)
		return
	}

	err = h.redis.Del(r.Context(), strconv.Itoa(postID)).Err()
	if err != nil {
		//беда случилась, редис отвалился
		// sendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		fmt.Printf("Ошибка при удалении из кеша поста с id %d:%s\n", postID, err.Error())
	} else {
		fmt.Printf("Положили в кэш пост с id=%d\n", postID)
	}

	sendJSONResponse(w, "", http.StatusNoContent)
}

// GetByAuthor возвращает посты конкретного автора
// GET /api/posts/author/{authorID}?limit=10&offset=0
// Не требует аутентификации
// @Summary Получение списка постов автора с лимитом и пагинацией
// @Tags         posts
// @produce json
// @Param authorID path integer true "ID автора"
// @Param limit path integer true "количество постов на страницу"
// @Param offset path integer true "смещение от начала списка"
// @Success 200 {object} PostsResponse "объект со списком постов, общим количеством постов лимитом и смещением"
// @Failure      400  {string}  string "неправильные параметры"
// @Failure      405  {string}  string "неподдерживаемый метод"
// @Router /api/posts/author/{authorID}{limit}{offset} [GET]
func (h *PostHandler) GetByAuthor(w http.ResponseWriter, r *http.Request) {
	// Проверяем метод запроса (должен быть GET)
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// Извлекаем параметры пагинации из query string
	limit, err := strconv.ParseInt(r.URL.Query().Get("limit"), 10, 64)
	if err != nil {
		sendErrorResponse(w, "limit must be positive integer", http.StatusBadRequest)
		return
	}
	offset, err := strconv.ParseInt(r.URL.Query().Get("offset"), 10, 64)
	if err != nil {
		sendErrorResponse(w, "offset must be positive integer", http.StatusBadRequest)
		return
	}
	// Извлекаем ID автора из URL
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		sendErrorResponse(w, "The input value given is not a valid integer", http.StatusBadRequest)
		return
	}
	authorID := int(id)
	// Получаем посты через postService.GetByAuthor
	posts, total, err := h.postService.GetByAuthor(r.Context(), authorID, int(limit), int(offset))
	if err != nil {
		sendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}
	sendJSONResponse(w, PostsResponse{Posts: posts, Total: total, Limit: int(limit), Offset: int(offset), AuthorID: authorID}, http.StatusOK)

}
