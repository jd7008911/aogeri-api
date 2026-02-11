package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func TestHashAndCheckPassword(t *testing.T) {
	pw := "Str0ng!Pass"
	h, err := HashPassword(pw)
	if err != nil {
		t.Fatalf("hash error: %v", err)
	}
	if !CheckPasswordHash(pw, h) {
		t.Fatalf("password did not verify")
	}
	if CheckPasswordHash("wrong", h) {
		t.Fatalf("expected wrong password to fail")
	}
}

func TestValidatePasswordStrength(t *testing.T) {
	cases := []struct {
		pw string
		ok bool
	}{
		{"short", false},
		{"alllowercase1!", false},
		{"ALLUPPER1!", false},
		{"NoNumber!", false},
		{"NoSpecial1A", false},
		{"Good1!Abc", true},
	}
	for _, c := range cases {
		err := ValidatePasswordStrength(c.pw)
		if (err == nil) != c.ok {
			t.Fatalf("pw %q expected ok=%v got err=%v", c.pw, c.ok, err)
		}
	}
}

func TestExtractBearerToken(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if _, err := ExtractBearerToken(req); err == nil {
		t.Fatalf("expected error when no header present")
	}

	req.Header.Set("Authorization", "BadFormat token")
	if _, err := ExtractBearerToken(req); err == nil {
		t.Fatalf("expected error for bad header format")
	}

	req.Header.Set("Authorization", "Bearer secret-token")
	tok, err := ExtractBearerToken(req)
	if err != nil || tok != "secret-token" {
		t.Fatalf("unexpected token extraction: %v %v", tok, err)
	}
}

func TestSignAndParseAccessToken(t *testing.T) {
	secret := "s3cr3t"
	uid := uuid.New()
	claims := NewClaims(uid, "me@example.com", "test", time.Minute*5)

	tokStr, err := SignAccessToken(secret, claims)
	if err != nil {
		t.Fatalf("sign error: %v", err)
	}

	parsed, err := ParseAccessToken(tokStr, secret)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if parsed.UserID != uid || parsed.Email != "me@example.com" {
		t.Fatalf("claims mismatch: got %+v want uid %v", parsed, uid)
	}

	// ensure ParseAccessToken rejects tampered token
	tampered := tokStr + "x"
	if _, err := ParseAccessToken(tampered, secret); err == nil {
		t.Fatalf("expected parse error for tampered token")
	}

	// ensure ParseAccessToken rejects wrong secret
	if _, err := ParseAccessToken(tokStr, "wrong"); err == nil {
		t.Fatalf("expected parse error for wrong secret")
	}

	// check jwt package interoperability: parse with jwt.ParseWithClaims
	c := &Claims{}
	if _, err := jwt.ParseWithClaims(tokStr, c, func(token *jwt.Token) (interface{}, error) { return []byte(secret), nil }); err != nil {
		t.Fatalf("jwt parse direct failed: %v", err)
	}
}

func TestTokenHashDeterministic(t *testing.T) {
	token := "tok"
	secret := "sec"
	// use same algorithm as hashToken
	h := sha256.Sum256([]byte(token + secret))
	want := hex.EncodeToString(h[:])
	got := hashToken(token, secret)
	if got != want {
		t.Fatalf("hash mismatch got=%s want=%s", got, want)
	}
}

func TestUUIDPgConversions(t *testing.T) {
	id := uuid.New()
	pg, err := uuidToPgUUID(id)
	if err != nil {
		t.Fatalf("uuidToPgUUID err: %v", err)
	}
	back, err := pgUUIDToUUID(pg)
	if err != nil {
		t.Fatalf("pgUUIDToUUID err: %v", err)
	}
	if back != id {
		t.Fatalf("uuid conversion mismatch: %v vs %v", back, id)
	}
}
