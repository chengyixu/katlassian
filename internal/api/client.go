package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const oauthBaseURL = "https://api.atlassian.com"

type Client struct {
	token      string
	serverURL  string
	cloudID    string
	httpClient *http.Client
}

type AccessibleResource struct {
	ID        string   `json:"id"`
	URL       string   `json:"url"`
	Name      string   `json:"name"`
	Scopes    []string `json:"scopes"`
	AvatarURL string   `json:"avatarUrl"`
}

func NewClient(token string, serverURL string) (*Client, error) {
	c := &Client{
		token:      token,
		serverURL:  serverURL,
		httpClient: &http.Client{},
	}

	// For OAuth tokens, discover the cloud ID
	resources, err := c.getAccessibleResources()
	if err != nil {
		return nil, fmt.Errorf("getting accessible resources: %w", err)
	}

	if len(resources) == 0 {
		return nil, fmt.Errorf("no accessible Atlassian resources found for this token")
	}

	// Use the first resource, or match by server URL
	c.cloudID = resources[0].ID
	if serverURL != "" {
		for _, r := range resources {
			if r.URL == serverURL {
				c.cloudID = r.ID
				break
			}
		}
	}

	return c, nil
}

func (c *Client) getAccessibleResources() ([]AccessibleResource, error) {
	req, err := http.NewRequest("GET", oauthBaseURL+"/oauth/token/accessible-resources", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	var resources []AccessibleResource
	if err := json.Unmarshal(body, &resources); err != nil {
		return nil, err
	}

	return resources, nil
}

func (c *Client) apiURL(path string) string {
	return fmt.Sprintf("%s/ex/jira/%s/rest/api/3%s", oauthBaseURL, c.cloudID, path)
}

func (c *Client) Get(path string) (json.RawMessage, error) {
	req, err := http.NewRequest("GET", c.apiURL(path), nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error (HTTP %d): %s", resp.StatusCode, string(body))
	}

	return body, nil
}

func (c *Client) Post(path string, payload interface{}) (json.RawMessage, error) {
	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	req, err := http.NewRequest("POST", c.apiURL(path), bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error (HTTP %d): %s", resp.StatusCode, string(body))
	}

	return body, nil
}

func (c *Client) Put(path string, payload interface{}) (json.RawMessage, error) {
	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	req, err := http.NewRequest("PUT", c.apiURL(path), bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error (HTTP %d): %s", resp.StatusCode, string(body))
	}

	return body, nil
}
