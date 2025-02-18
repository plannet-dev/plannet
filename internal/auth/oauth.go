package auth

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/oauth2"
)

type OAuthToken struct {
	AccessToken  string    `json:"access_token"`
	TokenType    string    `json:"token_type"`
	RefreshToken string    `json:"refresh_token"`
	Expiry       time.Time `json:"expiry"`
}

type OAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	AuthURL      string
	TokenURL     string
	Scopes       []string
}

// OAuthManager handles OAuth authentication flow
type OAuthManager struct {
	config *oauth2.Config
}

func NewOAuthManager(cfg OAuthConfig) *OAuthManager {
	return &OAuthManager{
		config: &oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			RedirectURL:  cfg.RedirectURL,
			Endpoint: oauth2.Endpoint{
				AuthURL:  cfg.AuthURL,
				TokenURL: cfg.TokenURL,
			},
			Scopes: cfg.Scopes,
		},
	}
}

// StartOAuthFlow begins the OAuth process and returns the auth URL
func (m *OAuthManager) StartOAuthFlow() (string, error) {
	// Generate random state
	state := generateRandomState()

	// Generate the authorization URL
	url := m.config.AuthCodeURL(state)

	return url, nil
}

// CompleteOAuthFlow handles the OAuth callback and returns the token
func (m *OAuthManager) CompleteOAuthFlow(code string) (*OAuthToken, error) {
	ctx := context.Background()
	token, err := m.config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("token exchange failed: %w", err)
	}

	return &OAuthToken{
		AccessToken:  token.AccessToken,
		TokenType:    token.TokenType,
		RefreshToken: token.RefreshToken,
		Expiry:       token.Expiry,
	}, nil
}

// internal helper functions
func generateRandomState() string {
	// Implement secure random state generation
	return "random-state"
}
