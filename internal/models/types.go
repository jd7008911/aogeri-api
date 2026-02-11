// internal/models/types.go
package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID                  uuid.UUID  `json:"id"`
	Email               string     `json:"email"`
	WalletAddress       *string    `json:"wallet_address,omitempty"`
	TwoFactorEnabled    bool       `json:"two_factor_enabled"`
	IsActive            bool       `json:"is_active"`
	FailedLoginAttempts int32      `json:"-"`
	LockedUntil         *time.Time `json:"-"`
	LastLogin           *time.Time `json:"last_login"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

type Stake struct {
	ID             uuid.UUID  `json:"id"`
	UserID         uuid.UUID  `json:"user_id"`
	TokenSymbol    string     `json:"token_symbol"`
	Amount         string     `json:"amount"`
	APY            float64    `json:"apy"`
	StartDate      time.Time  `json:"start_date"`
	EndDate        *time.Time `json:"end_date,omitempty"`
	Status         string     `json:"status"`
	AutoCompound   bool       `json:"auto_compound"`
	RewardsClaimed string     `json:"rewards_claimed"`
}

type Proposal struct {
	ID           uuid.UUID  `json:"id"`
	Title        string     `json:"title"`
	Description  string     `json:"description"`
	ProposerID   uuid.UUID  `json:"proposer_id"`
	Type         string     `json:"type"`
	Status       string     `json:"status"`
	VotingStart  *time.Time `json:"voting_start,omitempty"`
	VotingEnd    time.Time  `json:"voting_end"`
	Quorum       float64    `json:"quorum"`
	Threshold    float64    `json:"threshold"`
	ForVotes     string     `json:"for_votes"`
	AgainstVotes string     `json:"against_votes"`
	AbstainVotes string     `json:"abstain_votes"`
}

type Asset struct {
	ID               uuid.UUID `json:"id"`
	Symbol           string    `json:"symbol"`
	Name             string    `json:"name"`
	CurrentValue     string    `json:"current_value"`
	PriceChange24H   float64   `json:"price_change_24h"`
	Volume24H        string    `json:"volume_24h"`
	TotalValueLocked string    `json:"total_value_locked,omitempty"`
}

type DashboardStats struct {
	TotalValueLocked    string  `json:"total_value_locked"`
	ActiveMonitors      int32   `json:"active_monitors"`
	RemainingTime       string  `json:"remaining_time"`
	ActiveStakes        int32   `json:"active_stakes"`
	TotalRewards        string  `json:"total_rewards"`
	SecurityScore       float64 `json:"security_score"`
	GovernanceProposals int32   `json:"governance_proposals"`
}

// Request/Response types
type RegisterRequest struct {
	Email           string `json:"email" validate:"required,email"`
	Password        string `json:"password" validate:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=Password"`
	WalletAddress   string `json:"wallet_address,omitempty"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	User         *User  `json:"user"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"`
}

type StakeRequest struct {
	TokenSymbol  string `json:"token_symbol" validate:"required"`
	Amount       string `json:"amount" validate:"required,numeric"`
	AutoCompound bool   `json:"auto_compound"`
	DurationDays int    `json:"duration_days" validate:"min=30"`
}

type VoteRequest struct {
	ProposalID uuid.UUID `json:"proposal_id" validate:"required"`
	VoteChoice string    `json:"vote_choice" validate:"required,oneof=for against abstain"`
	VotePower  string    `json:"vote_power" validate:"required,numeric"`
}
