package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims defines the JWT payload structure for Activity sessions.
type Claims struct {
	PlayerID  string `json:"player_id"`
	DiscordID string `json:"discord_id"`
	jwt.RegisteredClaims
}

// JWTManager handles JWT issuance and validation for Activity sessions.
type JWTManager struct {
	signingKey []byte
	issuer     string
}

// NewJWTManager initializes the JWT manager with the signing secret.
func NewJWTManager(secret string) (*JWTManager, error) {
	if secret == "" {
		return nil, fmt.Errorf("JWT secret cannot be empty")
	}
	return &JWTManager{
		signingKey: []byte(secret),
		issuer:     "simlife-bot",
	}, nil
}

// IssueToken creates a new JWT for an authenticated player.
// Tokens are valid for 1 hour and signed with HS256.
func (m *JWTManager) IssueToken(playerID, discordID string) (string, error) {
	now := time.Now()
	claims := Claims{
		PlayerID:  playerID,
		DiscordID: discordID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.issuer,
			Subject:   playerID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(1 * time.Hour)),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(m.signingKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign JWT: %w", err)
	}

	return tokenString, nil
}

// Validate parses and validates a JWT string, returning the claims if valid.
// Returns an error if the token is malformed, expired, or has an invalid signature.
func (m *JWTManager) Validate(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return m.signingKey, nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse JWT: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid JWT claims")
	}

	return claims, nil
}

// RefreshToken issues a new token for an existing valid session.
// This extends the session without requiring re-authentication with Discord.
func (m *JWTManager) RefreshToken(playerID, discordID string) (string, error) {
	// Simply re-issue with a fresh 1-hour expiry
	return m.IssueToken(playerID, discordID)
}