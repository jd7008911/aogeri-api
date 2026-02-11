package services

import (
	"context"
	"strconv"

	"github.com/google/uuid"
	"github.com/jd7008911/aogeri-api/internal/db"
	"github.com/jd7008911/aogeri-api/internal/models"
)

// AssetsService provides asset listing and metrics.
type AssetsService struct {
	queries *db.Queries
}

func NewAssetsService(q *db.Queries) *AssetsService {
	return &AssetsService{queries: q}
}

func (s *AssetsService) GetAssets(ctx context.Context) ([]models.Asset, error) {
	tokens, err := s.queries.GetTokenList(ctx)
	if err != nil {
		return nil, err
	}

	out := make([]models.Asset, 0, len(tokens))
	for _, t := range tokens {
		curr := "0"
		if t.MarketPrice.Valid {
			if fv, err := t.MarketPrice.Float64Value(); err == nil {
				curr = strconv.FormatFloat(fv.Float64, 'f', -1, 64)
			}
		}
		priceChange := 0.0
		if t.PriceChange24h.Valid {
			if fv, err := t.PriceChange24h.Float64Value(); err == nil {
				priceChange = fv.Float64
			}
		}

		var id uuid.UUID
		if t.ID.Valid {
			id, _ = uuid.FromBytes(t.ID.Bytes[:])
		}

		out = append(out, models.Asset{
			ID:               id,
			Symbol:           t.Symbol,
			Name:             t.Name,
			CurrentValue:     curr,
			PriceChange24H:   priceChange,
			Volume24H:        "",
			TotalValueLocked: "",
		})
	}

	return out, nil
}
