-- internal/db/queries/governance.sql
-- name: CreateProposal :one
INSERT INTO governance_proposals (
    title, description, proposer_id, proposal_type, 
    voting_end, quorum, threshold
) VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetActiveProposals :many
SELECT * FROM governance_proposals 
WHERE status = 'active' AND voting_end > CURRENT_TIMESTAMP
ORDER BY created_at DESC;

-- name: GetProposalByID :one
SELECT * FROM governance_proposals WHERE id = $1;

-- name: CastVote :one
INSERT INTO user_votes (user_id, proposal_id, vote_power, vote_choice)
VALUES ($1, $2, $3, $4)
ON CONFLICT (user_id, proposal_id) 
DO UPDATE SET vote_power = EXCLUDED.vote_power, vote_choice = EXCLUDED.voice_choice
RETURNING *;

-- name: UpdateProposalVotes :exec
UPDATE governance_proposals 
SET 
    for_votes = $2,
    against_votes = $3,
    abstain_votes = $4,
    total_votes = $5,
    status = $6,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: GetUserVotes :many
SELECT * FROM user_votes WHERE user_id = $1;