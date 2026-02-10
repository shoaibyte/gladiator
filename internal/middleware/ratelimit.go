package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"gladiator/internal/database"
)

const (
	RateLimitAuthPerIP        = 5
	RateLimitAuthWindow       = time.Minute
	RateLimitAPIPerUser       = 100
	RateLimitAPIWindow        = time.Minute
	RateLimitExecPerNotebook  = 10
	RateLimitExecWindow       = time.Minute
)

// RateLimitAuth returns middleware that limits requests per IP (e.g. 5/min for auth routes).
func RateLimitAuth(redis *database.RedisClient) echo.MiddlewareFunc {
	return rateLimit(redis, "auth:", func(c echo.Context) string {
		return c.RealIP()
	}, RateLimitAuthPerIP, RateLimitAuthWindow)
}

// RateLimitAPI returns middleware that limits requests per user (e.g. 100/min). Requires user_id in context (use after JWTAuth).
func RateLimitAPI(redis *database.RedisClient) echo.MiddlewareFunc {
	return rateLimit(redis, "api:", func(c echo.Context) string {
		if uid, ok := c.Get("user_id").(string); ok && uid != "" {
			return uid
		}
		return c.RealIP()
	}, RateLimitAPIPerUser, RateLimitAPIWindow)
}

// RateLimitExecution returns middleware that limits execute requests per notebook (10/min). Use on execute route; param must be "id".
func RateLimitExecution(redis *database.RedisClient) echo.MiddlewareFunc {
	return rateLimit(redis, "exec:", func(c echo.Context) string {
		return c.Param("id")
	}, RateLimitExecPerNotebook, RateLimitExecWindow)
}

func rateLimit(redis *database.RedisClient, prefix string, keyFn func(echo.Context) string, limit int, window time.Duration) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			key := prefix + keyFn(c)
			ctx := c.Request().Context()
			n, err := redis.Increment(ctx, key)
			if err != nil {
				return next(c)
			}
			if n == 1 {
				_ = redis.Expire(ctx, key, window)
			}
			remaining := limit - int(n)
			if remaining < 0 {
				remaining = 0
			}
			c.Response().Header().Set("X-RateLimit-Limit", strconv.Itoa(limit))
			c.Response().Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
			c.Response().Header().Set("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(window).Unix(), 10))
			if n > int64(limit) {
				return c.JSON(http.StatusTooManyRequests, map[string]string{"error": "rate limit exceeded"})
			}
			return next(c)
		}
	}
}
