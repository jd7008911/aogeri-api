// internal/handlers/stake.go
package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jd7008911/aogeri-api/internal/auth"
	"github.com/jd7008911/aogeri-api/internal/db"
	"github.com/jd7008911/aogeri-api/internal/models"
	"github.com/jd7008911/aogeri-api/internal/services"
	"github.com/jd7008911/aogeri-api/pkg/web"
)

type StakeHandler struct {
	queries      *db.Queries
	stakeService *services.StakingService
	authService  *auth.AuthService
	validate     *validator.Validate
}

func NewStakeHandler(queries *db.Queries, stakeService *services.StakingService, authService *auth.AuthService) *StakeHandler {
	return &StakeHandler{
		queries:      queries,
		stakeService: stakeService,
		authService:  authService,
		validate:     validator.New(),
	}
}

func (h *StakeHandler) RegisterRoutes(r chi.Router) {
	r.Route("/stakes", func(r chi.Router) {
		r.Use(h.authService.AuthMiddleware)

		r.Get("/", h.GetUserStakes)
		r.Post("/", h.CreateStake)
		r.Get("/{id}", h.GetStake)
		r.Post("/{id}/unstake", h.Unstake)
		r.Post("/{id}/claim", h.ClaimRewards)
		r.Get("/stats", h.GetStakingStats)
	})
}

func (h *StakeHandler) CreateStake(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		web.Error(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var req models.StakeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		web.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		web.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	stake, err := h.stakeService.CreateStake(r.Context(), userID, req)
	if err != nil {
		web.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	web.Respond(w, http.StatusCreated, stake)
}

func (h *StakeHandler) GetUserStakes(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		web.Error(w, http.StatusUnauthorized, "User not authenticated")
		return
	}
	stakes, err := h.stakeService.GetUserStakes(r.Context(), userID)
	if err != nil {
		web.Error(w, http.StatusInternalServerError, "Failed to fetch stakes")
		return
	}
	web.Respond(w, http.StatusOK, stakes)
}

func (h *StakeHandler) GetStake(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		web.Error(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	stakeIDStr := chi.URLParam(r, "id")
	stakeID, err := uuid.Parse(stakeIDStr)
	if err != nil {
		web.Error(w, http.StatusBadRequest, "Invalid stake ID")
		return
	}

	stake, err := h.stakeService.GetStakeByID(r.Context(), stakeID)
	if err != nil {
		web.Error(w, http.StatusNotFound, "Stake not found")
		return
	}

	// Verify ownership
	if stake.UserID != userID {
		web.Error(w, http.StatusForbidden, "Access denied")
		return
	}

	web.Respond(w, http.StatusOK, stake)
}

func (h *StakeHandler) Unstake(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		web.Error(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	stakeIDStr := chi.URLParam(r, "id")
	stakeID, err := uuid.Parse(stakeIDStr)
	if err != nil {
		web.Error(w, http.StatusBadRequest, "Invalid stake ID")
		return
	}

	err = h.stakeService.Unstake(r.Context(), stakeID, userID)
	if err != nil {
		web.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	web.Respond(w, http.StatusOK, map[string]string{
		"message": "Stake unstaked successfully",
	})
}

// ClaimRewards placeholder
func (h *StakeHandler) ClaimRewards(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		web.Error(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	stakeIDStr := chi.URLParam(r, "id")
	stakeID, err := uuid.Parse(stakeIDStr)
	if err != nil {
		web.Error(w, http.StatusBadRequest, "Invalid stake ID")
		return
	}

	// Verify stake exists and ownership
	stake, err := h.stakeService.GetStakeByID(r.Context(), stakeID)
	if err != nil {
		web.Error(w, http.StatusNotFound, "Stake not found")
		return
	}
	if stake.UserID != userID {
		web.Error(w, http.StatusForbidden, "Access denied")
		return
	}

	// Calculate rewards
	rewardsStr, err := h.stakeService.CalculateRewards(r.Context(), stakeID)
	if err != nil {
		web.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	// If zero, nothing to claim
	if rewardsStr == "0" {
		web.Respond(w, http.StatusOK, map[string]any{"claimed": "0"})
		return
	}

	// Convert stakeID -> pgtype.UUID
	var idPg pgtype.UUID
	copy(idPg.Bytes[:], stakeID[:])
	idPg.Valid = true

	// Convert rewards string -> pgtype.Numeric
	var rn pgtype.Numeric
	if err := rn.Scan(rewardsStr); err != nil {
		web.Error(w, http.StatusInternalServerError, "invalid rewards value")
		return
	}

	// Update DB (increment rewards_claimed)
	if err := h.queries.UpdateStakeRewards(r.Context(), db.UpdateStakeRewardsParams{
		ID:             idPg,
		RewardsClaimed: rn,
	}); err != nil {
		web.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Return the claimed amount as number string
	if _, err := strconv.ParseFloat(rewardsStr, 64); err != nil {
		web.Respond(w, http.StatusOK, map[string]any{"claimed": rewardsStr})
		return
	}

	web.Respond(w, http.StatusOK, map[string]any{"claimed": rewardsStr})
}

// GetStakingStats placeholder
func (h *StakeHandler) GetStakingStats(w http.ResponseWriter, r *http.Request) {
	web.Respond(w, http.StatusOK, map[string]any{"total_staked": "0"})
}
