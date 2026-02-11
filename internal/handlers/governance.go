// internal/handlers/governance.go
package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jd7008911/aogeri-api/pkg/web"
)

type GovernanceHandler struct{}

func NewGovernanceHandler(queries any, svc any) *GovernanceHandler { // keep signature flexible
	return &GovernanceHandler{}
}

func (h *GovernanceHandler) RegisterRoutes(r chi.Router) {
	r.Get("/proposals", h.ListProposals)
}

func (h *GovernanceHandler) ListProposals(w http.ResponseWriter, r *http.Request) {
	web.Respond(w, http.StatusOK, []any{})
}
