package auth

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"golang.org/x/oauth2"
)

// Atlassian OAuth2 endpoints
var atlassianEndpoint = oauth2.Endpoint{
	AuthURL:  "https://auth.atlassian.com/authorize",
	TokenURL: "https://auth.atlassian.com/oauth/token",
}

// GetHTTPClient returns an http.Client with OAuth2 auto-refresh if refresh_token
// + client credentials are provided, or a static-token client otherwise.
//
// Priority:
//  1. JIRA_REFRESH_TOKEN + JIRA_CLIENT_ID + JIRA_CLIENT_SECRET → auto-refresh
//  2. JIRA_API_TOKEN → static access token (backward compat)
func GetHTTPClient() (*http.Client, error) {
	refreshToken := os.Getenv("JIRA_REFRESH_TOKEN")
	clientID := os.Getenv("JIRA_CLIENT_ID")
	clientSecret := os.Getenv("JIRA_CLIENT_SECRET")

	if refreshToken != "" && clientID != "" && clientSecret != "" {
		return oauthClient(refreshToken, clientID, clientSecret), nil
	}

	// Fallback: direct access token (for long-lived tokens or pre-refreshed)
	token := os.Getenv("JIRA_API_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("JIRA_REFRESH_TOKEN + JIRA_CLIENT_ID + JIRA_CLIENT_SECRET, or JIRA_API_TOKEN environment variable is required")
	}

	return staticTokenClient(token), nil
}

func oauthClient(refreshToken, clientID, clientSecret string) *http.Client {
	cfg := oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     atlassianEndpoint,
	}

	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, &http.Client{
		Timeout: 30 * time.Second,
	})

	ts := cfg.TokenSource(ctx, &oauth2.Token{RefreshToken: refreshToken})
	return oauth2.NewClient(ctx, ts)
}

func staticTokenClient(token string) *http.Client {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	return oauth2.NewClient(context.Background(), ts)
}

// GetToken returns the access token string. Kept for backward compatibility
// with code that needs the raw token. Prefer GetHTTPClient() for auto-refresh.
func GetToken() (string, error) {
	token := os.Getenv("JIRA_API_TOKEN")
	if token == "" {
		return "", fmt.Errorf("JIRA_API_TOKEN environment variable is required")
	}
	return token, nil
}

func GetServerURL() string {
	return os.Getenv("JIRA_SERVER")
}
