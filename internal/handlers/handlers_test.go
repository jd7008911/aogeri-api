package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jd7008911/aogeri-api/internal/auth"
	"github.com/jd7008911/aogeri-api/internal/db"
)

// fakeDBTX implements db.DBTX for tests and records calls.
type fakeDBTX struct {
	// stored values to return
	profile db.UserProfile
	user    db.User

	// recorded
	updatedPassword   string
	updated2fa        string
	updated2faEnabled bool
	updatedRewards    string
}

type fakeRow struct {
	scanFn func(dest ...interface{}) error
}

func (r *fakeRow) Scan(dest ...interface{}) error { return r.scanFn(dest...) }

func (f *fakeDBTX) Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
	// detect which exec is being called by examining SQL
	if strings.Contains(sql, "UPDATE users \nSET password_hash") {
		if len(args) >= 2 {
			if ph, ok := args[1].(string); ok {
				f.updatedPassword = ph
			}
		}
	}
	if strings.Contains(sql, "UPDATE users \nSET two_factor_secret") {
		if len(args) >= 3 {
			if tfs, ok := args[1].(pgtype.Text); ok && tfs.Valid {
				f.updated2fa = tfs.String
			}
			if tfe, ok := args[2].(pgtype.Bool); ok {
				f.updated2faEnabled = tfe.Bool
			}
		}
	}
	if strings.Contains(sql, "UPDATE stakes \nSET rewards_claimed") {
		if len(args) >= 2 {
			if rn, ok := args[1].(pgtype.Numeric); ok {
				if v, err := rn.Float64Value(); err == nil {
					f.updatedRewards = strconv.FormatFloat(v.Float64, 'f', -1, 64)
				}
			}
		}
	}
	return pgconn.CommandTag{}, nil
}

func (f *fakeDBTX) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	return nil, nil
}

func (f *fakeDBTX) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	// return appropriate fakeRow that scans into expected destinations
	if strings.Contains(sql, "FROM user_profiles") {
		return &fakeRow{scanFn: func(dest ...interface{}) error {
			// id, user_id, username, full_name, avatar_url, country, timezone, notifications_enabled, created_at, updated_at
			if len(dest) >= 10 {
				if idPtr, ok := dest[0].(*pgtype.UUID); ok {
					*idPtr = f.profile.ID
				}
				if uidPtr, ok := dest[1].(*pgtype.UUID); ok {
					*uidPtr = f.profile.UserID
				}
				if unPtr, ok := dest[2].(*pgtype.Text); ok {
					*unPtr = f.profile.Username
				}
				if fnPtr, ok := dest[3].(*pgtype.Text); ok {
					*fnPtr = f.profile.FullName
				}
				if avPtr, ok := dest[4].(*pgtype.Text); ok {
					*avPtr = f.profile.AvatarUrl
				}
				if cPtr, ok := dest[5].(*pgtype.Text); ok {
					*cPtr = f.profile.Country
				}
				if tzPtr, ok := dest[6].(*pgtype.Text); ok {
					*tzPtr = f.profile.Timezone
				}
				if nPtr, ok := dest[7].(*pgtype.Bool); ok {
					*nPtr = f.profile.NotificationsEnabled
				}
				if caPtr, ok := dest[8].(*pgtype.Timestamp); ok {
					*caPtr = f.profile.CreatedAt
				}
				if upPtr, ok := dest[9].(*pgtype.Timestamp); ok {
					*upPtr = f.profile.UpdatedAt
				}
			}
			return nil
		}}
	}
	if strings.Contains(sql, "FROM users WHERE id") {
		return &fakeRow{scanFn: func(dest ...interface{}) error {
			// id, email, password_hash, wallet_address, two_factor_secret, two_factor_enabled, is_active, failed_login_attempts, locked_until, last_login, created_at, updated_at
			if len(dest) >= 12 {
				if idPtr, ok := dest[0].(*pgtype.UUID); ok {
					*idPtr = f.user.ID
				}
				if emailPtr, ok := dest[1].(*string); ok {
					*emailPtr = f.user.Email
				}
				if phPtr, ok := dest[2].(*string); ok {
					*phPtr = f.user.PasswordHash
				}
				if waPtr, ok := dest[3].(*pgtype.Text); ok {
					*waPtr = f.user.WalletAddress
				}
				if tfPtr, ok := dest[4].(*pgtype.Text); ok {
					*tfPtr = f.user.TwoFactorSecret
				}
				if tfePtr, ok := dest[5].(*pgtype.Bool); ok {
					*tfePtr = f.user.TwoFactorEnabled
				}
				if iaPtr, ok := dest[6].(*pgtype.Bool); ok {
					*iaPtr = f.user.IsActive
				}
				if flaPtr, ok := dest[7].(*pgtype.Int4); ok {
					*flaPtr = f.user.FailedLoginAttempts
				}
				if luPtr, ok := dest[8].(*pgtype.Timestamp); ok {
					*luPtr = f.user.LockedUntil
				}
				if llPtr, ok := dest[9].(*pgtype.Timestamp); ok {
					*llPtr = f.user.LastLogin
				}
				if caPtr, ok := dest[10].(*pgtype.Timestamp); ok {
					*caPtr = f.user.CreatedAt
				}
				if upPtr, ok := dest[11].(*pgtype.Timestamp); ok {
					*upPtr = f.user.UpdatedAt
				}
			}
			return nil
		}}
	}
	// default: empty scan
	return &fakeRow{scanFn: func(dest ...interface{}) error { return nil }}
}

// helper converters
func mustPgUUID(id uuid.UUID) pgtype.UUID {
	var p pgtype.UUID
	copy(p.Bytes[:], id[:])
	p.Valid = true
	return p
}
func mustText(s string) pgtype.Text {
	if s == "" {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: s, Valid: true}
}

func TestUpdateProfileHandler(t *testing.T) {
	uid := uuid.New()
	fdb := &fakeDBTX{}
	fdb.profile = db.UserProfile{
		ID:                   mustPgUUID(uuid.New()),
		UserID:               mustPgUUID(uid),
		Username:             mustText("olduser"),
		FullName:             mustText("Old Name"),
		AvatarUrl:            mustText(""),
		Country:              mustText(""),
		Timezone:             mustText("UTC"),
		NotificationsEnabled: pgtype.Bool{Bool: true, Valid: true},
		CreatedAt:            pgtype.Timestamp{Time: time.Now(), Valid: true},
		UpdatedAt:            pgtype.Timestamp{Time: time.Now(), Valid: true},
	}

	queries := db.New(fdb)
	h := NewAuthHandler(nil, queries)

	body := `{"username":"newuser","full_name":"New Name","timezone":"PST"}`
	req := httptest.NewRequest(http.MethodPut, "/auth/profile", strings.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), auth.UserIDKey, uid))
	rr := httptest.NewRecorder()

	h.UpdateProfile(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d body=%s", rr.Code, rr.Body.String())
	}

	var got db.UserProfile
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("json unmarshal: %v", err)
	}
}
