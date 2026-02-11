// internal/auth/auth.go
package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jd7008911/aogeri-api/internal/config"
	"github.com/jd7008911/aogeri-api/internal/db"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrAccountLocked      = errors.New("account is locked")
	ErrTokenExpired       = errors.New("token expired")
	ErrTokenInvalid       = errors.New("invalid token")
)

type AuthService struct {
	queries *db.Queries
	config  *config.Config
	store   Store
}

type Store interface {
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Delete(ctx context.Context, key string) error
}

type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"`
}

func NewAuthService(queries *db.Queries, config *config.Config, store Store) *AuthService {
	return &AuthService{
		queries: queries,
		config:  config,
		store:   store,
	}
}

func (s *AuthService) Register(ctx context.Context, params db.CreateUserParams) (*db.User, error) {
	// Check if user exists
	existing, err := s.queries.GetUserByEmail(ctx, params.Email)
	if err == nil {
		// existing.ID is pgtype.UUID - treat non-zero bytes as present
		var zero [16]byte
		if existing.ID.Bytes != zero {
			return nil, errors.New("user already exists")
		}
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(params.PasswordHash), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	params.PasswordHash = string(hashedPassword)

	// Create user
	user, err := s.queries.CreateUser(ctx, params)
	if err != nil {
		return nil, err
	}

	// Create user profile (use pgtype values)
	var profileUsername pgtype.Text
	profileUsername = pgtype.Text{String: user.Email, Valid: true}
	_, err = s.queries.CreateUserProfile(ctx, db.CreateUserProfileParams{
		UserID:   user.ID,
		Username: profileUsername,
		FullName: pgtype.Text{Valid: false},
		Country:  pgtype.Text{Valid: false},
	})
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*TokenPair, *db.User, error) {
	// Check if account is locked
	lockKey := "login_lock:" + email
	if locked, _ := s.store.Get(ctx, lockKey); locked != "" {
		return nil, nil, ErrAccountLocked
	}

	// Get user
	user, err := s.queries.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Simulate delay to prevent timing attacks
			bcrypt.CompareHashAndPassword([]byte("$2a$10$fakehash"), []byte(password))
			return nil, nil, ErrInvalidCredentials
		}
		return nil, nil, err
	}

	// Check password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		// Increment failed attempts (user.FailedLoginAttempts is pgtype.Int4)
		var currentAttempts int32
		if user.FailedLoginAttempts.Valid {
			currentAttempts = user.FailedLoginAttempts.Int32
		}
		attempts := currentAttempts + 1
		if int(attempts) >= s.config.Security.MaxLoginAttempts {
			lockUntil := time.Now().Add(s.config.Security.LockoutDuration)
			s.store.Set(ctx, lockKey, "locked", s.config.Security.LockoutDuration)
			s.queries.UpdateLoginAttempts(ctx, db.UpdateLoginAttemptsParams{
				ID:                  user.ID,
				FailedLoginAttempts: pgtype.Int4{Int32: attempts, Valid: true},
				LockedUntil:         pgtype.Timestamp{Time: lockUntil, Valid: true},
				LastLogin:           user.LastLogin,
			})
		} else {
			s.queries.UpdateLoginAttempts(ctx, db.UpdateLoginAttemptsParams{
				ID:                  user.ID,
				FailedLoginAttempts: pgtype.Int4{Int32: attempts, Valid: true},
				LockedUntil:         user.LockedUntil,
				LastLogin:           user.LastLogin,
			})
		}
		return nil, nil, ErrInvalidCredentials
	}

	// Reset failed attempts on successful login
	now := time.Now()
	s.queries.UpdateLoginAttempts(ctx, db.UpdateLoginAttemptsParams{
		ID:                  user.ID,
		FailedLoginAttempts: pgtype.Int4{Int32: 0, Valid: true},
		LockedUntil:         pgtype.Timestamp{Valid: false},
		LastLogin:           pgtype.Timestamp{Time: now, Valid: true},
	})

	// Generate tokens (convert pgtype.UUID to uuid.UUID)
	uid, err := pgUUIDToUUID(user.ID)
	if err != nil {
		return nil, nil, err
	}
	tokenPair, err := s.generateTokenPair(uid, user.Email)
	if err != nil {
		return nil, nil, err
	}

	// Store refresh token
	refreshTokenHash := hashToken(tokenPair.RefreshToken, s.config.JWT.Secret)
	err = s.store.Set(ctx, "refresh_token:"+refreshTokenHash, uid.String(),
		s.config.JWT.RefreshDuration)
	if err != nil {
		return nil, nil, err
	}

	return tokenPair, &user, nil
}

func (s *AuthService) generateTokenPair(userID uuid.UUID, email string) (*TokenPair, error) {
	// Access token
	accessExp := time.Now().Add(s.config.JWT.AccessDuration)
	accessClaims := &Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(accessExp),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "aogeri-api",
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(s.config.JWT.Secret))
	if err != nil {
		return nil, err
	}

	// Refresh token
	refreshToken := generateRandomToken()

	return &TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshToken,
		ExpiresAt:    accessExp.Unix(),
	}, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*TokenPair, error) {
	// Validate refresh token
	refreshTokenHash := hashToken(refreshToken, s.config.JWT.Secret)
	userIDStr, err := s.store.Get(ctx, "refresh_token:"+refreshTokenHash)
	if err != nil || userIDStr == "" {
		return nil, ErrTokenInvalid
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, ErrTokenInvalid
	}

	// Get user
	var pgid pgtype.UUID
	copy(pgid.Bytes[:], userID[:])
	pgid.Valid = true
	user, err := s.queries.GetUserByID(ctx, pgid)
	if err != nil {
		return nil, ErrTokenInvalid
	}

	// Delete old refresh token
	s.store.Delete(ctx, "refresh_token:"+refreshTokenHash)

	uid, err := pgUUIDToUUID(user.ID)
	if err != nil {
		return nil, ErrTokenInvalid
	}
	return s.generateTokenPair(uid, user.Email)
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	refreshTokenHash := hashToken(refreshToken, s.config.JWT.Secret)
	return s.store.Delete(ctx, "refresh_token:"+refreshTokenHash)
}

func (s *AuthService) ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.config.JWT.Secret), nil
	})

	if err != nil || !token.Valid {
		return nil, ErrTokenInvalid
	}

	return claims, nil
}

// pgUUIDToUUID converts pgtype.UUID to uuid.UUID
func pgUUIDToUUID(u pgtype.UUID) (uuid.UUID, error) {
	var zero [16]byte
	if u.Bytes == zero {
		return uuid.Nil, errors.New("invalid uuid")
	}
	b := u.Bytes
	return uuid.FromBytes(b[:])
}

// uuidToPgUUID converts uuid.UUID to pgtype.UUID
func uuidToPgUUID(id uuid.UUID) (pgtype.UUID, error) {
	var pg pgtype.UUID
	copy(pg.Bytes[:], id[:])
	pg.Valid = true
	return pg, nil
}

func generateRandomToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func hashToken(token, secret string) string {
	hash := sha256.Sum256([]byte(token + secret))
	return hex.EncodeToString(hash[:])
}
