// internal/handlers/liquidity.go
package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jd7008911/aogeri-api/pkg/web"
)

type LiquidityHandler struct{}

func NewLiquidityHandler() *LiquidityHandler {
	return &LiquidityHandler{}
}

func (h *LiquidityHandler) RegisterRoutes(r chi.Router) {
	r.Get("/liquidity", h.GetLiquidity)
}

func (h *LiquidityHandler) GetLiquidity(w http.ResponseWriter, r *http.Request) {
	web.Respond(w, http.StatusOK, map[string]any{"liquidity": 0})
}
