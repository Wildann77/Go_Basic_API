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
pkg/utils/        # Utility functions
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
- Return errors, don't log in business logic
- Use `errors.New()` for simple errors
- Check `gorm.ErrRecordNotFound` for database operations
- Handlers use `utils.ErrorResponse()` for consistent error format

```go
func (r *userRepository) GetByID(id uint) (*models.User, error) {
    var user models.User
    if err := r.db.First(&user, id).Error; err != nil {
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

## Important Notes

- **No tests currently exist** - create tests when adding new features
- Use `go mod tidy` after adding imports
- Database runs on port 5433 (not default 5432)
- Redis runs on port 6380 (not default 6379)
- JWT secret is hardcoded for dev - should come from env in production
- Use `binding` tags for request validation (e.g., `binding:"required,email"`)
