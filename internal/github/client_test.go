package github

import (
	"encoding/json"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/ntsk/gh-issue-bulk-create/pkg/models"
)

// MockClient provides a mock GitHub client for testing
type MockClient struct {
	CreateIssueFunc       func(issue *models.Issue, repo string) (*models.IssueResponse, error)
	GetCurrentRepoFunc    func() (string, error)
	GetRateLimitFunc      func() (*models.RateLimitResponse, error)
	CreatedIssues         []*models.Issue
	GetCurrentRepoCounter int
}

// CreateIssue implements the ClientInterface for testing
func (m *MockClient) CreateIssue(issue *models.Issue, repo string) (*models.IssueResponse, error) {
	m.CreatedIssues = append(m.CreatedIssues, issue)
	if m.CreateIssueFunc != nil {
		return m.CreateIssueFunc(issue, repo)
	}
	return &models.IssueResponse{Number: 1, URL: "https://github.com/mock/repo/issues/1"}, nil
}

// GetCurrentRepository implements the ClientInterface for testing
func (m *MockClient) GetCurrentRepository() (string, error) {
	m.GetCurrentRepoCounter++
	if m.GetCurrentRepoFunc != nil {
		return m.GetCurrentRepoFunc()
	}
	return "mock/repo", nil
}

// GetRateLimit implements the ClientInterface for testing
func (m *MockClient) GetRateLimit() (*models.RateLimitResponse, error) {
	if m.GetRateLimitFunc != nil {
		return m.GetRateLimitFunc()
	}
	return &models.RateLimitResponse{
		Rate: models.RateLimit{
			Limit:     5000,
			Remaining: 4999,
			Reset:     1234567890,
		},
	}, nil
}

func TestMockClient(t *testing.T) {
	// Create mock client
	mockClient := &MockClient{}

	// Test issue
	issue := &models.Issue{
		Title: "Test Issue",
		Body:  "This is a test issue",
	}

	// Test CreateIssue
	response, err := mockClient.CreateIssue(issue, "test/repo")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Check response
	if response.Number != 1 {
		t.Errorf("Expected issue number 1, got %d", response.Number)
	}

	// Check that issue was stored
	if len(mockClient.CreatedIssues) != 1 {
		t.Errorf("Expected 1 created issue, got %d", len(mockClient.CreatedIssues))
	}

	// Test GetCurrentRepository
	repo, err := mockClient.GetCurrentRepository()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Check repo
	if repo != "mock/repo" {
		t.Errorf("Expected repo 'mock/repo', got '%s'", repo)
	}

	// Check counter
	if mockClient.GetCurrentRepoCounter != 1 {
		t.Errorf("Expected GetCurrentRepository to be called once, called %d times", mockClient.GetCurrentRepoCounter)
	}

	// Test with custom functions
	mockClient = &MockClient{
		CreateIssueFunc: func(issue *models.Issue, repo string) (*models.IssueResponse, error) {
			return &models.IssueResponse{Number: 42, URL: "custom-url"}, nil
		},
		GetCurrentRepoFunc: func() (string, error) {
			return "custom/repo", nil
		},
	}

	// Test custom CreateIssue
	response, _ = mockClient.CreateIssue(issue, "test/repo")
	if response.Number != 42 {
		t.Errorf("Expected custom issue number 42, got %d", response.Number)
	}

	// Test custom GetCurrentRepository
	repo, _ = mockClient.GetCurrentRepository()
	if repo != "custom/repo" {
		t.Errorf("Expected custom repo 'custom/repo', got '%s'", repo)
	}
}

// TestClientInterface demonstrates that MockClient implements ClientInterface
func TestClientInterface(t *testing.T) {
	// Verify that MockClient implements the interface by using it
	var client ClientInterface = &MockClient{}

	// Test that we can call interface methods
	issue := &models.Issue{
		Title: "Interface Test",
		Body:  "Testing interface implementation",
	}

	response, err := client.CreateIssue(issue, "test/repo")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if response.Number != 1 {
		t.Errorf("Expected issue number 1, got %d", response.Number)
	}

	repo, err := client.GetCurrentRepository()
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if repo != "mock/repo" {
		t.Errorf("Expected repo 'mock/repo', got '%s'", repo)
	}

	// Test GetRateLimit interface method
	rateLimit, err := client.GetRateLimit()
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if rateLimit.Rate.Limit != 5000 {
		t.Errorf("Expected rate limit 5000, got %d", rateLimit.Rate.Limit)
	}
}

