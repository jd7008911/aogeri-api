// internal/auth/middleware.go
package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jd7008911/aogeri-api/internal/db"
)

type contextKey string

const (
	UserIDKey contextKey = "user_id"
	UserKey   contextKey = "user"
)

func (s *AuthService) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}

		token := parts[1]
		claims, err := s.ValidateToken(token)
		if err != nil {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		// Get user from database
		// queries expect pgtype.UUID - convert
		var pgid pgtype.UUID
		copy(pgid.Bytes[:], claims.UserID[:])
		pgid.Valid = true
		user, err := s.queries.GetUserByID(r.Context(), pgid)
		if err != nil || !user.IsActive.Valid || !user.IsActive.Bool {
			http.Error(w, "User not found or inactive", http.StatusUnauthorized)
			return
		}

		// Add to context
		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, UserKey, &user)
		ctx = context.WithValue(ctx, middleware.RequestIDKey, r.Context().Value(middleware.RequestIDKey))

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *AuthService) OptionalAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
				token := parts[1]
				claims, err := s.ValidateToken(token)
				if err == nil {
					var pgid2 pgtype.UUID
					copy(pgid2.Bytes[:], claims.UserID[:])
					pgid2.Valid = true
					user, err := s.queries.GetUserByID(r.Context(), pgid2)
					if err == nil && user.IsActive.Valid && user.IsActive.Bool {
						ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
						ctx = context.WithValue(ctx, UserKey, &user)
						r = r.WithContext(ctx)
					}
				}
			}
		}
		next.ServeHTTP(w, r)
	})
}

func GetUserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value(UserIDKey).(uuid.UUID)
	return userID, ok
}

func GetUserFromContext(ctx context.Context) (*db.User, bool) {
	user, ok := ctx.Value(UserKey).(*db.User)
	return user, ok
}
