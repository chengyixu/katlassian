package auth

import (
	"fmt"
	"os"
)

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
