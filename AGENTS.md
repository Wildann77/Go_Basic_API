# AGENTS.md - Coding Agent Instructions

This is a **Go REST API** project using Clean Architecture. It uses Gin framework, GORM with PostgreSQL, and JWT authentication.

## Project Structure

```
cmd/api/          # Application entry point
internal/
  config/         # Configuration and DB initialization
  handlers/       # HTTP handlers (controllers)
  services/       # Business logic layer
  repository/     # Data access layer
  models/         # Data models and DTOs
  middleware/     # HTTP middleware
pkg/
  utils/          # Utility functions
  logger/         # Structured logger (slog)
```

**Architecture Pattern**: Clean Architecture with dependency injection
- Handlers depend on Service interfaces
- Services depend on Repository interfaces
- Dependencies are wired in `cmd/api/main.go`

## Build Commands

```bash
# Development (hot reload with Air)
make dev

# Run directly
make run

# Build binary
make build

# Download dependencies
make deps

# Full setup (dependencies + infrastructure)
make setup
```

## Test Commands

```bash
# Run all tests
make test

# Run tests in specific package
go test -v ./internal/services/...

# Run single test function
go test -v ./internal/services/... -run TestFunctionName

# Run with coverage
go test -v -cover ./...
```

## Infrastructure Commands

```bash
# Start PostgreSQL and Redis (Docker)
make up

# Stop services
make down

# Check service status
make status

# View logs
make logs

# Run database migrations
make migrate-up
make migrate-down
```

## Code Style Guidelines

### Imports
- Group imports: stdlib → internal → external
- Separate groups with blank lines

```go
import (
    "net/http"
    "time"

    "goapi/internal/models"
    "goapi/pkg/utils"

    "github.com/gin-gonic/gin"
)
```

### Naming Conventions
- **Interfaces**: `UserService`, `UserRepository` (exported)
- **Implementations**: `userService`, `userRepository` (unexported struct)
- **Constructors**: `NewUserService()`, `NewUserRepository()`
- **Handlers**: Method receivers use `h *UserHandler`
- **Services**: Method receivers use `s *userService`
- **Repositories**: Method receivers use `r *userRepository`
- **JSON fields**: Use `snake_case` (e.g., `full_name`)
- **DB columns**: Use `snake_case` in GORM tags

### Context Usage
- **Standard**: Always pass `context.Context` as the first parameter in Service and Repository layers.
- **Purpose**: Enables timeouts, cancellation, and transaction propagation.
- **Format**: `func (s *userService) Register(ctx context.Context, req *models.RegisterRequest)`

### Types and Structs
- Define interfaces for services and repositories
- Use struct tags for JSON and GORM
- Create separate Request/Response types for API contracts

```go
type User struct {
    ID        uint           `json:"id" gorm:"primaryKey"`
    Email     string         `json:"email" gorm:"uniqueIndex;not null"`
    Password  string         `json:"-" gorm:"not null"` // Hide sensitive fields
    FullName  string         `json:"full_name"`
    CreatedAt time.Time      `json:"created_at"`
}
```

### Error Handling
- Return errors from Service and Repository layers.
- **Handlers** use `utils.ErrorResponse()` for consistent error format.
- `utils.ErrorResponse` automatically logs the error to the structured logger via Gin context.
- Check `gorm.ErrRecordNotFound` for database operations.

```go
func (r *userRepository) GetByID(ctx context.Context, id uint) (*models.User, error) {
    var user models.User
    if err := utils.GetDBFromContext(ctx, r.db).First(&user, id).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, errors.New("user not found")
        }
        return nil, err
    }
    return &user, nil
}
```

### Response Format
Always use utility functions for responses:

```go
utils.SuccessResponse(c, http.StatusOK, "Message", data)
utils.ErrorResponse(c, http.StatusBadRequest, "Message", err.Error())
```

## Database Transactions (ACID)

To maintain **ACID** properties across multiple operations, transactions must be managed at the **Service Layer** to ensure business logic atomicity.

### 1. The Pattern: Transaction via Context
The transaction object (`*gorm.DB`) should be passed through `context.Context`. This keeps the Repository interfaces clean while allowing transaction propagation.

### 2. Implementation in Repository
Repositories should be "context-aware" by using a helper to extract the active transaction from the context.

