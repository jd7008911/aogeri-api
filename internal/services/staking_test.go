package services

import (
	"context"
	"math"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jd7008911/aogeri-api/internal/db"
)

type fakeQueries struct {
	getStakeRow db.GetStakeByIDRow
}

func (f *fakeQueries) GetTokenList(ctx context.Context) ([]db.GetTokenListRow, error) {
	return []db.GetTokenListRow{}, nil
}
func (f *fakeQueries) CreateStake(ctx context.Context, arg db.CreateStakeParams) (db.Stake, error) {
	return db.Stake{}, nil
}
func (f *fakeQueries) GetStakeByID(ctx context.Context, id pgtype.UUID) (db.GetStakeByIDRow, error) {
	return f.getStakeRow, nil
}
func (f *fakeQueries) GetUserStakes(ctx context.Context, userID pgtype.UUID) ([]db.GetUserStakesRow, error) {
	return []db.GetUserStakesRow{}, nil
}
func (f *fakeQueries) Unstake(ctx context.Context, arg db.UnstakeParams) error {
	return nil
}

func TestCalculateAPY(t *testing.T) {
	s := NewStakingService(&fakeQueries{}, nil)

	tests := []struct {
		sym  string
		want float64
	}{
		{"AOG", 33.29},
		{"BNB", 12.5},
		{"XYZ", 8.0},
	}

	for _, tc := range tests {
		got := s.calculateAPY(tc.sym)
		if math.Abs(got-tc.want) > 1e-9 {
			t.Fatalf("calculateAPY(%s) = %v; want %v", tc.sym, got, tc.want)
		}
	}
}

func TestCalculateRewards_Active(t *testing.T) {
	// create a stake that started 10 days ago, amount=100, apy=10%
	start := time.Now().Add(-10 * 24 * time.Hour)
	var amt pgtype.Numeric
	_ = amt.Scan("100")
	var apy pgtype.Numeric
	_ = apy.Scan("10")

	row := db.GetStakeByIDRow{
		Amount:    amt,
		Apy:       apy,
		StartDate: pgtype.Timestamp{Time: start, Valid: true},
		Status:    pgtype.Text{String: "active", Valid: true},
	}

	fq := &fakeQueries{getStakeRow: row}
	s := NewStakingService(fq, nil)

	rewardsStr, err := s.CalculateRewards(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := strconv.ParseFloat(rewardsStr, 64)
	if err != nil {
		t.Fatalf("parse float: %v", err)
	}

	// expected = 100 * (10/365/100) * 10days = 100 * 0.1/365 * 10
	expected := 100.0 * (10.0 / 365.0 / 100.0) * 10.0

	if math.Abs(got-expected) > 1e-9 {
		t.Fatalf("rewards = %v; want %v", got, expected)
	}
}

func TestCalculateRewards_NotActiveOrNoStart(t *testing.T) {
	// not active
	var amt pgtype.Numeric
	_ = amt.Scan("50")
	var apy pgtype.Numeric
	_ = apy.Scan("5")

	rowInactive := db.GetStakeByIDRow{
		Amount:    amt,
		Apy:       apy,
		StartDate: pgtype.Timestamp{Time: time.Now().Add(-1 * time.Hour), Valid: true},
		Status:    pgtype.Text{String: "inactive", Valid: true},
	}
	fq1 := &fakeQueries{getStakeRow: rowInactive}
	s1 := NewStakingService(fq1, nil)
	if _, err := s1.CalculateRewards(context.Background(), uuid.New()); err == nil {
		t.Fatalf("expected error for inactive stake")
	}

	// no start date
	rowNoStart := db.GetStakeByIDRow{
		Amount:    amt,
		Apy:       apy,
		StartDate: pgtype.Timestamp{Valid: false},
		Status:    pgtype.Text{String: "active", Valid: true},
	}
	fq2 := &fakeQueries{getStakeRow: rowNoStart}
	s2 := NewStakingService(fq2, nil)
	if _, err := s2.CalculateRewards(context.Background(), uuid.New()); err == nil {
		t.Fatalf("expected error for stake with no start date")
	}
}
