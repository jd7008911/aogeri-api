-- internal/db/queries/users.sql
-- name: CreateUser :one
INSERT INTO users (email, password_hash, wallet_address)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: GetUserByWallet :one
SELECT * FROM users WHERE wallet_address = $1;

-- name: UpdateUserPassword :exec
UPDATE users 
SET password_hash = $2, updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: UpdateUser2FA :exec
UPDATE users 
SET two_factor_secret = $2, two_factor_enabled = $3, updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: UpdateLoginAttempts :exec
UPDATE users 
SET failed_login_attempts = $2, locked_until = $3, last_login = $4, updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: CreateUserProfile :one
INSERT INTO user_profiles (user_id, username, full_name, country)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetUserProfile :one
SELECT * FROM user_profiles WHERE user_id = $1;

-- name: UpdateUserProfile :exec
UPDATE user_profiles 
SET username = $2, full_name = $3, avatar_url = $4, country = $5, 
    timezone = $6, updated_at = CURRENT_TIMESTAMP
WHERE user_id = $1;