```go
func (r *userRepository) Create(ctx context.Context, user *models.User) error {
    // GetDB returns tx from context or the default DB
    db := utils.GetDBFromContext(ctx, r.db) 
    return db.Create(user).Error
}
```

### 3. Usage in Service
Services orchestrate the transaction using a closure-based wrapper.

```go
func (s *userService) RegisterWithProfile(ctx context.Context, req *models.RegisterRequest) error {
    return s.repo.WithTransaction(ctx, func(txCtx context.Context) error {
        // Step 1: Create User
        if err := s.repo.Create(txCtx, user); err != nil {
            return err // Automatically triggers Rollback
        }

        // Step 2: Create Initial Profile
        if err := s.profileRepo.Create(txCtx, profile); err != nil {
            return err // Automatically triggers Rollback
        }

        return nil // Automatically triggers Commit
    })
}
```

### 4. ACID Alignment
- **Atomicity**: The `WithTransaction` wrapper ensures all-or-nothing execution.
- **Consistency**: Database constraints and GORM hooks are fully respected.
- **Isolation**: Defaults to PostgreSQL 'Read Committed'. Higher levels can be set per transaction.
- **Durability**: Managed by PostgreSQL's write-ahead logging (WAL).

```go
func (s *userService) RegisterWithProfile(ctx context.Context, req *models.RegisterRequest) error {
    return s.repo.WithTransaction(ctx, func(txCtx context.Context) error {
        // Step 1: Create User
        if err := s.repo.Create(txCtx, user); err != nil {
            return err // Automatically triggers Rollback
        }

        // Step 2: Create Initial Profile
        if err := s.profileRepo.Create(txCtx, profile); err != nil {
            return err // Automatically triggers Rollback
        }

        return nil // Automatically triggers Commit
    })
}
```

## Rate Limiting

Implement **Rate Limiting** to protect the API from brute-force attacks and abuse. Use a distributed approach with **Redis**.

### 1. The Strategy: Redis-Backed Limiting
Use `ulule/limiter` for a robust, distributed rate limiting solution. This ensures limits are consistent across multiple server instances.

### 2. Implementation Pattern
Define the rate limiter in `internal/middleware` and initialize it with Redis store.

```go
func RateLimiter(redisClient *redis.Client) gin.HandlerFunc {
    // 1. Define rate (e.g., 5 requests per second)
    rate := limiter.Rate{
        Period: 1 * time.Second,
        Limit:  5,
    }

    // 2. Create Redis store
    store, _ := redisstore.NewStore(redisClient)

    // 3. Create limiter instance
    instance := limiter.New(store, rate)

    return func(c *gin.Context) {
        key := c.ClientIP() // Use IP for public or user_id for protected
        
        context, err := instance.Get(c, key)
        if err != nil {
            c.AbortWithStatusJSON(500, gin.H{"error": "limiter error"})
            return
        }

        if context.Reached {
            c.AbortWithStatusJSON(429, gin.H{"error": "too many requests"})
            return
        }

        c.Next()
    }
}
```

### 3. Application
- **Global**: Apply to `router.Use()` for general protection.
- **Route-specific**: Apply to sensitive routes like `/login` or `/register` with stricter limits.

## Redis Caching

Use **Redis** for caching expensive database queries or frequently accessed data using the **Cache-Aside** pattern.

### 1. The Strategy: Cache-Aside
1. Check if data exists in Redis.
2. If found (**Cache Hit**), return the data immediately.
3. If not found (**Cache Miss**), query the Database.
4. Store the result in Redis with a **TTL (Time To Live)** and return it.

### 2. Implementation Pattern
Caching should be handled in the **Service Layer** to keep the Repository clean and allow business logic to decide when to cache.

```go
func (s *userService) GetByID(ctx context.Context, id uint) (*models.User, error) {
    cacheKey := fmt.Sprintf("user:%d", id)

    // 1. Try to get from Cache
    cachedData, err := s.redis.Get(ctx, cacheKey).Result()
    if err == nil {
        var user models.User
        if json.Unmarshal([]byte(cachedData), &user) == nil {
            return &user, nil
        }
    }

    // 2. Cache Miss - Get from DB
    user, err := s.repo.GetByID(id)
    if err != nil {
        return nil, err
    }

    // 3. Store in Cache (with TTL, e.g., 10 minutes)
    jsonData, _ := json.Marshal(user)
    s.redis.Set(ctx, cacheKey, jsonData, 10*time.Minute)

    return user, nil
}
```

