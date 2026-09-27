// Package template provides functionality for parsing and rendering markdown templates.
// It includes both parsing of front matter in markdown files and rendering templates with data.
package template

import (
	"fmt"
	"strings"

	"github.com/ntsk/gh-issue-bulk-create/pkg/models"
	"gopkg.in/yaml.v3"
)

// Parser provides markdown parsing functionality
type Parser struct{}

// NewParser creates a new markdown parser
func NewParser() *Parser {
	return &Parser{}
}

// ParseIssueTemplate parses a markdown template with front matter
// and returns an Issue model
func (p *Parser) ParseIssueTemplate(content string) (*models.Issue, error) {
	// Check if the content contains front matter
	if !strings.HasPrefix(content, "---") {
		return nil, fmt.Errorf("content does not start with front matter delimiter '---'")
	}

	// Split content into front matter and body
	parts := strings.SplitN(content, "---", 3)
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid front matter format")
	}

	frontMatter := parts[1]
	body := strings.TrimSpace(parts[2])

	// Parse front matter as YAML
	metadata := make(map[string]interface{})
	err := yaml.Unmarshal([]byte(frontMatter), &metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to parse front matter: %v", err)
	}

	// Extract metadata and create the Issue
	var issue models.Issue
	issue.Body = body

	// Extract title
	if title, ok := metadata["title"].(string); ok {
		issue.Title = title
	}

	// Extract labels
	issue.Labels = parseStringList(metadata["labels"])

	// Extract assignees
	issue.Assignees = parseStringList(metadata["assignees"])

	// Extract milestone
	if milestone, ok := metadata["milestone"].(string); ok {
		issue.Milestone = milestone
	}

	return &issue, nil
}

// parseStringList converts a front matter value into a list of strings,
// skipping entries that are empty
func parseStringList(value interface{}) []string {
	result := []string{}

	switch v := value.(type) {
	case string:
		for _, item := range strings.Split(v, ",") {
			if item = strings.TrimSpace(item); item != "" {
				result = append(result, item)
			}
		}
	case []interface{}:
		for _, item := range v {
			if itemStr, ok := item.(string); ok && itemStr != "" {
				result = append(result, itemStr)
			}
		}
	}

	return result
}
