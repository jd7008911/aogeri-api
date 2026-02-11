// internal/services/governance.go
package services

import (
	"context"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jd7008911/aogeri-api/internal/db"
	"github.com/jd7008911/aogeri-api/internal/models"
)

type GovernanceService struct {
	queries *db.Queries
}

func NewGovernanceService(queries *db.Queries, _ any) *GovernanceService {
	return &GovernanceService{queries: queries}
}

func (g *GovernanceService) GetActiveProposals(ctx context.Context) ([]models.Proposal, error) {
	rows, err := g.queries.GetActiveProposals(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]models.Proposal, 0, len(rows))
	for _, r := range rows {
		var id uuid.UUID
		if r.ID.Valid {
			id, _ = uuid.FromBytes(r.ID.Bytes[:])
		}
		var proposer uuid.UUID
		if r.ProposerID.Valid {
			proposer, _ = uuid.FromBytes(r.ProposerID.Bytes[:])
		}

		quorum := 0.0
		if r.Quorum.Valid {
			if fv, err := r.Quorum.Float64Value(); err == nil {
				quorum = fv.Float64
			}
		}
		threshold := 0.0
		if r.Threshold.Valid {
			if fv, err := r.Threshold.Float64Value(); err == nil {
				threshold = fv.Float64
			}
		}

		forVotes := "0"
		if r.ForVotes.Valid {
			if fv, err := r.ForVotes.Float64Value(); err == nil {
				forVotes = strconv.FormatFloat(fv.Float64, 'f', -1, 64)
			}
		}
		againstVotes := "0"
		if r.AgainstVotes.Valid {
			if fv, err := r.AgainstVotes.Float64Value(); err == nil {
				againstVotes = strconv.FormatFloat(fv.Float64, 'f', -1, 64)
			}
		}
		abstainVotes := "0"
		if r.AbstainVotes.Valid {
			if fv, err := r.AbstainVotes.Float64Value(); err == nil {
				abstainVotes = strconv.FormatFloat(fv.Float64, 'f', -1, 64)
			}
		}

		var votingEnd time.Time
		if r.VotingEnd.Valid {
			votingEnd = r.VotingEnd.Time
		}

		out = append(out, models.Proposal{
			ID:           id,
			Title:        r.Title,
			Description:  r.Description,
			ProposerID:   proposer,
			Type:         r.ProposalType,
			Status:       r.Status.String,
			VotingEnd:    votingEnd,
			Quorum:       quorum,
			Threshold:    threshold,
			ForVotes:     forVotes,
			AgainstVotes: againstVotes,
			AbstainVotes: abstainVotes,
		})
	}
	return out, nil
}
