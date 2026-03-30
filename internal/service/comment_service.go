package service

import (
	"blog-api/internal/model"
	"blog-api/internal/repository"
	"context"
	"errors"

	"fmt"
	"log"
	"strings"
	"time"
)

const (
	MIN_LIMIT  = 1
	MAX_LIMIT  = 100
	MIN_OFFSET = 0
)

var (
	ErrCommentNotFound = errors.New("comment not found")
	ErrPostNotExists   = errors.New("post does not exist")
	ErrDBError         = errors.New("db write error")
)

type CommentService struct {
	commentRepo repository.CommentRepository
	postRepo    repository.PostRepository
	userRepo    repository.UserRepository
	logger      *log.Logger
}

func NewCommentService(
	commentRepo repository.CommentRepository,
	postRepo repository.PostRepository,
	userRepo repository.UserRepository,
	logger *log.Logger,
) *CommentService {
	return &CommentService{
		commentRepo: commentRepo,
		postRepo:    postRepo,
		userRepo:    userRepo,
		logger:      logger,
	}
}

func (s *CommentService) Create(ctx context.Context, userID int, postID int, req *model.CommentCreateRequest) (*model.Comment, error) {
	//проверяем существует ли пост, к которому надо добавлять комментарий
	exists, err := s.postRepo.Exists(ctx, postID)
	if err != nil {
		s.logger.Printf("Failed check post existance exist %t, err %v\n", exists, err)
		return nil, err
	}
	if !exists {
		return nil, ErrPostNotExists
	}
	comment := model.Comment{Content: req.Content, PostID: postID, AuthorID: userID}
	err = s.commentRepo.Create(ctx, &comment)
	if err != nil {
		return nil, err
	}
	return &comment, nil
}

func (s *CommentService) GetByID(ctx context.Context, id int) (*model.Comment, error) {
	comment, err := s.commentRepo.GetByID(ctx, id)
	if err != nil {
		if err == ErrCommentNotFound {
			return nil, ErrCommentNotFound
		}
		if strings.Contains(err.Error(), "Error fetching comment with id =") {
			return nil, ErrDBError
		}
	}
	return comment, nil
}

func (s *CommentService) GetByPost(ctx context.Context, postID int, limit, offset int) ([]*model.Comment, int, error) {
	//проверяем limit
	if limit < MIN_LIMIT || limit > MAX_LIMIT {
		return nil, 0, fmt.Errorf("Limit must be in range [%d,%d]", MIN_LIMIT, MAX_LIMIT)
	}
	//проверяем offset
	if offset < MIN_OFFSET {
		return nil, 0, fmt.Errorf("Offset can't be negative")
	}
	//проверяем существует ли пост
	exist, err := s.postRepo.Exists(ctx, postID)
	if err != nil {
		return nil, 0, err
	}
	if !exist {
		return nil, 0, ErrPostNotExists
	}
	//если существует получаем общее количество комментариев
	total, err := s.commentRepo.GetCountByPostID(ctx, postID)
	if err != nil {
		return nil, 0, err
	}
	//и сами комментарии с учетом limit и offset
	comments, err := s.commentRepo.GetByPostID(ctx, postID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	return comments, total, nil
}

func (s *CommentService) Update(ctx context.Context, id int, userID int, req *model.CommentUpdateRequest) (*model.Comment, error) {
	//получаем комментарий по ID
	comment, err := s.commentRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	//проверяем можно ли пользователю его изменять
	if !comment.CanBeEditedBy(userID) {
		return nil, ErrForbidden
	}
	//и обновляем комсментарий
	comment.Content = req.Content
	comment.UpdatedAt = time.Now()
	err = s.commentRepo.Update(ctx, comment)
	if err != nil {
		return nil, err
	}
	return comment, nil
}

func (s *CommentService) Delete(ctx context.Context, id int, userID int) error {
	//проверяем что комментарий существует
	exist, err := s.commentRepo.Exists(ctx, id)
	if err != nil {
		return err
	}
	if !exist {
		return ErrCommentNotFound
	}
	//получаем комментарий
	comment, err := s.commentRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	//проверяем права пользователя
	if !comment.CanBeDeletedBy(userID) {
		return ErrForbidden
	}
	err = s.commentRepo.Delete(ctx, id)
	if err != nil {
		return err
	}
	return nil
}

func (s *CommentService) GetByAuthor(ctx context.Context, authorID int, limit, offset int) ([]*model.Comment, int, error) {
	if limit < MIN_LIMIT || limit > MAX_LIMIT {
		return nil, 0, fmt.Errorf("Limit must be in range [%d,%d]", MIN_LIMIT, MAX_LIMIT)
	}
	//проверяем offset
	if offset < 0 {
		return nil, 0, fmt.Errorf("Offset can't be negative")
	}
	total, err := s.commentRepo.GetCountByAuthorID(ctx, authorID)
	if err != nil {
		return nil, 0, err
	}
	comments, err := s.commentRepo.GetByPostID(ctx, authorID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	return comments, total, nil
}
