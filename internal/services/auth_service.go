package services

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gladiator/ent"
	"gladiator/ent/user"
	"gladiator/internal/database"
)

var (
	ErrEmailExists        = errors.New("email already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type AuthService struct {
	ent   *ent.Client
	redis *database.RedisClient
	cfg   AuthConfig
}

type AuthConfig interface {
	JWTSecret() string
	JWTAccessTTL() int
	JWTRefreshTTL() int
}

type authConfig struct {
	secret      string
	accessTTL   int
	refreshTTL  int
}

func (c *authConfig) JWTSecret() string     { return c.secret }
func (c *authConfig) JWTAccessTTL() int     { return c.accessTTL }
func (c *authConfig) JWTRefreshTTL() int    { return c.refreshTTL }

func NewAuthConfig(secret string, accessTTL, refreshTTL int) AuthConfig {
	return &authConfig{secret: secret, accessTTL: accessTTL, refreshTTL: refreshTTL}
}

func NewAuthService(entClient *ent.Client, redis *database.RedisClient, cfg AuthConfig) *AuthService {
	return &AuthService{ent: entClient, redis: redis, cfg: cfg}
}

// Secret returns the JWT secret for token validation (e.g. in middleware).
func (s *AuthService) Secret() string { return s.cfg.JWTSecret() }

// UserFromToken parses the JWT, validates session, and returns userID and userName (for WebSocket).
func (s *AuthService) UserFromToken(ctx context.Context, tokenStr string) (userID, userName string, err error) {
	var claims accessClaims
	tok, err := jwt.ParseWithClaims(tokenStr, &claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("bad method")
		}
		return []byte(s.cfg.JWTSecret()), nil
	})
	if err != nil || !tok.Valid {
		return "", "", ErrInvalidCredentials
	}
	if !s.ValidateSession(ctx, claims.UserID, claims.TokenID) {
		return "", "", ErrInvalidCredentials
	}
	u, err := s.ent.User.Get(ctx, uuid.MustParse(claims.UserID))
	if err != nil {
		return "", "", err
	}
	return claims.UserID, u.Name, nil
}

type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=8,max=72"`
	Name     string `json:"name" validate:"required,min=2,max=255"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	User         UserResponse
}

type UserResponse struct {
	ID        string  `json:"id"`
	Email     string  `json:"email"`
	Name      string  `json:"name"`
	AvatarURL *string `json:"avatar_url,omitempty"`
	CreatedAt string  `json:"created_at"`
}

const sessionTTL = 7 * 24 * time.Hour
const sessionKeyPrefix = "session:"

func userToResponse(u *ent.User) UserResponse {
	return UserResponse{
		ID:        u.ID.String(),
		Email:     u.Email,
		Name:      u.Name,
		AvatarURL: u.AvatarURL,
		CreatedAt: u.CreatedAt.Format(time.RFC3339),
	}
}

func (s *AuthService) Register(ctx context.Context, req RegisterRequest) (*ent.User, error) {
	exists, _ := s.ent.User.Query().Where(user.EmailEQ(req.Email)).Exist(ctx)
	if exists {
		return nil, ErrEmailExists
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		return nil, err
	}
	u, err := s.ent.User.Create().
		SetEmail(req.Email).
		SetPasswordHash(string(hash)).
		SetName(req.Name).
		Save(ctx)
	return u, err
}

type accessClaims struct {
	UserID  string `json:"user_id"`
	Email   string `json:"email"`
	TokenID string `json:"token_id"`
	jwt.RegisteredClaims
}

type refreshClaims struct {
	UserID  string `json:"user_id"`
	TokenID string `json:"token_id"`
	Type    string `json:"type"`
	jwt.RegisteredClaims
}

func (s *AuthService) Login(ctx context.Context, req LoginRequest) (*TokenResponse, error) {
	u, err := s.ent.User.Query().
		Where(user.EmailEQ(req.Email), user.IsActiveEQ(true)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}
	tokenID := uuid.New().String()
	accessTTL := time.Duration(s.cfg.JWTAccessTTL()) * time.Minute
	refreshTTL := time.Duration(s.cfg.JWTRefreshTTL()) * time.Minute
	now := time.Now()
	accessClaims := accessClaims{
		UserID:  u.ID.String(),
		Email:   u.Email,
		TokenID: tokenID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(accessTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	refreshClaims := refreshClaims{
		UserID:  u.ID.String(),
		TokenID: tokenID,
		Type:    "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(refreshTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessStr, err := accessToken.SignedString([]byte(s.cfg.JWTSecret()))
	if err != nil {
		return nil, err
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshStr, err := refreshToken.SignedString([]byte(s.cfg.JWTSecret()))
	if err != nil {
		return nil, err
	}
	key := sessionKeyPrefix + u.ID.String() + ":" + tokenID
	if err := s.redis.SetWithExpiry(ctx, key, "1", sessionTTL); err != nil {
		// non-fatal: log and continue
	}
	_, _ = s.ent.User.UpdateOneID(u.ID).SetLastLoginAt(now).Save(ctx)
	return &TokenResponse{
		AccessToken:  accessStr,
		RefreshToken:  refreshStr,
		ExpiresIn:     int(accessTTL.Seconds()),
		User:          userToResponse(u),
	}, nil
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*TokenResponse, error) {
	var claims refreshClaims
	tok, err := jwt.ParseWithClaims(refreshToken, &claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.cfg.JWTSecret()), nil
	})
	if err != nil || !tok.Valid || claims.Type != "refresh" {
		return nil, ErrInvalidCredentials
	}
	key := sessionKeyPrefix + claims.UserID + ":" + claims.TokenID
	ok, _ := s.redis.Exists(ctx, key)
	if !ok {
		return nil, ErrInvalidCredentials
	}
	u, err := s.ent.User.Get(ctx, uuid.MustParse(claims.UserID))
	if err != nil {
		return nil, err
	}
	now := time.Now()
	accessTTL := time.Duration(s.cfg.JWTAccessTTL()) * time.Minute
	accessClaims := accessClaims{
		UserID:  claims.UserID,
		Email:   u.Email,
		TokenID: claims.TokenID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(accessTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessStr, err := accessToken.SignedString([]byte(s.cfg.JWTSecret()))
	if err != nil {
		return nil, err
	}
	refreshTTL := time.Duration(s.cfg.JWTRefreshTTL()) * time.Minute
	refreshClaimsNew := refreshClaims{
		UserID:  claims.UserID,
		TokenID: claims.TokenID,
		Type:    "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(refreshTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	refreshTok := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaimsNew)
	refreshStr, err := refreshTok.SignedString([]byte(s.cfg.JWTSecret()))
	if err != nil {
		return nil, err
	}
	_ = s.redis.SetWithExpiry(ctx, key, "1", sessionTTL)
	return &TokenResponse{
		AccessToken:  accessStr,
		RefreshToken:  refreshStr,
		ExpiresIn:     int(accessTTL.Seconds()),
		User:          userToResponse(u),
	}, nil
}

func (s *AuthService) ValidateSession(ctx context.Context, userID, tokenID string) bool {
	key := sessionKeyPrefix + userID + ":" + tokenID
	ok, _ := s.redis.Exists(ctx, key)
	return ok
}

func (s *AuthService) Logout(ctx context.Context, userID, tokenID string) error {
	key := sessionKeyPrefix + userID + ":" + tokenID
	return s.redis.Delete(ctx, key)
}

func (s *AuthService) LogoutAll(ctx context.Context, userID string) error {
	pattern := sessionKeyPrefix + userID + ":*"
	return s.redis.DeleteByPattern(ctx, pattern)
}
