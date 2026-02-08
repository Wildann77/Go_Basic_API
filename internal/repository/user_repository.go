package repository

import (
	"context"
	"errors"
	"goapi/internal/models"
	"goapi/pkg/utils"

	"gorm.io/gorm"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id uint) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetAll(ctx context.Context) ([]models.User, error)
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id uint) error
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return utils.RunInTransaction(ctx, r.db, fn)
}

func (r *userRepository) Create(ctx context.Context, user *models.User) error {
	db := utils.GetDBFromContext(ctx, r.db)
	return db.Create(user).Error
}

func (r *userRepository) GetByID(ctx context.Context, id uint) (*models.User, error) {
	db := utils.GetDBFromContext(ctx, r.db)
	var user models.User
	if err := db.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	db := utils.GetDBFromContext(ctx, r.db)
	var user models.User
	if err := db.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetAll(ctx context.Context) ([]models.User, error) {
	db := utils.GetDBFromContext(ctx, r.db)
	var users []models.User
	if err := db.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (r *userRepository) Update(ctx context.Context, user *models.User) error {
	db := utils.GetDBFromContext(ctx, r.db)
	return db.Save(user).Error
}

func (r *userRepository) Delete(ctx context.Context, id uint) error {
	db := utils.GetDBFromContext(ctx, r.db)
	return db.Delete(&models.User{}, id).Error
}
