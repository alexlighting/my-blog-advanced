package service

import (
	"blog-api/internal/model"
	"blog-api/internal/repository"
	"blog-api/pkg/auth"
	"context"
	"errors"
	"fmt"
	"log"
)

var (
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
	ErrTokenSignFailed    = errors.New("failed to sign token")
)

type UserService struct {
	userRepo   repository.UserRepository
	jwtManager *auth.JWTManager
	logger     *log.Logger
}

func NewUserService(userRepo repository.UserRepository, jwtManager *auth.JWTManager, logger *log.Logger) *UserService {
	return &UserService{
		userRepo:   userRepo,
		jwtManager: jwtManager,
		logger:     logger,
	}
}

func (s *UserService) Register(ctx context.Context, req *model.UserCreateRequest) (*model.TokenResponse, error) {
	//проверяем используется-ли уже этот email
	exists, err := s.userRepo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("Email %s allready in use", req.Email)
	}
	//проверяем используется-ли уже этот username
	exists, err = s.userRepo.ExistsByUsername(ctx, req.Username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrUserAlreadyExists

	}
	//хэшируем пароль
	hpassword, err := auth.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}
	//создаем экземпляр для хранения данных пользователя
	user := model.User{Username: req.Username, Email: req.Email, Password: hpassword}
	err = s.userRepo.Create(ctx, &user)
	if err != nil {
		return nil, err
	}
	//генерируем токен
	token, expiresAt, err := s.jwtManager.GenerateToken(user.ID, user.Email, user.Username)
	if err != nil {
		return nil, ErrTokenSignFailed
	}
	//создаем и заполняем структуру TokenResponse, возвращаем ее
	return &model.TokenResponse{Token: token, ExpiresAt: expiresAt, User: *user.ToResponse()}, nil
}

func (s *UserService) Login(ctx context.Context, req *model.UserLoginRequest) (*model.TokenResponse, error) {

	user, err := s.GetByEmail(ctx, req.Email)
	if err != nil {
		//в логи пишем реальную причину отказа
		s.logger.Printf("User with email %s not exist\n", req.Email)
		//а пользователю пишем что пароль не совпал
		return nil, ErrInvalidCredentials
	}
	//проверяем пароль
	ok := auth.CheckPassword(req.Password, user.Password)
	if !ok {
		return nil, ErrInvalidCredentials
	}
	//генерируем токен
	token, expiresAt, err := s.jwtManager.GenerateToken(user.ID, user.Email, user.Username)
	if err != nil {
		return nil, ErrTokenSignFailed
	}
	//создаем и заполняем структуру TokenResponse, возвращаем ее
	return &model.TokenResponse{Token: token, ExpiresAt: expiresAt, User: *user.ToResponse()}, nil
}

func (s *UserService) GetByID(ctx context.Context, id int) (*model.User, error) {
	// TODO: Получить пользователя по ID через репозиторий
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (s *UserService) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	// TODO: Получить пользователя по email через репозиторий
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}
