// internal/handlers/dashboard.go
package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jd7008911/aogeri-api/internal/auth"
	"github.com/jd7008911/aogeri-api/internal/services"
	"github.com/jd7008911/aogeri-api/pkg/web"
)

type DashboardHandler struct {
	dashboardService *services.DashboardService
}

func NewDashboardHandler(dashboardService *services.DashboardService) *DashboardHandler {
	return &DashboardHandler{
		dashboardService: dashboardService,
	}
}

func (h *DashboardHandler) RegisterRoutes(r chi.Router) {
	r.Route("/dashboard", func(r chi.Router) {
		r.Use(h.dashboardService.AuthMiddleware)
		r.Get("/stats", h.GetDashboardStats)
		r.Get("/overview", h.GetOverview)
		r.Get("/security", h.GetSecurityStatus)
	})
}

func (h *DashboardHandler) GetDashboardStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.dashboardService.GetStats(r.Context())
	if err != nil {
		web.Error(w, http.StatusInternalServerError, "Failed to fetch dashboard stats")
		return
	}

	web.Respond(w, http.StatusOK, stats)
}

func (h *DashboardHandler) GetOverview(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		web.Error(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	overview, err := h.dashboardService.GetUserOverview(r.Context(), userID)
	if err != nil {
		web.Error(w, http.StatusInternalServerError, "Failed to fetch overview")
		return
	}

	web.Respond(w, http.StatusOK, overview)
}

func (h *DashboardHandler) GetSecurityStatus(w http.ResponseWriter, r *http.Request) {
	status, err := h.dashboardService.GetSecurityStatus(r.Context())
	if err != nil {
		web.Error(w, http.StatusInternalServerError, "Failed to fetch security status")
		return
	}
	web.Respond(w, http.StatusOK, status)
}