// TestRateLimit tests the GetRateLimit functionality
func TestRateLimit(t *testing.T) {
	// Create mock client
	mockClient := &MockClient{}

	// Test GetRateLimit with default values
	rateLimit, err := mockClient.GetRateLimit()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Check default values
	if rateLimit.Rate.Limit != 5000 {
		t.Errorf("Expected rate limit 5000, got %d", rateLimit.Rate.Limit)
	}
	if rateLimit.Rate.Remaining != 4999 {
		t.Errorf("Expected remaining 4999, got %d", rateLimit.Rate.Remaining)
	}
	if rateLimit.Rate.Reset != 1234567890 {
		t.Errorf("Expected reset 1234567890, got %d", rateLimit.Rate.Reset)
	}

	// Test with custom function
	mockClient = &MockClient{
		GetRateLimitFunc: func() (*models.RateLimitResponse, error) {
			return &models.RateLimitResponse{
				Rate: models.RateLimit{
					Limit:     60,
					Remaining: 30,
					Reset:     1234567900,
				},
			}, nil
		},
	}

	rateLimit, err = mockClient.GetRateLimit()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if rateLimit.Rate.Limit != 60 {
		t.Errorf("Expected custom limit 60, got %d", rateLimit.Rate.Limit)
	}
	if rateLimit.Rate.Remaining != 30 {
		t.Errorf("Expected custom remaining 30, got %d", rateLimit.Rate.Remaining)
	}
}

// captureTransport records the body of the request it handles and returns a
// canned issue creation response
type captureTransport struct {
	body []byte
}

func (t *captureTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.Body != nil {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		t.body = body
	}

	return &http.Response{
		StatusCode: http.StatusCreated,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(`{"number": 1, "html_url": "https://github.com/test/repo/issues/1"}`)),
		Request:    req,
	}, nil
}

// TestCreateIssueRequestBody checks the list fields that the GitHub API
// rejects as null
func TestCreateIssueRequestBody(t *testing.T) {
	// Test cases
	testCases := []struct {
		name              string
		issue             *models.Issue
		expectedLabels    []string
		expectedAssignees []string
	}{
		{
			name: "Nil list fields",
			issue: &models.Issue{
				Title: "Test Issue",
				Body:  "This is a test issue",
			},
			expectedLabels:    []string{},
			expectedAssignees: []string{},
		},
		{
			name: "Empty list fields",
			issue: &models.Issue{
				Title:     "Test Issue",
				Body:      "This is a test issue",
				Labels:    []string{},
				Assignees: []string{},
			},
			expectedLabels:    []string{},
			expectedAssignees: []string{},
		},
		{
			name: "Populated list fields",
			issue: &models.Issue{
				Title:     "Test Issue",
				Body:      "This is a test issue",
				Labels:    []string{"bug"},
				Assignees: []string{"user1"},
			},
			expectedLabels:    []string{"bug"},
			expectedAssignees: []string{"user1"},
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			transport := &captureTransport{}
			restClient, err := api.NewRESTClient(api.ClientOptions{
				AuthToken: "test-token",
				Transport: transport,
			})
			if err != nil {
				t.Fatalf("Expected no error, got: %v", err)
			}

			client := WithClient(restClient)
			if _, err := client.CreateIssue(tc.issue, "test/repo"); err != nil {
				t.Fatalf("Expected no error, got: %v", err)
			}

			var requestBody struct {
				Labels    []string `json:"labels"`
				Assignees []string `json:"assignees"`
			}
			if err := json.Unmarshal(transport.body, &requestBody); err != nil {
				t.Fatalf("Expected no error, got: %v", err)
			}

			if !reflect.DeepEqual(requestBody.Labels, tc.expectedLabels) {
				t.Errorf("Expected labels %#v, got %#v", tc.expectedLabels, requestBody.Labels)
			}

			if !reflect.DeepEqual(requestBody.Assignees, tc.expectedAssignees) {
				t.Errorf("Expected assignees %#v, got %#v", tc.expectedAssignees, requestBody.Assignees)
			}
		})
	}
}
