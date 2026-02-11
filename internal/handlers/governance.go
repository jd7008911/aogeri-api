// internal/handlers/governance.go
package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jd7008911/aogeri-api/internal/services"
	"github.com/jd7008911/aogeri-api/pkg/web"
)

type GovernanceHandler struct {
	svc *services.GovernanceService
}

func NewGovernanceHandler(_ any, svc *services.GovernanceService) *GovernanceHandler {
	return &GovernanceHandler{svc: svc}
}

func (h *GovernanceHandler) RegisterRoutes(r chi.Router) {
	r.Get("/proposals", h.ListProposals)
}

func (h *GovernanceHandler) ListProposals(w http.ResponseWriter, r *http.Request) {
	list, err := h.svc.GetActiveProposals(r.Context())
	if err != nil {
		web.Error(w, http.StatusInternalServerError, "Failed to fetch proposals")
		return
	}
	web.Respond(w, http.StatusOK, list)
}
