// internal/services/staking.go
package services

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jd7008911/aogeri-api/internal/auth"
	"github.com/jd7008911/aogeri-api/internal/db"
	"github.com/jd7008911/aogeri-api/internal/models"
)

type stakingQuerier interface {
	GetTokenList(ctx context.Context) ([]db.GetTokenListRow, error)
	CreateStake(ctx context.Context, arg db.CreateStakeParams) (db.Stake, error)
	GetStakeByID(ctx context.Context, id pgtype.UUID) (db.GetStakeByIDRow, error)
	GetUserStakes(ctx context.Context, userID pgtype.UUID) ([]db.GetUserStakesRow, error)
	Unstake(ctx context.Context, arg db.UnstakeParams) error
}

type StakingService struct {
	queries stakingQuerier
	auth    *auth.AuthService
}

func NewStakingService(queries stakingQuerier, auth *auth.AuthService) *StakingService {
	return &StakingService{
		queries: queries,
		auth:    auth,
	}
}

func (s *StakingService) CreateStake(ctx context.Context, userID uuid.UUID, req models.StakeRequest) (*models.Stake, error) {
	// Get token ID from symbol
	tokens, err := s.queries.GetTokenList(ctx)
	if err != nil {
		return nil, err
	}

	var tokenID pgtype.UUID
	found := false
	for _, t := range tokens {
		if t.Symbol == req.TokenSymbol {
			tokenID = t.ID
			found = true
			break
		}
	}

	if !found {
		return nil, errors.New("token not found")
	}

	// Calculate end date
	endDate := time.Now().Add(time.Duration(req.DurationDays) * 24 * time.Hour)

	// Get current APY from external service or database
	apy := s.calculateAPY(req.TokenSymbol)

	// Convert uuid.UUID -> pgtype.UUID
	var uid pgtype.UUID
	copy(uid.Bytes[:], userID[:])
	uid.Valid = true

	// Convert amount string -> pgtype.Numeric
	var amt pgtype.Numeric
	if err := amt.Scan(req.Amount); err != nil {
		return nil, err
	}

	// Convert apy float64 -> pgtype.Numeric (use string scan)
	var apyn pgtype.Numeric
	if err := apyn.Scan(strconv.FormatFloat(apy, 'f', -1, 64)); err != nil {
		return nil, err
	}

	// Convert end date -> pgtype.Timestamp
	endPg := pgtype.Timestamp{Time: endDate, Valid: true}

	stake, err := s.queries.CreateStake(ctx, db.CreateStakeParams{
		UserID:       uid,
		TokenID:      tokenID,
		Amount:       amt,
		Apy:          apyn,
		EndDate:      endPg,
		AutoCompound: pgtype.Bool{Bool: req.AutoCompound, Valid: true},
	})
	if err != nil {
		return nil, err
	}

	// Convert returned pgtype fields to models.Stake
	// Convert returned pgtype.UUID -> uuid.UUID
	var id2 uuid.UUID
	if stake.ID.Valid {
		id2, _ = uuid.FromBytes(stake.ID.Bytes[:])
	}
	var uid2 uuid.UUID
	if stake.UserID.Valid {
		uid2, _ = uuid.FromBytes(stake.UserID.Bytes[:])
	}

	// Amount
	amountStr := "0"
	if stake.Amount.Valid {
		if fv, err := stake.Amount.Float64Value(); err == nil {
			amountStr = strconv.FormatFloat(fv.Float64, 'f', -1, 64)
		}
	}

	// APY
	apyFloat := 0.0
	if stake.Apy.Valid {
		if fv, err := stake.Apy.Float64Value(); err == nil {
			apyFloat = fv.Float64
		}
	}

	// Start/End dates
	var startDate time.Time
	if stake.StartDate.Valid {
		startDate = stake.StartDate.Time
	}
	var endTime *time.Time
	if stake.EndDate.Valid {
		t := stake.EndDate.Time
		endTime = &t
	}

	// Status
	status := ""
	if stake.Status.Valid {
		status = stake.Status.String
	}

	return &models.Stake{
		ID:           id2,
		UserID:       uid2,
		TokenSymbol:  req.TokenSymbol,
		Amount:       amountStr,
		APY:          apyFloat,
		StartDate:    startDate,
		EndDate:      endTime,
		Status:       status,
		AutoCompound: stake.AutoCompound.Bool,
		RewardsClaimed: func() string {
			if stake.RewardsClaimed.Valid {
				if rv, err := stake.RewardsClaimed.Float64Value(); err == nil {
					return strconv.FormatFloat(rv.Float64, 'f', -1, 64)
				}
			}
			return "0"
		}(),
	}, nil
}

func (s *StakingService) Unstake(ctx context.Context, stakeID, userID uuid.UUID) error {
	var id pgtype.UUID
	copy(id.Bytes[:], stakeID[:])
	id.Valid = true
	var uid pgtype.UUID
	copy(uid.Bytes[:], userID[:])
	uid.Valid = true
	return s.queries.Unstake(ctx, db.UnstakeParams{
		ID:     id,
		UserID: uid,
	})
}

