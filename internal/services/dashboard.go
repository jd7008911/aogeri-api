package services

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

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
	var out models.DashboardStats

	// Asset-level metrics (TVL, active tokens, total assets)
	am, err := d.queries.GetAssetMetrics(ctx)
	if err != nil {
		return out, err
	}

	// Convert TotalTvl (driver returns various types) to string
	totalTvlStr := "0"
	switch v := am.TotalTvl.(type) {
	case nil:
		totalTvlStr = "0"
	case float64:
		totalTvlStr = strconv.FormatFloat(v, 'f', -1, 64)
	case string:
		totalTvlStr = v
	case []byte:
		totalTvlStr = string(v)
	default:
		totalTvlStr = fmt.Sprintf("%v", v)
	}

	// Total staked value (explicit query)
	totalStaked, err := d.queries.GetTotalStakedValue(ctx)
	if err != nil {
		return out, err
	}
	totalStakedStr := "0"
	if totalStaked.Valid {
		if fv, err := totalStaked.Float64Value(); err == nil {
			totalStakedStr = strconv.FormatFloat(fv.Float64, 'f', -1, 64)
		}
	}

	// Active governance proposals count
	props, err := d.queries.GetActiveProposals(ctx)
	if err != nil {
		return out, err
	}

	out = models.DashboardStats{
		TotalValueLocked:    totalTvlStr,
		ActiveMonitors:      int32(am.ActiveTokens),
		RemainingTime:       "N/A",
		ActiveStakes:        int32(am.TotalAssets),
		TotalRewards:        totalStakedStr,
		SecurityScore:       100.0,
		GovernanceProposals: int32(len(props)),
	}

	return out, nil
}

// GetUserOverview returns a minimal overview for now.
func (d *DashboardService) GetUserOverview(ctx context.Context, userID uuid.UUID) (any, error) {
	return map[string]any{"overview": "none"}, nil
}

// GetSecurityStatus returns a minimal security status.
func (d *DashboardService) GetSecurityStatus(ctx context.Context) (any, error) {
	return map[string]any{"status": "ok"}, nil
}
