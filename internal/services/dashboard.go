package services

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/jd7008911/aogeri-api/internal/auth"
	"github.com/jd7008911/aogeri-api/internal/db"
	"github.com/jd7008911/aogeri-api/internal/models"
)

// DashboardService provides dashboard-related operations.
type DashboardService struct {
	queries *db.Queries
	auth    *auth.AuthService
}

func NewDashboardService(queries *db.Queries, a *auth.AuthService) *DashboardService {
	return &DashboardService{queries: queries, auth: a}
}

// AuthMiddleware proxies to auth service middleware so handlers can use it.
func (d *DashboardService) AuthMiddleware(next http.Handler) http.Handler {
	return d.auth.AuthMiddleware(next)
}

// GetStats returns a minimal DashboardStats for now.
func (d *DashboardService) GetStats(ctx context.Context) (models.DashboardStats, error) {
	return models.DashboardStats{}, nil
}

// GetUserOverview returns a minimal overview for now.
func (d *DashboardService) GetUserOverview(ctx context.Context, userID uuid.UUID) (any, error) {
	return map[string]any{"overview": "none"}, nil
}

// GetSecurityStatus returns a minimal security status.
func (d *DashboardService) GetSecurityStatus(ctx context.Context) (any, error) {
	return map[string]any{"status": "ok"}, nil
}
