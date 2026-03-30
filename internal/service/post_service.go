package service

import (
	"blog-api/internal/model"
	"blog-api/internal/repository"
	"context"
	"errors"
	"fmt"
	"time"
)

var (
	ErrPostNotFound = errors.New("post not found")
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
)

type PostService struct {
	postRepo repository.PostRepository
	userRepo repository.UserRepository
}

func NewPostService(postRepo repository.PostRepository, userRepo repository.UserRepository) *PostService {
	return &PostService{
		postRepo: postRepo,
		userRepo: userRepo,
	}
}

func (s *PostService) Create(ctx context.Context, userID int, req *model.PostCreateRequest) (*model.Post, error) {
	//проверяем время публикации, если оно в будущем то помечаем пост как черновик который в последующем
	// будет автоматически опубликован.
	draft := false
	if req.Publish_at.After(time.Now()) {
		draft = true
	} else {
		//иначе публикуем пост сразу
		req.Publish_at = time.Now()
	}

	post := model.Post{Title: req.Title, Content: req.Content, AuthorID: userID, Draft: draft, Publish_at: req.Publish_at}
	err := s.postRepo.Create(ctx, &post)
	if err != nil {
		return nil, err
	}
	return &post, nil
}

func (s *PostService) GetByID(ctx context.Context, id int) (*model.Post, error) {
	post, err := s.postRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return post, nil
}

func (s *PostService) GetAll(ctx context.Context, limit, offset int) ([]*model.Post, int, error) {
	//проверяем limit
	if limit < MIN_LIMIT || limit > MAX_LIMIT {
		return nil, 0, fmt.Errorf("Limit must be in range [%d,%d]", MIN_LIMIT, MAX_LIMIT)
	}
	//проверяем offset
	if offset < MIN_OFFSET {
		return nil, 0, fmt.Errorf("Offset can't be negative")
	}
	//получаем общее количество постов
	total, err := s.postRepo.GetTotalCount(ctx)
	if err != nil {
		return nil, 0, err
	}
	//получаем посты
	posts, err := s.postRepo.GetAll(ctx, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	//возвращаем полученные данные
	return posts, total, nil
}

func (s *PostService) Update(ctx context.Context, id int, userID int, req *model.PostUpdateRequest) (*model.Post, error) {

	post, err := s.postRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrPostNotFound
	}
	//проверяем автора
	if !post.CanBeEditedBy(userID) {
		return nil, ErrForbidden
	}
	//проверяем время публикации, если оно в будущем то помечаем пост как черновик который в последующем
	// будет автоматически опубликован.
	draft := false
	if req.Publish_at.After(time.Now()) {
		draft = true
	} else {
		//иначе публикуем пост сразу
		req.Publish_at = time.Now()
	}
	//заменяем измененные поля в post
	post.Draft = draft
	post.Publish_at = req.Publish_at

	if post.Title != req.Title {
		post.Title = req.Title
	}
	if post.Content != req.Content {
		post.Content = req.Content
	}
	//сохраняем их в базу
	err = s.postRepo.Update(ctx, post)
	if err != nil {
		return nil, err
	}
	//возвращаем результат
	return post, nil
}

func (s *PostService) Delete(ctx context.Context, id int, userID int) error {
	//проверяем существование поста
	ok, err := s.postRepo.Exists(ctx, id)
	if err != nil {
		return err
	}
	if !ok {
		return ErrPostNotFound
	}
	//получаем данные поста
	post, err := s.postRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("Can't fetch post with id %d", id)
	}

	//проверяем автора поста
	if !post.CanBeDeletedBy(userID) {
		return ErrForbidden
	}
	//пытаемся удалить пост
	err = s.postRepo.Delete(ctx, id)
	if err != nil {
		return err
	}
	return nil
}

func (s *PostService) GetByAuthor(ctx context.Context, authorID int, limit, offset int) ([]*model.Post, int, error) {
	//проверяем limit
	if limit < MIN_LIMIT || limit > MAX_LIMIT {
		return nil, 0, fmt.Errorf("Limit must be in range [%d,%d]", MIN_LIMIT, MAX_LIMIT)
	}
	//проверяем offset
	if offset < MIN_OFFSET {
		return nil, 0, fmt.Errorf("Offset can't be negative")
	}
	//получаем общее количество постов
	total, err := s.postRepo.GetTotalCountByAuthor(ctx, authorID)
	if err != nil {
		return nil, 0, err
	}
	//получаем посты
	posts, err := s.postRepo.GetByAuthorID(ctx, authorID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	//возвращаем полученные данные
	return posts, total, nil
}