### 3. Data Invalidation
Always invalidate (delete) the cache when data is updated or deleted to maintain consistency.

```go
func (s *userService) Update(ctx context.Context, user *models.User) error {
    if err := s.repo.Update(user); err != nil {
        return err
    }
    
    // Invalidate cache
    cacheKey := fmt.Sprintf("user:%d", user.ID)
    s.redis.Del(ctx, cacheKey)
    return nil
}
```

### Service Interface Pattern
```go
type UserService interface {
    Register(req *models.RegisterRequest) (*models.UserResponse, error)
    GetByID(id uint) (*models.UserResponse, error)
}

type userService struct {
    repo      repository.UserRepository
    jwtSecret string
}

func NewUserService(repo repository.UserRepository) UserService {
    return &userService{repo: repo, jwtSecret: "key"}
}
```

### Handler Pattern
```go
type UserHandler struct {
    service services.UserService
}

func NewUserHandler(service services.UserService) *UserHandler {
    return &UserHandler{service: service}
}

func (h *UserHandler) GetUser(c *gin.Context) {
    // Implementation
}
```

## Database Indexes

Optimize query performance by implementing strategic **Database Indexes**. In this project, we use **GORM tags** to manage indexes directly in the models.

### 1. The Strategy: Query-Driven Indexing
Apply indexes based on common query patterns:
- **Unique Indexes**: For fields that must be unique (e.g., `email`, `username`).
- **Single Column Indexes**: For fields used frequently in `WHERE` clauses (e.g., `status`, `active`).
- **Composite Indexes**: For queries involving multiple columns (e.g., `WHERE user_id = ? AND status = ?`).
- **Partial/Conditional Indexes**: For indexing a subset of rows (PostgreSQL specific).

### 2. Implementation Pattern
Define indexes using `gorm` tags in the **Model Structs**.

```go
type User struct {
    ID        uint      `gorm:"primaryKey"`
    Email     string    `gorm:"uniqueIndex;not null"` // Unique Index
    Username  string    `gorm:"uniqueIndex;not null"` // Unique Index
    Status    string    `gorm:"index"`                // Regular Index
    CreatedAt time.Time `gorm:"index:,sort:desc"`     // Index with sorting
}
```

### 3. Composite Indexes
For queries that filter by multiple fields, use named indexes to group columns.

```go
type AuditLog struct {
    ID        uint   `gorm:"primaryKey"`
    UserID    uint   `gorm:"index:idx_user_action"` // Part of composite index
    Action    string `gorm:"index:idx_user_action"` // Part of composite index
    CreatedAt time.Time
}
```

### 4. Indexing Best Practices
1.  **Don't Over-Index**: Every index adds overhead to `INSERT`, `UPDATE`, and `DELETE` operations.
2.  **Index Selective Columns**: Avoid indexing columns with low cardinality (e.g., Boolean fields like `is_deleted` unless part of a composite index).
3.  **Order Matters in Composite Indexes**: Place the most selective column (the one that filters out the most rows) first.
## N+1 Problem Resolution (DataLoader)

Efficiently resolve the **N+1 query problem** by batching and caching database requests. While GORM's `Preload` is suitable for simple cases, **DataLoader** is preferred for complex, nested, or dynamic relationships.

### 1. The Strategy: Batch Loading
- **Batching**: Instead of executing $N$ queries for $N$ records, the application collects IDs and executes **one** query (e.g., `WHERE id IN (...)`).
- **Caching**: Results are cached within the scope of a single request to avoid redundant lookups.
- **Isolation**: DataLoaders are request-scoped to prevent data leakage between different users/requests.

### 2. Implementation Pattern (DataLoader)
The implementation uses `github.com/graph-gophers/dataloader/v7` and is split across multiple layers.

#### Repository Layer (Batch Fetcher)
Add a batch method to your repository that fetches multiple records in one query:

