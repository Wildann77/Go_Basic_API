package handlers

import (
	"net/http"
	"strconv"

	"goapi/internal/models"
	"goapi/internal/services"
	"goapi/pkg/utils"

	"github.com/gin-gonic/gin"
)

type PostHandler struct {
	service services.PostService
}

func NewPostHandler(service services.PostService) *PostHandler {
	return &PostHandler{service: service}
}

// CreatePost creates a new post
func (h *PostHandler) CreatePost(c *gin.Context) {
	var req models.CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// Get user ID from JWT claims
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", "user not authenticated")
		return
	}

	post, err := h.service.Create(c.Request.Context(), &req, userID.(uint))
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create post", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Post created successfully", post)
}

// GetPost retrieves a single post by ID
func (h *PostHandler) GetPost(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid post ID", err.Error())
		return
	}

	post, err := h.service.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Post not found", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Post retrieved successfully", post)
}

// GetAllPosts retrieves all posts (demonstrates DataLoader batching)
// Supports optional ?user_id=X query parameter to filter by user
func (h *PostHandler) GetAllPosts(c *gin.Context) {
	// Check if filtering by user_id
	userIDParam := c.Query("user_id")
	if userIDParam != "" {
		userID, err := strconv.ParseUint(userIDParam, 10, 32)
		if err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID", err.Error())
			return
		}

		posts, err := h.service.GetByUserID(c.Request.Context(), uint(userID))
		if err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve posts", err.Error())
			return
		}

		utils.SuccessResponse(c, http.StatusOK, "Posts retrieved successfully", posts)
		return
	}

	// Get all posts
	posts, err := h.service.GetAll(c.Request.Context())
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve posts", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Posts retrieved successfully", posts)
}

// DeletePost deletes a post (only by owner)
func (h *PostHandler) DeletePost(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid post ID", err.Error())
		return
	}

	// Get user ID from JWT claims
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", "user not authenticated")
		return
	}

	if err := h.service.Delete(c.Request.Context(), uint(id), userID.(uint)); err != nil {
		utils.ErrorResponse(c, http.StatusForbidden, "Failed to delete post", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Post deleted successfully", nil)
}
