// internal/handlers/auth.go
package handlers

import (
	"encoding/json"
	"net/http"

	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jd7008911/aogeri-api/internal/auth"
	"github.com/jd7008911/aogeri-api/internal/db"
	"github.com/jd7008911/aogeri-api/internal/models"
	"github.com/jd7008911/aogeri-api/pkg/web"
)

type AuthHandler struct {
	authService *auth.AuthService
	queries     *db.Queries
	validate    *validator.Validate
}

func NewAuthHandler(authService *auth.AuthService, queries *db.Queries) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		queries:     queries,
		validate:    validator.New(),
	}
}

func (h *AuthHandler) RegisterRoutes(r chi.Router) {
	r.Post("/register", h.Register)
	r.Post("/login", h.Login)
	r.Post("/refresh", h.RefreshToken)
	r.Post("/logout", h.Logout)

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(h.authService.AuthMiddleware)
		r.Get("/profile", h.GetProfile)
		r.Put("/profile", h.UpdateProfile)
		r.Post("/change-password", h.ChangePassword)
		r.Post("/enable-2fa", h.Enable2FA)
	})
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		web.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		web.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	// Validate password strength
	if err := auth.ValidatePasswordStrength(req.Password); err != nil {
		web.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	// Build CreateUserParams with pgtype.Text for wallet address
	wa := pgtype.Text{}
	if req.WalletAddress != "" {
		wa = pgtype.Text{String: req.WalletAddress, Valid: true}
	} else {
		wa = pgtype.Text{Valid: false}
	}

	userParams := db.CreateUserParams{
		Email:         req.Email,
		PasswordHash:  req.Password,
		WalletAddress: wa,
	}

	user, err := h.authService.Register(r.Context(), userParams)
	if err != nil {
		web.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Convert user.ID (pgtype.UUID) to uuid.UUID for response
	var uid uuid.UUID
	if user.ID.Valid {
		u, err := uuid.FromBytes(user.ID.Bytes[:])
		if err == nil {
			uid = u
		}
	}

	web.Respond(w, http.StatusCreated, map[string]interface{}{
		"message": "User registered successfully",
		"user_id": uid,
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		web.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		web.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	tokenPair, user, err := h.authService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		switch err {
		case auth.ErrInvalidCredentials:
			web.Error(w, http.StatusUnauthorized, "Invalid credentials")
		case auth.ErrAccountLocked:
			web.Error(w, http.StatusLocked, "Account is locked")
		default:
			web.Error(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	// Convert db.User (pgtype fields) to models.User
	var uid uuid.UUID
	if user.ID.Valid {
		if u, err := uuid.FromBytes(user.ID.Bytes[:]); err == nil {
			uid = u
		}
	}

	var lastLogin *time.Time
	if user.LastLogin.Valid {
		t := user.LastLogin.Time
		lastLogin = &t
	}

	createdAt := time.Time{}
	if user.CreatedAt.Valid {
		createdAt = user.CreatedAt.Time
	}
	updatedAt := time.Time{}
	if user.UpdatedAt.Valid {
		updatedAt = user.UpdatedAt.Time
	}

	response := models.LoginResponse{
		User: &models.User{
			ID:        uid,
			Email:     user.Email,
			IsActive:  user.IsActive.Valid && user.IsActive.Bool,
			LastLogin: lastLogin,
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
		},
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresAt:    tokenPair.ExpiresAt,
	}

	web.Respond(w, http.StatusOK, response)
}

func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		web.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	tokenPair, err := h.authService.RefreshToken(r.Context(), req.RefreshToken)
	if err != nil {
		web.Error(w, http.StatusUnauthorized, "Invalid refresh token")
		return
	}

	web.Respond(w, http.StatusOK, tokenPair)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		web.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.authService.Logout(r.Context(), req.RefreshToken); err != nil {
		web.Error(w, http.StatusInternalServerError, "Failed to logout")
		return
	}

	web.Respond(w, http.StatusOK, map[string]string{
		"message": "Logged out successfully",
	})
}

func (h *AuthHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		web.Error(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var pgid pgtype.UUID
	copy(pgid.Bytes[:], userID[:])
	pgid.Valid = true
	profile, err := h.queries.GetUserProfile(r.Context(), pgid)
	if err != nil {
		web.Error(w, http.StatusNotFound, "Profile not found")
		return
	}

	web.Respond(w, http.StatusOK, profile)
}

// UpdateProfile is a placeholder until full profile update is implemented.
func (h *AuthHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	web.Error(w, http.StatusNotImplemented, "not implemented")
}

// ChangePassword is a placeholder until implemented.
func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	web.Error(w, http.StatusNotImplemented, "not implemented")
}

// Enable2FA is a placeholder until implemented.
func (h *AuthHandler) Enable2FA(w http.ResponseWriter, r *http.Request) {
	web.Error(w, http.StatusNotImplemented, "not implemented")
}
