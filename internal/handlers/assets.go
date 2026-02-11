// internal/handlers/assets.go
package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jd7008911/aogeri-api/pkg/web"
)

type AssetsHandler struct{}

func NewAssetsHandler() *AssetsHandler {
	return &AssetsHandler{}
}

func (h *AssetsHandler) RegisterRoutes(r chi.Router) {
	r.Get("/assets", h.GetAssets)
}

func (h *AssetsHandler) GetAssets(w http.ResponseWriter, r *http.Request) {
	// Return an empty list for now; real implementation lives in services.
	web.Respond(w, http.StatusOK, []any{})
}
