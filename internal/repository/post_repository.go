package repository

import (
	"context"
	"errors"

	"goapi/internal/models"
	"goapi/pkg/utils"

	"gorm.io/gorm"
)

type PostRepository interface {
	Create(ctx context.Context, post *models.Post) error
	GetByID(ctx context.Context, id uint) (*models.Post, error)
	GetAll(ctx context.Context) ([]models.Post, error)
	GetByUserID(ctx context.Context, userID uint) ([]models.Post, error)
	Update(ctx context.Context, post *models.Post) error
	Delete(ctx context.Context, id uint) error
}

type postRepository struct {
	db *gorm.DB
}

func NewPostRepository(db *gorm.DB) PostRepository {
	return &postRepository{db: db}
}

func (r *postRepository) Create(ctx context.Context, post *models.Post) error {
	db := utils.GetDBFromContext(ctx, r.db)
	return db.Create(post).Error
}

func (r *postRepository) GetByID(ctx context.Context, id uint) (*models.Post, error) {
	db := utils.GetDBFromContext(ctx, r.db)
	var post models.Post
	if err := db.First(&post, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("post not found")
		}
		return nil, err
	}
	return &post, nil
}

func (r *postRepository) GetAll(ctx context.Context) ([]models.Post, error) {
	db := utils.GetDBFromContext(ctx, r.db)
	var posts []models.Post
	// Without Preload - this is where N+1 would happen if we load users individually
	if err := db.Order("created_at DESC").Find(&posts).Error; err != nil {
		return nil, err
	}
	return posts, nil
}

func (r *postRepository) GetByUserID(ctx context.Context, userID uint) ([]models.Post, error) {
	db := utils.GetDBFromContext(ctx, r.db)
	var posts []models.Post
	if err := db.Where("user_id = ?", userID).Order("created_at DESC").Find(&posts).Error; err != nil {
		return nil, err
	}
	return posts, nil
}

func (r *postRepository) Update(ctx context.Context, post *models.Post) error {
	db := utils.GetDBFromContext(ctx, r.db)
	return db.Save(post).Error
}

func (r *postRepository) Delete(ctx context.Context, id uint) error {
	db := utils.GetDBFromContext(ctx, r.db)
	return db.Delete(&models.Post{}, id).Error
}
