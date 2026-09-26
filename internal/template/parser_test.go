package template

import (
	"reflect"
	"testing"
)

func TestParseIssueTemplate(t *testing.T) {
	// Test cases
	testCases := []struct {
		name           string
		content        string
		expectedTitle  string
		expectedBody   string
		expectedLabels []string
		expectedError  bool
	}{
		{
			name: "Valid issue template",
			content: `---
title: "Test Issue"
labels: bug, enhancement
assignees: user1, user2
---
This is the body of the issue.
With multiple lines.`,
			expectedTitle:  "Test Issue",
			expectedBody:   "This is the body of the issue.\nWith multiple lines.",
			expectedLabels: []string{"bug", "enhancement"},
			expectedError:  false,
		},
		{
			name:          "No front matter",
			content:       "This is just content without front matter.",
			expectedError: true,
		},
		{
			name: "Invalid front matter format",
			content: `---
This is not valid YAML
---
Content`,
			expectedError: true,
		},
		{
			name: "Missing closing delimiter",
			content: `---
title: "Test"
Content without closing front matter delimiter`,
			expectedError: true,
		},
		{
			name: "Array labels",
			content: `---
title: "Test Issue"
labels:
  - bug
  - enhancement
---
Body content`,
			expectedTitle:  "Test Issue",
			expectedBody:   "Body content",
			expectedLabels: []string{"bug", "enhancement"},
			expectedError:  false,
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			parser := NewParser()
			issue, err := parser.ParseIssueTemplate(tc.content)

			// Check error expectation
			if tc.expectedError && err == nil {
				t.Error("Expected error, got nil")
				return
			}

			if !tc.expectedError && err != nil {
				t.Errorf("Expected no error, got: %v", err)
				return
			}

			// Skip further checks if we expected an error
			if tc.expectedError {
				return
			}

			// Check issue properties
			if issue.Title != tc.expectedTitle {
				t.Errorf("Expected title '%s', got '%s'", tc.expectedTitle, issue.Title)
			}

			if issue.Body != tc.expectedBody {
				t.Errorf("Expected body '%s', got '%s'", tc.expectedBody, issue.Body)
			}

			if !reflect.DeepEqual(issue.Labels, tc.expectedLabels) {
				t.Errorf("Expected labels %v, got %v", tc.expectedLabels, issue.Labels)
			}
		})
	}
}

func TestParseIssueTemplateListFields(t *testing.T) {
	// Test cases
	testCases := []struct {
		name              string
		content           string
		expectedLabels    []string
		expectedAssignees []string
	}{
		{
			name: "Empty value",
			content: `---
title: "Test Issue"
labels: ""
assignees: ""
---
Body content`,
			expectedLabels:    []string{},
			expectedAssignees: []string{},
		},
		{
			name: "Trailing empty value",
			content: `---
title: "Test Issue"
labels: "bug, "
assignees: "user1, "
---
Body content`,
			expectedLabels:    []string{"bug"},
			expectedAssignees: []string{"user1"},
		},
		{
			name: "Array with empty value",
			content: `---
title: "Test Issue"
labels:
  - "bug"
  - ""
assignees:
  - "user1"
  - ""
---
Body content`,
			expectedLabels:    []string{"bug"},
			expectedAssignees: []string{"user1"},
		},
		{
			name: "Empty array",
			content: `---
title: "Test Issue"
labels: []
assignees: []
---
Body content`,
			expectedLabels:    []string{},
			expectedAssignees: []string{},
		},
		{
			name: "Missing keys",
			content: `---
title: "Test Issue"
---
Body content`,
			expectedLabels:    []string{},
			expectedAssignees: []string{},
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			parser := NewParser()
			issue, err := parser.ParseIssueTemplate(tc.content)
			if err != nil {
				t.Fatalf("Expected no error, got: %v", err)
			}

			if !reflect.DeepEqual(issue.Labels, tc.expectedLabels) {
				t.Errorf("Expected labels %v, got %v", tc.expectedLabels, issue.Labels)
			}

			if !reflect.DeepEqual(issue.Assignees, tc.expectedAssignees) {
				t.Errorf("Expected assignees %v, got %v", tc.expectedAssignees, issue.Assignees)
			}
		})
	}
}
