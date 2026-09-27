// Package github provides a client for interacting with the GitHub API.
// It includes functionality for creating issues and retrieving repository information.
package github

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/cli/go-gh/v2"
	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/ntsk/gh-issue-bulk-create/pkg/models"
)

// ClientInterface defines the interface for GitHub API operations
type ClientInterface interface {
	CreateIssue(issue *models.Issue, repo string) (*models.IssueResponse, error)
	GetCurrentRepository() (string, error)
	GetRateLimit() (*models.RateLimitResponse, error)
}

// Client provides GitHub API functionality
type Client struct {
	client *api.RESTClient
}

// NewClient creates a new GitHub API client
func NewClient() (*Client, error) {
	client, err := api.DefaultRESTClient()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize GitHub API client: %v", err)
	}
	return &Client{client: client}, nil
}

// WithClient creates a new GitHub client with a given REST client (for testing)
func WithClient(client *api.RESTClient) *Client {
	return &Client{client: client}
}

// CreateIssue creates a new GitHub issue
func (c *Client) CreateIssue(issue *models.Issue, repo string) (*models.IssueResponse, error) {
	// GitHub API response structure
	response := &models.IssueResponse{}

	// The API rejects null for these fields, and a nil slice marshals as null
	labels := issue.Labels
	if labels == nil {
		labels = []string{}
	}

	assignees := issue.Assignees
	if assignees == nil {
		assignees = []string{}
	}

	// Build request body
	requestBody := map[string]interface{}{
		"title":     issue.Title,
		"body":      issue.Body,
		"labels":    labels,
		"assignees": assignees,
	}

	if issue.Milestone != "" {
		requestBody["milestone"] = issue.Milestone
	}

	// Convert request body to JSON
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %v", err)
	}

	// Send POST request
	path := fmt.Sprintf("repos/%s/issues", repo)
	err = c.client.Post(path, bytes.NewReader(jsonData), response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// GetCurrentRepository gets the repository information for the current directory
func (c *Client) GetCurrentRepository() (string, error) {
	// RepoInfo structure to parse JSON output
	type RepoInfo struct {
		Owner struct {
			Login string `json:"login"`
		} `json:"owner"`
		Name string `json:"name"`
	}

	// Use gh command to get repository information
	output, stderr, err := gh.Exec("repo", "view", "--json", "owner,name")
	if err != nil {
		return "", fmt.Errorf("failed to get repository information: %s - %s", err, stderr.String())
	}

	// Parse output as JSON
	var info RepoInfo
	if err := json.Unmarshal(output.Bytes(), &info); err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/%s", info.Owner.Login, info.Name), nil
}

// GetRateLimit gets the current GitHub API rate limit information
func (c *Client) GetRateLimit() (*models.RateLimitResponse, error) {
	response := &models.RateLimitResponse{}

	// Send GET request to rate_limit endpoint
	err := c.client.Get("rate_limit", response)
	if err != nil {
		return nil, fmt.Errorf("failed to get rate limit: %v", err)
	}

	return response, nil
}
