// internal/handlers/assets.go
package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jd7008911/aogeri-api/internal/services"
	"github.com/jd7008911/aogeri-api/pkg/web"
)

type AssetsHandler struct {
	assetsService *services.AssetsService
}

func NewAssetsHandler(s *services.AssetsService) *AssetsHandler {
	return &AssetsHandler{assetsService: s}
}

func (h *AssetsHandler) RegisterRoutes(r chi.Router) {
	r.Get("/assets", h.GetAssets)
}

func (h *AssetsHandler) GetAssets(w http.ResponseWriter, r *http.Request) {
	list, err := h.assetsService.GetAssets(r.Context())
	if err != nil {
		web.Error(w, http.StatusInternalServerError, "Failed to fetch assets")
		return
	}
	web.Respond(w, http.StatusOK, list)
}
