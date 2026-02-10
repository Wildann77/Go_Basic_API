package services

import (
	"context"
	"errors"

	"goapi/internal/models"
	"goapi/internal/repository"
	"goapi/pkg/logger"
	"goapi/pkg/utils"
)

type PostService interface {
	Create(ctx context.Context, req *models.CreatePostRequest, userID uint) (*models.PostResponse, error)
	GetByID(ctx context.Context, id uint) (*models.PostResponse, error)
	GetAll(ctx context.Context) ([]models.PostResponse, error)
	GetByUserID(ctx context.Context, userID uint) ([]models.PostResponse, error)
	Delete(ctx context.Context, id uint, userID uint) error
}

type postService struct {
	repo repository.PostRepository
}

func NewPostService(repo repository.PostRepository) PostService {
	return &postService{repo: repo}
}

func (s *postService) Create(ctx context.Context, req *models.CreatePostRequest, userID uint) (*models.PostResponse, error) {
	post := &models.Post{
		Title:   req.Title,
		Content: req.Content,
		UserID:  userID,
	}

	if err := s.repo.Create(ctx, post); err != nil {
		logger.WithContext(ctx).Error("Failed to create post", "error", err)
		return nil, err
	}

	// Load author using DataLoader
	user, err := utils.LoadUser(ctx, post.UserID)
	if err != nil {
		logger.WithContext(ctx).Warn("Failed to load post author", "user_id", post.UserID, "error", err)
	}

	post.User = user
	response := post.ToResponse()
	return &response, nil
}

func (s *postService) GetByID(ctx context.Context, id uint) (*models.PostResponse, error) {
	post, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Load author using DataLoader to avoid N+1
	user, err := utils.LoadUser(ctx, post.UserID)
	if err != nil {
		logger.WithContext(ctx).Warn("Failed to load post author", "user_id", post.UserID, "error", err)
	}

	post.User = user
	response := post.ToResponse()
	return &response, nil
}

func (s *postService) GetAll(ctx context.Context) ([]models.PostResponse, error) {
	posts, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	// Collect all user IDs
	userIDs := make([]uint, 0, len(posts))
	for _, post := range posts {
		userIDs = append(userIDs, post.UserID)
	}

	// Batch load all users at once using DataLoader (solves N+1 problem)
	users, errs := utils.LoadUsers(ctx, userIDs)

	// Create a map for quick lookup
	userMap := make(map[uint]*models.User)
	for i, user := range users {
		if errs[i] == nil && user != nil {
			userMap[userIDs[i]] = user
		}
	}

	// Build responses with loaded users
	responses := make([]models.PostResponse, len(posts))
	for i, post := range posts {
		post.User = userMap[post.UserID]
		responses[i] = post.ToResponse()
	}

	return responses, nil
}

func (s *postService) GetByUserID(ctx context.Context, userID uint) ([]models.PostResponse, error) {
	posts, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Load author once using DataLoader
	user, err := utils.LoadUser(ctx, userID)
	if err != nil {
		logger.WithContext(ctx).Warn("Failed to load post author", "user_id", userID, "error", err)
	}

	// Build responses
	responses := make([]models.PostResponse, len(posts))
	for i, post := range posts {
		post.User = user
		responses[i] = post.ToResponse()
	}

	return responses, nil
}

func (s *postService) Delete(ctx context.Context, id uint, userID uint) error {
	post, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Check ownership
	if post.UserID != userID {
		return errors.New("unauthorized to delete this post")
	}

	return s.repo.Delete(ctx, id)
}
