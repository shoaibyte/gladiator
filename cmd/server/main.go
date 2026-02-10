package main

import (
	"context"
	"embed"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
	"gladiator/internal/config"
	"gladiator/internal/database"
	"gladiator/internal/middleware"
)

//go:embed all:frontend_dist
var frontendEmbed embed.FS

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic("config: " + err.Error())
	}

	entClient, err := database.NewEntClient(cfg.DatabaseURL)
	if err != nil {
		panic("database: " + err.Error())
	}
	defer entClient.Close()

	redisClient, err := database.NewRedisClient(cfg.RedisURL)
	if err != nil {
		panic("redis: " + err.Error())
	}
	defer redisClient.Close()

	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	if cfg.Environment == "development" {
		logger = logger.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	e := echo.New()
	e.HideBanner = true
	e.Use(echomw.Recover())
	e.Use(echomw.RequestID())
	e.Use(middleware.Logger(logger))
	e.Use(echomw.CORSWithConfig(echomw.CORSConfig{
		AllowOrigins: []string{"http://localhost:5173", "http://127.0.0.1:5173"},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
	}))

	e.GET("/health", func(c echo.Context) error {
		ctx := c.Request().Context()
		if _, err := entClient.User.Query().Count(ctx); err != nil {
			return c.JSON(http.StatusServiceUnavailable, map[string]string{"status": "unhealthy", "error": "database"})
		}
		if redisClient.Ping(ctx).Err() != nil {
			return c.JSON(http.StatusServiceUnavailable, map[string]string{"status": "unhealthy", "error": "redis"})
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "healthy"})
	})

	// SPA: serve embedded frontend (skip for /api and /health)
	e.Use(echomw.StaticWithConfig(echomw.StaticConfig{
		Root:       "frontend_dist",
		Filesystem: http.FS(frontendEmbed),
		HTML5:      true,
		Skipper: func(c echo.Context) bool {
			path := c.Request().URL.Path
			return (len(path) >= 4 && path[:4] == "/api") || path == "/health"
		},
	}))

	portStr := fmt.Sprintf("%d", cfg.Port)
	addr := ":" + portStr

	go func() {
		if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal(err)
		}
	}()

	e.Logger.Printf("Server started on port %s", portStr)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}
