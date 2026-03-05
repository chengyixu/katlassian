package main

import (
	"fmt"
	"net/url"
	"os"
	"strconv"

	"github.com/minervacap2022/klik-atlassian-cli/internal/api"
	"github.com/minervacap2022/klik-atlassian-cli/internal/auth"
	"github.com/minervacap2022/klik-atlassian-cli/internal/output"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "katlassian",
	Short: "Atlassian/Jira CLI for KLIK platform",
	Long:  "Command-line interface for Atlassian Jira REST API via OAuth. Auth via JIRA_API_TOKEN env var.",
}

func getClient() *api.Client {
	token, err := auth.GetToken()
	if err != nil {
		output.Error(err.Error())
		os.Exit(1)
	}
	serverURL := auth.GetServerURL()

	client, err := api.NewClient(token, serverURL)
	if err != nil {
		output.Error(err.Error())
		os.Exit(1)
	}
	return client
}

// --- issue commands ---

var issueCmd = &cobra.Command{
	Use:   "issue",
	Short: "Manage Jira issues",
}

var issueListCmd = &cobra.Command{
	Use:   "list",
	Short: "List issues",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := getClient()
		project, _ := cmd.Flags().GetString("project")
		jql, _ := cmd.Flags().GetString("jql")
		maxResults, _ := cmd.Flags().GetInt("limit")

		var searchJQL string
		if jql != "" {
			searchJQL = jql
		} else if project != "" {
			searchJQL = fmt.Sprintf("project = %s ORDER BY updated DESC", project)
		} else {
			searchJQL = "ORDER BY updated DESC"
		}

		path := fmt.Sprintf("/search?jql=%s&maxResults=%d&fields=summary,status,assignee,priority,created,updated",
			url.QueryEscape(searchJQL), maxResults)

		result, err := client.Get(path)
		if err != nil {
			return err
		}
		output.RawJSON(result)
		return nil
	},
}

var issueViewCmd = &cobra.Command{
	Use:   "view",
	Short: "View an issue",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := getClient()
		key, _ := cmd.Flags().GetString("key")

		result, err := client.Get(fmt.Sprintf("/issue/%s", key))
		if err != nil {
			return err
		}
		output.RawJSON(result)
		return nil
	},
}

var issueCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an issue",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := getClient()
		project, _ := cmd.Flags().GetString("project")
		issueType, _ := cmd.Flags().GetString("type")
		summary, _ := cmd.Flags().GetString("summary")
		description, _ := cmd.Flags().GetString("description")

		payload := map[string]interface{}{
			"fields": map[string]interface{}{
				"project": map[string]string{
					"key": project,
				},
				"issuetype": map[string]string{
					"name": issueType,
				},
				"summary": summary,
			},
		}

		if description != "" {
			payload["fields"].(map[string]interface{})["description"] = map[string]interface{}{
				"type":    "doc",
				"version": 1,
				"content": []map[string]interface{}{
					{
						"type": "paragraph",
						"content": []map[string]interface{}{
							{
								"type": "text",
								"text": description,
							},
						},
					},
				},
			}
		}

		result, err := client.Post("/issue", payload)
		if err != nil {
			return err
		}
		output.RawJSON(result)
		return nil
	},
}

var issueTransitionCmd = &cobra.Command{
	Use:   "transition",
	Short: "Transition an issue to a new status",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := getClient()
		key, _ := cmd.Flags().GetString("key")
		transitionID, _ := cmd.Flags().GetString("transition-id")

		// If no transition ID, list available transitions
		if transitionID == "" {
			result, err := client.Get(fmt.Sprintf("/issue/%s/transitions", key))
			if err != nil {
				return err
			}
			output.RawJSON(result)
			return nil
		}

		payload := map[string]interface{}{
			"transition": map[string]string{
				"id": transitionID,
			},
		}

		_, err := client.Post(fmt.Sprintf("/issue/%s/transitions", key), payload)
		if err != nil {
			return err
		}
		output.JSON(map[string]interface{}{"ok": true, "message": fmt.Sprintf("Issue %s transitioned", key)})
		return nil
	},
}