func (s *StakingService) CalculateRewards(ctx context.Context, stakeID uuid.UUID) (string, error) {
	var id pgtype.UUID
	copy(id.Bytes[:], stakeID[:])
	id.Valid = true
	stake, err := s.queries.GetStakeByID(ctx, id)
	if err != nil {
		return "0", err
	}
	if !stake.Status.Valid || stake.Status.String != "active" {
		return "0", errors.New("stake is not active")
	}

	// Get start date
	if !stake.StartDate.Valid {
		return "0", errors.New("stake has no start date")
	}
	start := stake.StartDate.Time
	durationDays := time.Since(start).Hours() / 24

	// Convert amount and apy to float64
	amountF := 0.0
	if stake.Amount.Valid {
		if fv, err := stake.Amount.Float64Value(); err == nil {
			amountF = fv.Float64
		}
	}
	apyF := 0.0
	if stake.Apy.Valid {
		if fv, err := stake.Apy.Float64Value(); err == nil {
			apyF = fv.Float64
		}
	}

	dailyRate := apyF / 365.0 / 100.0
	rewards := amountF * dailyRate * durationDays

	return strconv.FormatFloat(rewards, 'f', -1, 64), nil
}

func (s *StakingService) calculateAPY(tokenSymbol string) float64 {
	// In production, fetch from external API or contract
	switch tokenSymbol {
	case "AOG":
		return 33.29
	case "BNB":
		return 12.5
	default:
		return 8.0
	}
}

// helper: convert pgtype.UUID to uuid.UUID
func pgToUUID(u pgtype.UUID) (uuid.UUID, error) {
	if !u.Valid {
		return uuid.Nil, errors.New("invalid uuid")
	}
	return uuid.FromBytes(u.Bytes[:])
}

func (s *StakingService) GetUserStakes(ctx context.Context, userID uuid.UUID) ([]models.Stake, error) {
	var uid pgtype.UUID
	copy(uid.Bytes[:], userID[:])
	uid.Valid = true

	rows, err := s.queries.GetUserStakes(ctx, uid)
	if err != nil {
		return nil, err
	}

	var out []models.Stake
	for _, r := range rows {
		id, _ := pgToUUID(r.ID)
		uid2, _ := pgToUUID(r.UserID)
		amt := "0"
		if r.Amount.Valid {
			if fv, err := r.Amount.Float64Value(); err == nil {
				amt = strconv.FormatFloat(fv.Float64, 'f', -1, 64)
			}
		}
		apy := 0.0
		if r.Apy.Valid {
			if fv, err := r.Apy.Float64Value(); err == nil {
				apy = fv.Float64
			}
		}
		var start time.Time
		if r.StartDate.Valid {
			start = r.StartDate.Time
		}
		var end *time.Time
		if r.EndDate.Valid {
			t := r.EndDate.Time
			end = &t
		}

		status := ""
		if r.Status.Valid {
			status = r.Status.String
		}

		out = append(out, models.Stake{
			ID:           id,
			UserID:       uid2,
			TokenSymbol:  r.Symbol,
			Amount:       amt,
			APY:          apy,
			StartDate:    start,
			EndDate:      end,
			Status:       status,
			AutoCompound: r.AutoCompound.Bool,
			RewardsClaimed: func() string {
				if r.RewardsClaimed.Valid {
					if rv, err := r.RewardsClaimed.Float64Value(); err == nil {
						return strconv.FormatFloat(rv.Float64, 'f', -1, 64)
					}
				}
				return "0"
			}(),
		})
	}

	return out, nil
}

func (s *StakingService) GetStakeByID(ctx context.Context, stakeID uuid.UUID) (models.Stake, error) {
	var id pgtype.UUID
	copy(id.Bytes[:], stakeID[:])
	id.Valid = true

	r, err := s.queries.GetStakeByID(ctx, id)
	if err != nil {
		return models.Stake{}, err
	}

	pid, _ := pgToUUID(r.ID)
	uid2, _ := pgToUUID(r.UserID)
	amt := "0"
	if r.Amount.Valid {
		if fv, err := r.Amount.Float64Value(); err == nil {
			amt = strconv.FormatFloat(fv.Float64, 'f', -1, 64)
		}
	}
	apy := 0.0
	if r.Apy.Valid {
		if fv, err := r.Apy.Float64Value(); err == nil {
			apy = fv.Float64
		}
	}
	var start time.Time
	if r.StartDate.Valid {
		start = r.StartDate.Time
	}
	var end *time.Time
	if r.EndDate.Valid {
		t := r.EndDate.Time
		end = &t
	}
	status := ""
	if r.Status.Valid {
		status = r.Status.String
	}

	return models.Stake{
		ID:           pid,
		UserID:       uid2,
		TokenSymbol:  r.Symbol,
		Amount:       amt,
		APY:          apy,
		StartDate:    start,
		EndDate:      end,
		Status:       status,
		AutoCompound: r.AutoCompound.Bool,
		RewardsClaimed: func() string {
			if r.RewardsClaimed.Valid {
				if rv, err := r.RewardsClaimed.Float64Value(); err == nil {
					return strconv.FormatFloat(rv.Float64, 'f', -1, 64)
				}
			}
			return "0"
		}(),
	}, nil
}