```go
func (r *userRepository) GetUsersByIDs(ctx context.Context, ids []uint) (map[uint]*models.User, error) {
    db := utils.GetDBFromContext(ctx, r.db)
    
    var users []models.User
    if err := db.Where("id IN ?", ids).Find(&users).Error; err != nil {
        return nil, err
    }
    
    // Map users by ID to preserve order and handle missing records
    userMap := make(map[uint]*models.User, len(users))
    for i := range users {
        userMap[users[i].ID] = &users[i]
    }
    
    return userMap, nil
}
```

#### Middleware Layer (Request Scoping)
Create a middleware that initializes DataLoaders for each request:

```go
func DataLoaderMiddleware(userRepo repository.UserRepository) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Create batch function for users
        userBatchFn := func(ctx context.Context, keys []uint) []*dataloader.Result[*models.User] {
            userMap, err := userRepo.GetUsersByIDs(ctx, keys)
            
            results := make([]*dataloader.Result[*models.User], len(keys))
            for i, key := range keys {
                if err != nil {
                    results[i] = &dataloader.Result[*models.User]{Error: err}
                    continue
                }
                
                user, found := userMap[key]
                if !found {
                    results[i] = &dataloader.Result[*models.User]{Data: nil}
                } else {
                    results[i] = &dataloader.Result[*models.User]{Data: user}
                }
            }
            return results
        }

        loaders := utils.NewLoaders(userBatchFn)
        ctx := context.WithValue(c.Request.Context(), utils.LoaderKey, loaders)
        c.Request = c.Request.WithContext(ctx)
        c.Next()
    }
}
```

### 3. Usage in Services
Services use the loader to resolve dependencies lazily and efficiently:

```go
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

    // Batch load all users at once (solves N+1 problem)
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
```

### 4. Integration in Main
Register the middleware globally or for specific route groups:

```go
router.Use(middleware.DataLoaderMiddleware(userRepo))
```

### 5. Best Practices
1.  **Request Scoping**: Always initialize new loaders in a middleware for each request.
2.  **Concurrency**: DataLoader handles concurrency automatically; use it to resolve multiple types of entities in parallel.
3.  **Fallback to Preload**: For simple 1:1 or 1:N relations that are always needed, GORM's `.Preload()` is still acceptable and often more performant than a DataLoader for REST endpoints.
4.  **Error Handling**: DataLoader returns errors per-key, allowing partial success scenarios.
5.  **Batch Size**: Configure batch capacity based on your use case (default: 100).

### 6. Example Use Case
The project includes a **Post** model that demonstrates DataLoader usage:
- `GET /api/v1/posts` - Fetches all posts and batches author loading (prevents N+1)
- `GET /api/v1/posts?user_id=X` - Fetches posts by specific user with efficient author loading
- `POST /api/v1/posts` - Create a new post
- `GET /api/v1/posts/:id` - Get a single post with author
- `DELETE /api/v1/posts/:id` - Delete a post (owner only)




## Logging & Observability

This project uses structured logging and distributed tracing patterns for better observability.

### 1. Structured Logging with `slog`
Use the custom logger in `pkg/logger` for all application logs. Logs are output in JSON format.

```go
import "goapi/pkg/logger"

// Simple logging
logger.Info("Message", "key", value)
logger.Error("Failed operation", "error", err)

// Context-aware logging (includes RequestID)
logger.WithContext(ctx).Info("User action", "user_id", id)
```

### 2. Request Identification
The `RequestID` middleware generates or propagates a unique ID for every HTTP request.
- **Key**: `RequestID` (accessible via `c.GetString("RequestID")`)
- **Header**: `X-Request-ID`

### 3. Custom Recovery
The `CustomRecovery` middleware catches panics, logs the stack trace in a structured format, and returns a sanitized JSON error to the client including the `request_id`.

### 4. Health Checks
The `/health` endpoint performs real-time checks on:
- **Database**: Ping to ensure PostgreSQL is reachable.
- **Redis**: Ping to ensure Redis is reachable.

```json
{
  "status": "healthy",
  "components": {
    "db": "up",
    "redis": "up"
  }
}
```

## Important Notes

- **No tests currently exist** - create tests when adding new features
- Use `go mod tidy` after adding imports
- Database runs on port 5433 (not default 5432)
- Redis runs on port 6380 (not default 6379)
- JWT secret is hardcoded for dev - should come from env in production
- Use `binding` tags for request validation (e.g., `binding:"required,email"`)
