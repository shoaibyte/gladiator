package middleware

import (
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

// Logger returns an Echo middleware that logs requests with zerolog.
func Logger(log zerolog.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			req := c.Request()
			res := c.Response()
			rid := res.Header().Get(echo.HeaderXRequestID)
			if rid == "" {
				rid = req.Header.Get(echo.HeaderXRequestID)
			}
			err := next(c)
			ev := log.Info()
			if err != nil {
				ev = log.Error().Err(err)
			}
			ev.Str("method", req.Method).
				Str("path", req.URL.Path).
				Int("status", res.Status).
				Dur("latency_ms", time.Since(start)).
				Str("ip", c.RealIP()).
				Str("user_agent", req.UserAgent()).
				Str("request_id", rid).
				Msg("request")
			return err
		}
	}
}
