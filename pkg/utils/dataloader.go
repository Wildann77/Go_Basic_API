package utils

import (
	"context"
	"fmt"

	"goapi/internal/models"

	"github.com/graph-gophers/dataloader/v7"
)

type loaderContextKey string

const (
	// LoaderKey is the key used to store the dataloader in the context
	LoaderKey loaderContextKey = "loader_key"
)

// Loaders holds all dataloaders for the application
type Loaders struct {
	UserLoader *dataloader.Loader[uint, *models.User]
}

// GetLoadersFromContext retrieves the Loaders from the context
func GetLoadersFromContext(ctx context.Context) *Loaders {
	loaders, ok := ctx.Value(LoaderKey).(*Loaders)
	if !ok || loaders == nil {
		// Return nil if loaders not found in context
		return nil
	}
	return loaders
}

// NewLoaders creates a new instance of Loaders with configured dataloaders
func NewLoaders(
	userBatchFn func(ctx context.Context, keys []uint) []*dataloader.Result[*models.User],
) *Loaders {
	// Configure batch function for user loader
	userLoader := dataloader.NewBatchedLoader(
		userBatchFn,
		dataloader.WithBatchCapacity[uint, *models.User](100),
	)

	return &Loaders{
		UserLoader: userLoader,
	}
}

// LoadUser loads a single user by ID using the dataloader
func LoadUser(ctx context.Context, userID uint) (*models.User, error) {
	loaders := GetLoadersFromContext(ctx)
	if loaders == nil {
		return nil, fmt.Errorf("loaders not found in context")
	}

	thunk := loaders.UserLoader.Load(ctx, userID)
	return thunk()
}

// LoadUsers loads multiple users by IDs using the dataloader
func LoadUsers(ctx context.Context, userIDs []uint) ([]*models.User, []error) {
	loaders := GetLoadersFromContext(ctx)
	if loaders == nil {
		return nil, []error{fmt.Errorf("loaders not found in context")}
	}

	thunk := loaders.UserLoader.LoadMany(ctx, userIDs)
	return thunk()
}
