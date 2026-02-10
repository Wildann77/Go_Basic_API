package middleware

import (
	"context"

	"goapi/internal/models"
	"goapi/internal/repository"
	"goapi/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/graph-gophers/dataloader/v7"
)

// DataLoaderMiddleware creates request-scoped dataloaders
func DataLoaderMiddleware(userRepo repository.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Create batch function for users
		userBatchFn := func(ctx context.Context, keys []uint) []*dataloader.Result[*models.User] {
			// Fetch users from repository in a single query
			userMap, err := userRepo.GetUsersByIDs(ctx, keys)

			// Build results array preserving order
			results := make([]*dataloader.Result[*models.User], len(keys))
			for i, key := range keys {
				if err != nil {
					results[i] = &dataloader.Result[*models.User]{Error: err}
					continue
				}

				user, found := userMap[key]
				if !found {
					results[i] = &dataloader.Result[*models.User]{Error: nil, Data: nil}
				} else {
					results[i] = &dataloader.Result[*models.User]{Data: user}
				}
			}

			return results
		}

		// Create loaders instance
		loaders := utils.NewLoaders(userBatchFn)

		// Store loaders in context
		ctx := context.WithValue(c.Request.Context(), utils.LoaderKey, loaders)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}
