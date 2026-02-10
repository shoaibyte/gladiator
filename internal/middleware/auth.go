package middleware

import (
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"gladiator/internal/services"
)

type accessClaims struct {
	UserID  string `json:"user_id"`
	Email   string `json:"email"`
	TokenID string `json:"token_id"`
	jwt.RegisteredClaims
}

// JWTAuth returns Echo middleware that validates JWT and Redis session, sets user_id and token_id in context.
func JWTAuth(auth *services.AuthService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" || len(authHeader) < 8 || authHeader[:7] != "Bearer " {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "missing or invalid authorization"})
			}
			tokenStr := authHeader[7:]
			var claims accessClaims
			tok, err := jwt.ParseWithClaims(tokenStr, &claims, func(t *jwt.Token) (interface{}, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(auth.Secret()), nil
			})
			if err != nil || !tok.Valid {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid token"})
			}
			if !auth.ValidateSession(c.Request().Context(), claims.UserID, claims.TokenID) {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "session expired"})
			}
			c.Set("user_id", claims.UserID)
			c.Set("token_id", claims.TokenID)
			return next(c)
		}
	}
}
