package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/signalroot/signalroot/internal/config"
	"github.com/signalroot/signalroot/internal/incident"
)

type contextKey string

const (
	ContextUser  contextKey = "user"
	ContextOrgID contextKey = "org_id"
)

// Claims represents JWT claims.
type Claims struct {
	jwt.RegisteredClaims
	UserID uuid.UUID `json:"user_id"`
	OrgID  uuid.UUID `json:"org_id"`
	Role   string    `json:"role"`
	Email  string    `json:"email"`
}

// Service handles authentication and authorization.
type Service struct {
	cfg    *config.Config
	pool   *pgxpool.Pool
	logger *zap.Logger
}

func NewService(cfg *config.Config, pool *pgxpool.Pool, logger *zap.Logger) *Service {
	return &Service{cfg: cfg, pool: pool, logger: logger}
}

// GenerateJWT creates a signed JWT for a user.
func (s *Service) GenerateJWT(user *incident.User) (string, error) {
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   user.ID.String(),
		},
		UserID: user.ID,
		OrgID:  user.OrgID,
		Role:   user.Role,
		Email:  user.Email,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWTSecret))
}

// ValidateJWT parses and validates a JWT.
func (s *Service) ValidateJWT(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(s.cfg.JWTSecret), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}

// ValidateAPIKey validates a Bearer API key.
func (s *Service) ValidateAPIKey(ctx context.Context, key string) (*incident.User, error) {
	if !strings.HasPrefix(key, "sr_live_") && !strings.HasPrefix(key, "sr_test_") {
		return nil, fmt.Errorf("invalid key format")
	}

	hash := sha256.Sum256([]byte(key))
	keyHash := hex.EncodeToString(hash[:])

	var userID uuid.UUID
	var orgID uuid.UUID
	var revokedAt *time.Time
	var expiresAt *time.Time

	err := s.pool.QueryRow(ctx, `SELECT user_id, org_id, revoked_at, expires_at FROM api_keys WHERE key_hash = $1`, keyHash).Scan(&userID, &orgID, &revokedAt, &expiresAt)
	if err != nil {
		return nil, fmt.Errorf("key not found")
	}
	if revokedAt != nil {
		return nil, fmt.Errorf("key revoked")
	}
	if expiresAt != nil && expiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("key expired")
	}

	// Update last_used_at
	s.pool.Exec(ctx, `UPDATE api_keys SET last_used_at = NOW() WHERE key_hash = $1`, keyHash)

	var user incident.User
	err = s.pool.QueryRow(ctx, `SELECT id, org_id, email, name, role FROM users WHERE id = $1`, userID).Scan(&user.ID, &user.OrgID, &user.Email, &user.Name, &user.Role)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}
	return &user, nil
}

// Middleware creates an authentication middleware.
func (s *Service) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip health checks
		if r.URL.Path == "/healthz" || r.URL.Path == "/readyz" {
			next.ServeHTTP(w, r)
			return
		}

		// Try Bearer token
		auth := r.Header.Get("Authorization")
		if strings.HasPrefix(auth, "Bearer ") {
			token := strings.TrimPrefix(auth, "Bearer ")

			// API key auth
			if strings.HasPrefix(token, "sr_") {
				user, err := s.ValidateAPIKey(r.Context(), token)
				if err != nil {
					http.Error(w, `{"error":"invalid API key"}`, http.StatusUnauthorized)
					return
				}
				ctx := context.WithValue(r.Context(), ContextUser, user)
				ctx = context.WithValue(ctx, ContextOrgID, user.OrgID)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// JWT auth
			claims, err := s.ValidateJWT(token)
			if err != nil {
				http.Error(w, `{"error":"invalid token"}`, http.StatusUnauthorized)
				return
			}
			user := &incident.User{
				ID:    claims.UserID,
				OrgID: claims.OrgID,
				Email: claims.Email,
				Role:  claims.Role,
			}
			ctx := context.WithValue(r.Context(), ContextUser, user)
			ctx = context.WithValue(ctx, ContextOrgID, claims.OrgID)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
	})
}

// RequireRole creates a middleware that requires a minimum role.
func RequireRole(minRole string, next http.Handler) http.Handler {
	roleLevel := map[string]int{
		"viewer": 0, "member": 1, "admin": 2, "owner": 3,
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := UserFromContext(r.Context())
		if user == nil {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}
		if roleLevel[user.Role] < roleLevel[minRole] {
			http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// UserFromContext extracts the user from request context.
func UserFromContext(ctx context.Context) *incident.User {
	user, _ := ctx.Value(ContextUser).(*incident.User)
	return user
}

// OrgIDFromContext extracts the org ID from request context.
func OrgIDFromContext(ctx context.Context) uuid.UUID {
	orgID, _ := ctx.Value(ContextOrgID).(uuid.UUID)
	return orgID
}
