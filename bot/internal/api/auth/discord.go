package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/0xProgress/simlife/bot/internal/logger"
)

const (
	discordTokenURL = "https://discord.com/api/v10/oauth2/token"
	discordUserURL  = "https://discord.com/api/v10/users/@me"
)

// DiscordAuth handles the Discord OAuth2 token exchange flow for the Activity.
type DiscordAuth struct {
	appID     string
	clientID  string
	client    *http.Client
}

// NewDiscordAuth initializes the Discord OAuth handler.
// Note: The actual client secret is provided per-call via the exchange method,
// as it may be rotated without restarting the bot.
func NewDiscordAuth(appID, clientID string) *DiscordAuth {
	return &DiscordAuth{
		appID:    appID,
		clientID: clientID,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// TokenResponse represents Discord's OAuth token response.
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
}

// DiscordUser represents the authenticated Discord user.
type DiscordUser struct {
	ID            string `json:"id"`
	Username      string `json:"username"`
	Discriminator string `json:"discriminator"`
	Avatar        string `json:"avatar"`
	Email         string `json:"email,omitempty"`
}

// ExchangeCode exchanges an OAuth authorization code for an access token,
// then retrieves the authenticated user's Discord profile.
func (d *DiscordAuth) ExchangeCode(ctx context.Context, code, redirectURI, clientSecret string) (*DiscordUser, error) {
	log := logger.FromContext(ctx, "auth.discord")

	// 1. Exchange the authorization code for an access token
	tokenResp, err := d.exchangeCodeForToken(ctx, code, redirectURI, clientSecret)
	if err != nil {
		return nil, fmt.Errorf("code exchange failed: %w", err)
	}

	log.Debug().
		Str("token_type", tokenResp.TokenType).
		Int("expires_in", tokenResp.ExpiresIn).
		Msg("discord token exchange successful")

	// 2. Use the access token to fetch the user profile
	user, err := d.fetchDiscordUser(ctx, tokenResp.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch discord user: %w", err)
	}

	log.Info().
		Str("discord_id", user.ID).
		Str("username", user.Username).
		Msg("discord user authenticated")

	return user, nil
}

// exchangeCodeForToken performs the OAuth2 code exchange with Discord.
func (d *DiscordAuth) exchangeCodeForToken(ctx context.Context, code, redirectURI, clientSecret string) (*TokenResponse, error) {
	data := url.Values{
		"client_id":     {d.clientID},
		"client_secret": {clientSecret},
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"redirect_uri":  {redirectURI},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, discordTokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := d.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("token request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read token response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("discord token endpoint returned status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	if tokenResp.AccessToken == "" {
		return nil, fmt.Errorf("discord token response missing access_token")
	}

	return &tokenResp, nil
}

// fetchDiscordUser retrieves the authenticated user's profile from Discord.
func (d *DiscordAuth) fetchDiscordUser(ctx context.Context, accessToken string) (*DiscordUser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, discordUserURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create user request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := d.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("user request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read user response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("discord user endpoint returned status %d: %s", resp.StatusCode, string(body))
	}

	var user DiscordUser
	if err := json.Unmarshal(body, &user); err != nil {
		return nil, fmt.Errorf("failed to parse user response: %w", err)
	}

	if user.ID == "" {
		return nil, fmt.Errorf("discord user response missing id")
	}

	return &user, nil
}