var issueCommentCmd = &cobra.Command{
	Use:   "comment",
	Short: "Add a comment to an issue",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := getClient()
		key, _ := cmd.Flags().GetString("key")
		text, _ := cmd.Flags().GetString("text")

		payload := map[string]interface{}{
			"body": map[string]interface{}{
				"type":    "doc",
				"version": 1,
				"content": []map[string]interface{}{
					{
						"type": "paragraph",
						"content": []map[string]interface{}{
							{
								"type": "text",
								"text": text,
							},
						},
					},
				},
			},
		}

		result, err := client.Post(fmt.Sprintf("/issue/%s/comment", key), payload)
		if err != nil {
			return err
		}
		output.RawJSON(result)
		return nil
	},
}

// --- project commands ---

var projectCmd = &cobra.Command{
	Use:   "project",
	Short: "Manage Jira projects",
}

var projectListCmd = &cobra.Command{
	Use:   "list",
	Short: "List projects",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := getClient()
		limit, _ := cmd.Flags().GetInt("limit")

		result, err := client.Get(fmt.Sprintf("/project/search?maxResults=%d", limit))
		if err != nil {
			return err
		}
		output.RawJSON(result)
		return nil
	},
}

// --- search commands ---

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search issues using JQL",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := getClient()
		jql, _ := cmd.Flags().GetString("jql")
		maxResults, _ := cmd.Flags().GetInt("limit")

		path := fmt.Sprintf("/search?jql=%s&maxResults=%d",
			url.QueryEscape(jql), maxResults)

		result, err := client.Get(path)
		if err != nil {
			return err
		}
		output.RawJSON(result)
		return nil
	},
}

// --- user commands ---

var userCmd = &cobra.Command{
	Use:   "user",
	Short: "Manage users",
}

var userMeCmd = &cobra.Command{
	Use:   "me",
	Short: "Get current user info",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := getClient()
		result, err := client.Get("/myself")
		if err != nil {
			return err
		}
		output.RawJSON(result)
		return nil
	},
}

var userSearchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search users",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := getClient()
		query, _ := cmd.Flags().GetString("query")
		maxResults, _ := cmd.Flags().GetInt("limit")

		path := fmt.Sprintf("/user/search?query=%s&maxResults=%s",
			url.QueryEscape(query), strconv.Itoa(maxResults))

		result, err := client.Get(path)
		if err != nil {
			return err
		}
		output.RawJSON(result)
		return nil
	},
}

func init() {
	// issue
	issueListCmd.Flags().String("project", "", "Project key")
	issueListCmd.Flags().String("jql", "", "JQL query")
	issueListCmd.Flags().Int("limit", 25, "Max results")

	issueViewCmd.Flags().String("key", "", "Issue key (e.g., PROJ-123)")
	issueViewCmd.MarkFlagRequired("key")

	issueCreateCmd.Flags().String("project", "", "Project key")
	issueCreateCmd.Flags().String("type", "Task", "Issue type")
	issueCreateCmd.Flags().String("summary", "", "Issue summary")
	issueCreateCmd.Flags().String("description", "", "Issue description")
	issueCreateCmd.MarkFlagRequired("project")
	issueCreateCmd.MarkFlagRequired("summary")

	issueTransitionCmd.Flags().String("key", "", "Issue key")
	issueTransitionCmd.Flags().String("transition-id", "", "Transition ID (omit to list)")
	issueTransitionCmd.MarkFlagRequired("key")

	issueCommentCmd.Flags().String("key", "", "Issue key")
	issueCommentCmd.Flags().String("text", "", "Comment text")
	issueCommentCmd.MarkFlagRequired("key")
	issueCommentCmd.MarkFlagRequired("text")

	issueCmd.AddCommand(issueListCmd, issueViewCmd, issueCreateCmd, issueTransitionCmd, issueCommentCmd)

	// project
	projectListCmd.Flags().Int("limit", 25, "Max projects")
	projectCmd.AddCommand(projectListCmd)

	// search
	searchCmd.Flags().String("jql", "", "JQL query")
	searchCmd.Flags().Int("limit", 25, "Max results")
	searchCmd.MarkFlagRequired("jql")

	// user
	userSearchCmd.Flags().String("query", "", "Search query")
	userSearchCmd.Flags().Int("limit", 10, "Max results")
	userSearchCmd.MarkFlagRequired("query")
	userCmd.AddCommand(userMeCmd, userSearchCmd)

	rootCmd.AddCommand(issueCmd, projectCmd, searchCmd, userCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
