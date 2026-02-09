package services

import (
	"context"
	"errors"
	"goapi/internal/models"
	"goapi/internal/repository"
	"time"

	"encoding/json"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
)

type UserService interface {
	Register(ctx context.Context, req *models.RegisterRequest) (*models.UserResponse, error)
	Login(ctx context.Context, req *models.LoginRequest) (string, *models.UserResponse, error)
	GetByID(ctx context.Context, id uint) (*models.UserResponse, error)
	GetAll(ctx context.Context) ([]models.UserResponse, error)
	Update(ctx context.Context, id uint, updates *models.User) (*models.UserResponse, error)
	Delete(ctx context.Context, id uint) error
}

type userService struct {
	repo      repository.UserRepository
	redis     *redis.Client
	jwtSecret string
}

func NewUserService(repo repository.UserRepository, redisClient *redis.Client) UserService {
	return &userService{
		repo:      repo,
		redis:     redisClient,
		jwtSecret: "your-secret-key-change-in-production",
	}
}

func (s *userService) Register(ctx context.Context, req *models.RegisterRequest) (*models.UserResponse, error) {
	var response models.UserResponse

	err := s.repo.WithTransaction(ctx, func(txCtx context.Context) error {
		// Check if email exists
		if _, err := s.repo.GetByEmail(txCtx, req.Email); err == nil {
			return errors.New("email already registered")
		}

		user := &models.User{
			Email:    req.Email,
			Username: req.Username,
			Password: req.Password,
			FullName: req.FullName,
		}

		// Hash password
		if err := user.HashPassword(); err != nil {
			return err
		}

		if err := s.repo.Create(txCtx, user); err != nil {
			return err
		}

		response = user.ToResponse()
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &response, nil
}

func (s *userService) Login(ctx context.Context, req *models.LoginRequest) (string, *models.UserResponse, error) {
	user, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		return "", nil, errors.New("invalid credentials")
	}

	if !user.CheckPassword(req.Password) {
		return "", nil, errors.New("invalid credentials")
	}

	// Generate JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"role":    user.Role,
		"exp":     time.Now().Add(time.Hour * 24).Unix(), // 24 hours
	})

	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", nil, err
	}

	response := user.ToResponse()
	return tokenString, &response, nil
}

func (s *userService) GetByID(ctx context.Context, id uint) (*models.UserResponse, error) {
	cacheKey := fmt.Sprintf("user:%d", id)

	// 1. Try Cache
	val, err := s.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var cachedUser models.UserResponse
		if err := json.Unmarshal([]byte(val), &cachedUser); err == nil {
			return &cachedUser, nil
		}
	}

	// 2. Cache Miss - Query DB
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	response := user.ToResponse()

	// 3. Set Cache (TTL 10 mins)
	if data, err := json.Marshal(response); err == nil {
		s.redis.Set(ctx, cacheKey, data, 10*time.Minute)
	}

	return &response, nil
}

func (s *userService) GetAll(ctx context.Context) ([]models.UserResponse, error) {
	users, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	var responses []models.UserResponse
	for _, user := range users {
		responses = append(responses, user.ToResponse())
	}
	return responses, nil
}

func (s *userService) Update(ctx context.Context, id uint, updates *models.User) (*models.UserResponse, error) {
	// Start a transaction for update (even though it's single record, good practice)
	var response models.UserResponse
	err := s.repo.WithTransaction(ctx, func(txCtx context.Context) error {
		user, err := s.repo.GetByID(txCtx, id)
		if err != nil {
			return err
		}

		// Update fields
		if updates.FullName != "" {
			user.FullName = updates.FullName
		}
		if updates.Username != "" {
			user.Username = updates.Username
		}

		if err := s.repo.Update(txCtx, user); err != nil {
			return err
		}

		// Invalidate cache
		cacheKey := fmt.Sprintf("user:%d", id)
		s.redis.Del(ctx, cacheKey)

		response = user.ToResponse()
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &response, nil
}

func (s *userService) Delete(ctx context.Context, id uint) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	// Invalidate cache
	return s.redis.Del(ctx, fmt.Sprintf("user:%d", id)).Err()
}
