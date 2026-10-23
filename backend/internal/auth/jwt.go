package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/Dimonchester/ttk-room-map/backend/internal/user"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token expired")
)

type TokenType string

const (
	TokenTypeAccess  TokenType = "access"
	TokenTypeRefresh TokenType = "refresh"
)

type Claims struct {
	Sub       string    `json:"sub"`
	Email     string    `json:"email"`
	Role      user.Role `json:"role"`
	TokenType TokenType `json:"type"`
	IssuedAt  int64     `json:"iat"`
	ExpiresAt int64     `json:"exp"`
}

type TokenManager struct {
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewTokenManager(secret string, accessTTL, refreshTTL time.Duration) *TokenManager {
	if secret == "" {
		secret = "ttk-default-dev-secret-key"
	}
	return &TokenManager{
		secret:     []byte(secret),
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

func (tm *TokenManager) AccessTTLSeconds() int {
	return int(tm.accessTTL.Seconds())
}

func (tm *TokenManager) RefreshTTL() time.Duration {
	return tm.refreshTTL
}

func (tm *TokenManager) GenerateAccessToken(u *user.User) (string, error) {
	return tm.generate(u, TokenTypeAccess, tm.accessTTL)
}

func (tm *TokenManager) GenerateRefreshToken(u *user.User) (string, error) {
	return tm.generate(u, TokenTypeRefresh, tm.refreshTTL)
}

func (tm *TokenManager) generate(u *user.User, tokenType TokenType, ttl time.Duration) (string, error) {
	now := time.Now()
	headerJSON := []byte(`{"alg":"HS256","typ":"JWT"}`)

	claims := Claims{
		Sub:       u.ID,
		Email:     u.Email,
		Role:      u.Role,
		TokenType: tokenType,
		IssuedAt:  now.Unix(),
		ExpiresAt: now.Add(ttl).Unix(),
	}

	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	headerEncoded := base64.RawURLEncoding.EncodeToString(headerJSON)
	claimsEncoded := base64.RawURLEncoding.EncodeToString(claimsJSON)
	signingInput := headerEncoded + "." + claimsEncoded

	sig := tm.sign(signingInput)
	return signingInput + "." + sig, nil
}

func (tm *TokenManager) Verify(tokenStr string, expectedType TokenType) (*Claims, error) {
	parts := strings.Split(tokenStr, ".")
	if len(parts) != 3 {
		return nil, ErrInvalidToken
	}

	signingInput := parts[0] + "." + parts[1]
	expectedSig := tm.sign(signingInput)
	if !hmac.Equal([]byte(parts[2]), []byte(expectedSig)) {
		return nil, ErrInvalidToken
	}

	claimsJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, ErrInvalidToken
	}

	var claims Claims
	if err := json.Unmarshal(claimsJSON, &claims); err != nil {
		return nil, ErrInvalidToken
	}

	if claims.TokenType != expectedType {
		return nil, ErrInvalidToken
	}

	if time.Now().Unix() > claims.ExpiresAt {
		return nil, ErrExpiredToken
	}

	return &claims, nil
}

func (tm *TokenManager) sign(input string) string {
	mac := hmac.New(sha256.New, tm.secret)
	mac.Write([]byte(input))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
