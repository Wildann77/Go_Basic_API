package services

import (
	"errors"
	"goapi/internal/models"
	"goapi/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"time"
)

type UserService interface {
	Register(req *models.RegisterRequest) (*models.UserResponse, error)
	Login(req *models.LoginRequest) (string, *models.UserResponse, error)
	GetByID(id uint) (*models.UserResponse, error)
	GetAll() ([]models.UserResponse, error)
	Update(id uint, updates *models.User) (*models.UserResponse, error)
	Delete(id uint) error
}

type userService struct {
	repo      repository.UserRepository
	jwtSecret string
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{
		repo:      repo,
		jwtSecret: "your-secret-key-change-in-production",
	}
}

func (s *userService) Register(req *models.RegisterRequest) (*models.UserResponse, error) {
	// Check if email exists
	if _, err := s.repo.GetByEmail(req.Email); err == nil {
		return nil, errors.New("email already registered")
	}

	user := &models.User{
		Email:    req.Email,
		Username: req.Username,
		Password: req.Password,
		FullName: req.FullName,
	}

	// Hash password
	if err := user.HashPassword(); err != nil {
		return nil, err
	}

	if err := s.repo.Create(user); err != nil {
		return nil, err
	}

	response := user.ToResponse()
	return &response, nil
}

func (s *userService) Login(req *models.LoginRequest) (string, *models.UserResponse, error) {
	user, err := s.repo.GetByEmail(req.Email)
	if err != nil {
		return "", nil, errors.New("invalid credentials")
	}

	if !user.CheckPassword(req.Password) {
		return "", nil, errors.New("invalid credentials")
	}

	// Generate JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  user.ID,
		"email":    user.Email,
		"role":     user.Role,
		"exp":      time.Now().Add(time.Hour * 24).Unix(), // 24 hours
	})

	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", nil, err
	}

	response := user.ToResponse()
	return tokenString, &response, nil
}

func (s *userService) GetByID(id uint) (*models.UserResponse, error) {
	user, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	response := user.ToResponse()
	return &response, nil
}

func (s *userService) GetAll() ([]models.UserResponse, error) {
	users, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}

	var responses []models.UserResponse
	for _, user := range users {
		responses = append(responses, user.ToResponse())
	}
	return responses, nil
}

func (s *userService) Update(id uint, updates *models.User) (*models.UserResponse, error) {
	user, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Update fields
	if updates.FullName != "" {
		user.FullName = updates.FullName
	}
	if updates.Username != "" {
		user.Username = updates.Username
	}

	if err := s.repo.Update(user); err != nil {
		return nil, err
	}

	response := user.ToResponse()
	return &response, nil
}

func (s *userService) Delete(id uint) error {
	return s.repo.Delete(id)
}