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
	"github.com/jd7008911/aogeri-api/internal/models"
)

type fakeQueries struct {
	tokens        []db.GetTokenListRow
	created       db.Stake
	createdCalled bool
	getStakeRow   db.GetStakeByIDRow
	userStakes    []db.GetUserStakesRow
	unstakeCalled bool
}

func (f *fakeQueries) GetTokenList(ctx context.Context) ([]db.GetTokenListRow, error) {
	return f.tokens, nil
}
func (f *fakeQueries) CreateStake(ctx context.Context, arg db.CreateStakeParams) (db.Stake, error) {
	f.createdCalled = true
	return f.created, nil
}
func (f *fakeQueries) GetStakeByID(ctx context.Context, id pgtype.UUID) (db.GetStakeByIDRow, error) {
	return f.getStakeRow, nil
}
func (f *fakeQueries) GetUserStakes(ctx context.Context, userID pgtype.UUID) ([]db.GetUserStakesRow, error) {
	return f.userStakes, nil
}
func (f *fakeQueries) Unstake(ctx context.Context, arg db.UnstakeParams) error {
	f.unstakeCalled = true
	return nil
}

func TestCalculateAPY(t *testing.T) {
	s := NewStakingService(&fakeQueries{}, nil)

	tests := []struct {
		sym  string
		want float64
	}{
		{"AOG", 33.29}, {"BNB", 12.5}, {"XYZ", 8.0},
	}

	for _, tc := range tests {
		got := s.calculateAPY(tc.sym)
		if math.Abs(got-tc.want) > 1e-9 {
			t.Fatalf("calculateAPY(%s) = %v; want %v", tc.sym, got, tc.want)
		}
	}
}

func TestCalculateRewards_Active(t *testing.T) {
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

	expected := 100.0 * (10.0 / 365.0 / 100.0) * 10.0
	if math.Abs(got-expected) > 1e-9 {
		t.Fatalf("rewards = %v; want %v", got, expected)
	}
}

func TestCalculateRewards_NotActiveOrNoStart(t *testing.T) {
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

	rowNoStart := db.GetStakeByIDRow{Amount: amt, Apy: apy, StartDate: pgtype.Timestamp{Valid: false}, Status: pgtype.Text{String: "active", Valid: true}}
	fq2 := &fakeQueries{getStakeRow: rowNoStart}
	s2 := NewStakingService(fq2, nil)
	if _, err := s2.CalculateRewards(context.Background(), uuid.New()); err == nil {
		t.Fatalf("expected error for stake with no start date")
	}
}

func TestCreateStake_TokenNotFound(t *testing.T) {
	s := NewStakingService(&fakeQueries{tokens: []db.GetTokenListRow{}}, nil)
	_, err := s.CreateStake(context.Background(), uuid.New(), models.StakeRequest{TokenSymbol: "NOPE", Amount: "1", DurationDays: 30})
	if err == nil {
		t.Fatalf("expected token not found error")
	}
}

func TestCreateStake_Success(t *testing.T) {
	tid := uuid.New()
	var tidPg pgtype.UUID
	copy(tidPg.Bytes[:], tid[:])
	tidPg.Valid = true
	tokens := []db.GetTokenListRow{{ID: tidPg, Symbol: "AOG"}}

	sid := uuid.New()
	uid := uuid.New()
	var sidPg, uidPg pgtype.UUID
	copy(sidPg.Bytes[:], sid[:])
	sidPg.Valid = true
	copy(uidPg.Bytes[:], uid[:])
	uidPg.Valid = true

	var amt pgtype.Numeric
	_ = amt.Scan("123.45")
	var apy pgtype.Numeric
	_ = apy.Scan("33.29")

	start := time.Now()
	end := start.Add(30 * 24 * time.Hour)
	created := db.Stake{
		ID: sidPg, UserID: uidPg, TokenID: tidPg, Amount: amt, Apy: apy,
		StartDate: pgtype.Timestamp{Time: start, Valid: true}, EndDate: pgtype.Timestamp{Time: end, Valid: true},
		Status: pgtype.Text{String: "active", Valid: true}, AutoCompound: pgtype.Bool{Bool: true, Valid: true},
		RewardsClaimed: func() pgtype.Numeric { var n pgtype.Numeric; _ = n.Scan("0"); return n }(),
	}

	fq := &fakeQueries{tokens: tokens, created: created}
	s := NewStakingService(fq, nil)

	got, err := s.CreateStake(context.Background(), uid, models.StakeRequest{TokenSymbol: "AOG", Amount: "123.45", DurationDays: 30, AutoCompound: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.TokenSymbol != "AOG" || got.Amount != "123.45" {
		t.Fatalf("unexpected stake returned: %+v", got)
	}
}

func TestGetUserStakes_GetStakeByID_Unstake(t *testing.T) {
	var amt pgtype.Numeric
	_ = amt.Scan("50")
	var apy pgtype.Numeric
	_ = apy.Scan("5")
	start := time.Now().Add(-5 * 24 * time.Hour)
	end := start.Add(25 * 24 * time.Hour)

	uid := uuid.New()
	var uidPg pgtype.UUID
	copy(uidPg.Bytes[:], uid[:])
	uidPg.Valid = true
	var stakeID uuid.UUID = uuid.New()
	var stakeIDPg pgtype.UUID
	copy(stakeIDPg.Bytes[:], stakeID[:])
	stakeIDPg.Valid = true

	userRow := db.GetUserStakesRow{
		ID: stakeIDPg, UserID: uidPg, Amount: amt, Apy: apy,
		StartDate: pgtype.Timestamp{Time: start, Valid: true}, EndDate: pgtype.Timestamp{Time: end, Valid: true},
		Status: pgtype.Text{String: "active", Valid: true}, Symbol: "AOG",
	}

	fq := &fakeQueries{userStakes: []db.GetUserStakesRow{userRow}, getStakeRow: db.GetStakeByIDRow(userRow)}
	s := NewStakingService(fq, nil)

	list, err := s.GetUserStakes(context.Background(), uid)
	if err != nil {
		t.Fatalf("GetUserStakes error: %v", err)
	}
	if len(list) != 1 || list[0].Amount == "0" {
		t.Fatalf("unexpected user stakes: %+v", list)
	}

	st, err := s.GetStakeByID(context.Background(), stakeID)
	if err != nil {
		t.Fatalf("GetStakeByID error: %v", err)
	}
	if st.ID != stakeID || st.TokenSymbol != "AOG" {
		t.Fatalf("unexpected stake: %+v", st)
	}

	if err := s.Unstake(context.Background(), stakeID, uid); err != nil {
		t.Fatalf("unstake error: %v", err)
	}
	if !fq.unstakeCalled {
		t.Fatalf("expected Unstake to be called on querier")
	}
}
