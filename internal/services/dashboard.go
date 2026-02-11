package services

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
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
	// Convert uuid.UUID -> pgtype.UUID
	var uid pgtype.UUID
	copy(uid.Bytes[:], userID[:])
	uid.Valid = true

	// Profile (optional)
	profile, err := d.queries.GetUserProfile(ctx, uid)
	var prof map[string]any
	if err == nil {
		prof = map[string]any{
			"username":   profile.Username.String,
			"full_name":  profile.FullName.String,
			"avatar_url": profile.AvatarUrl.String,
			"country":    profile.Country.String,
			"timezone":   profile.Timezone.String,
		}
	} else {
		prof = map[string]any{}
	}

	// Stakes
	stakes, err := d.queries.GetUserStakes(ctx, uid)
	if err != nil {
		return nil, err
	}

	totalStaked := 0.0
	totalRewards := 0.0
	recent := make([]map[string]any, 0, len(stakes))
	for i, s := range stakes {
		amt := 0.0
		if s.Amount.Valid {
			if fv, err := s.Amount.Float64Value(); err == nil {
				amt = fv.Float64
			}
		}
		rewards := 0.0
		if s.RewardsClaimed.Valid {
			if rv, err := s.RewardsClaimed.Float64Value(); err == nil {
				rewards = rv.Float64
			}
		}
		totalStaked += amt
		totalRewards += rewards

		// include up to 5 most recent
		if i < 5 {
			apy := 0.0
			if s.Apy.Valid {
				if fv, err := s.Apy.Float64Value(); err == nil {
					apy = fv.Float64
				}
			}
			recent = append(recent, map[string]any{
				"id":           s.ID,
				"token_symbol": s.Symbol,
				"amount":       fmt.Sprintf("%f", amt),
				"apy":          apy,
				"start_date":   s.StartDate.Time,
				"status":       s.Status.String,
			})
		}
	}

	overview := map[string]any{
		"profile":               prof,
		"active_stakes_count":   len(stakes),
		"total_staked":          strconv.FormatFloat(totalStaked, 'f', -1, 64),
		"total_rewards_claimed": strconv.FormatFloat(totalRewards, 'f', -1, 64),
		"recent_stakes":         recent,
	}

	return overview, nil
}

// GetSecurityStatus returns a minimal security status.
func (d *DashboardService) GetSecurityStatus(ctx context.Context) (any, error) {
	// If request is authenticated, include per-user security details
	if u, ok := auth.GetUserFromContext(ctx); ok {
		twoFA := false
		if u.TwoFactorEnabled.Valid {
			twoFA = u.TwoFactorEnabled.Bool
		}
		var failed int32
		if u.FailedLoginAttempts.Valid {
			failed = u.FailedLoginAttempts.Int32
		}
		locked := false
		lockedUntil := ""
		if u.LockedUntil.Valid {
			locked = true
			lockedUntil = u.LockedUntil.Time.String()
		}
		lastLogin := ""
		if u.LastLogin.Valid {
			lastLogin = u.LastLogin.Time.String()
		}

		// Simple scoring: start at 100, penalize for missing 2FA and failed attempts
		score := 100.0
		recommendations := []string{}
		if !twoFA {
			score -= 40.0
			recommendations = append(recommendations, "Enable two-factor authentication")
		}
		if failed > 0 {
			penalty := float64(failed) * 2.5
			if penalty > 30 {
				penalty = 30
			}
			score -= penalty
			recommendations = append(recommendations, "Review recent failed login attempts")
		}
		if locked {
			recommendations = append(recommendations, "Account locked — follow unlock procedures")
		}
		if score < 0 {
			score = 0
		}

		return map[string]any{
			"status":                "ok",
			"security_score":        score,
			"two_factor_enabled":    twoFA,
			"failed_login_attempts": failed,
			"locked":                locked,
			"locked_until":          lockedUntil,
			"last_login":            lastLogin,
			"recommendations":       recommendations,
		}, nil
	}

	// No user in context: return a generic system-level status
	return map[string]any{
		"status":         "ok",
		"security_score": 90.0,
		"notes":          "authenticated users receive per-account details",
	}, nil
}
