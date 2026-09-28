package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ntsk/gh-issue-bulk-create/internal/csv"
	"github.com/ntsk/gh-issue-bulk-create/internal/github"
	"github.com/ntsk/gh-issue-bulk-create/internal/template"
)

// CommandLineOptions holds the command line options
type CommandLineOptions struct {
	templateFile string
	csvFile      string
	dryRun       bool
	repo         string
	showHelp     bool
}

func printHelp() {
	helpText := `Usage: gh issue-bulk-create [options]

Create multiple GitHub issues in bulk using a template file and CSV data.

Options:
  --template FILE       Path to the template markdown file (required)
  --csv FILE            Path to the CSV file containing data (required)
  --repo OWNER/REPO     Target repository (default: current repository)
  --dry-run             Only show the content of issues without creating them
  -h, --help            Show this help message

Examples:
  gh issue-bulk-create --template sample-template.md --csv sample-data.csv
  gh issue-bulk-create --template sample-template.md --csv sample-data.csv --repo owner/repo
  gh issue-bulk-create --template sample-template.md --csv sample-data.csv --dry-run
`
	fmt.Println(helpText)
}

func parseFlags() CommandLineOptions {
	opts := CommandLineOptions{}

	fs := flag.NewFlagSet("gh-issue-bulk-create", flag.ExitOnError)

	fs.StringVar(&opts.templateFile, "template", "", "")
	fs.StringVar(&opts.csvFile, "csv", "", "")
	fs.BoolVar(&opts.dryRun, "dry-run", false, "")
	fs.StringVar(&opts.repo, "repo", "", "")
	fs.BoolVar(&opts.showHelp, "help", false, "")
	fs.BoolVar(&opts.showHelp, "h", false, "")

	fs.Usage = printHelp

	// Check for -h or --help in arguments
	for _, arg := range os.Args[1:] {
		if arg == "-h" || arg == "--help" {
			opts.showHelp = true
			break
		}
	}

	// Parse flags
	fs.Parse(os.Args[1:])

	return opts
}

func main() {
	// Parse command line arguments
	opts := parseFlags()

	// Show help and exit
	if opts.showHelp {
		printHelp()
		os.Exit(0)
	}

	// Check required arguments
	if opts.templateFile == "" || opts.csvFile == "" {
		fmt.Println("Error: Both template file and CSV file must be specified")
		printHelp()
		os.Exit(1)
	}

	// Initialize components
	csvParser := csv.NewParser()
	templateRenderer := template.NewRenderer()
	templateParser := template.NewParser()

	// Initialize GitHub client
	githubClient, err := github.NewClient()
	if err != nil {
		fmt.Printf("Failed to initialize GitHub API client: %v\n", err)
		os.Exit(1)
	}

	// Read template file
	tmplContent, err := os.ReadFile(opts.templateFile)
	if err != nil {
		fmt.Printf("Failed to read template file: %v\n", err)
		os.Exit(1)
	}

	// Extract variables from template
	templateVars := templateRenderer.ExtractVariables(string(tmplContent))

	// Read CSV file
	records, headers, err := csvParser.Parse(opts.csvFile)
	if err != nil {
		// Provide more user-friendly error messages for CSV validation errors
		if strings.Contains(err.Error(), "CSV file is empty") {
			fmt.Printf("Error: The CSV file '%s' is empty. Please add headers and data.\n", opts.csvFile)
		} else if strings.Contains(err.Error(), "empty header") {
			fmt.Printf("Error: CSV validation failed: %v\n", err)
			fmt.Println("All columns in the CSV file must have headers. Please check your CSV file.")
		} else {
			fmt.Printf("Failed to read CSV file: %v\n", err)
		}
		os.Exit(1)
	}

	// Validate headers against template variables
	warnings, err := csvParser.ValidateHeadersAgainstTemplate(headers, templateVars)
	if err != nil {
		fmt.Printf("Error: Failed to validate CSV headers: %v\n", err)
		os.Exit(1)
	}

	if len(warnings) > 0 {
		fmt.Println("Validation warnings:")
		for _, warning := range warnings {
			fmt.Println(" -", warning)
		}

		if strings.Contains(warnings[0], "missing from CSV headers") {
			fmt.Println("These missing variables will be left empty in the generated issues.")
			fmt.Println("Do you want to continue? (y/N)")
			var response string
			fmt.Scanln(&response)
			response = strings.ToLower(strings.TrimSpace(response))
			if response != "y" && response != "yes" {
				fmt.Println("Aborted.")
				os.Exit(0)
			}
		}
	}

	// Map records to data maps
	dataMaps := csvParser.MapRecords(records, headers)

	// Determine repository
	targetRepo := opts.repo
	if targetRepo == "" {
		// If not specified as a flag, try to get from current directory
		targetRepo, err = githubClient.GetCurrentRepository()
		if err != nil {
			fmt.Printf("Failed to determine repository: %v\n", err)
			fmt.Println("Please specify the repository using --repo option or run in a git repository")
			os.Exit(1)
		}
	}

	fmt.Printf("Target repository: %s\n", targetRepo)

	// Check rate limit before creating issues
	if !opts.dryRun {
		rateLimit, err := githubClient.GetRateLimit()
		if err != nil {
			fmt.Printf("Warning: Failed to check rate limit: %v\n", err)
		} else {
			fmt.Printf("Current rate limit: %d remaining out of %d\n",
				rateLimit.Rate.Remaining, rateLimit.Rate.Limit)

			resetTime := time.Unix(int64(rateLimit.Rate.Reset), 0)
			fmt.Printf("Reset time: %s (in %s)\n",
				resetTime.Format(time.RFC3339),
				time.Until(resetTime).Round(time.Minute))

			issueCount := len(dataMaps)
			if rateLimit.Rate.Remaining < issueCount {
				fmt.Printf("Warning: Not enough rate limit remaining (%d) for %d issues\n",
					rateLimit.Rate.Remaining, issueCount)
				fmt.Printf("You may hit the rate limit during execution.\n")
				fmt.Printf("Do you want to continue? (y/N): ")
				var response string
				fmt.Scanln(&response)
				response = strings.ToLower(strings.TrimSpace(response))
				if response != "y" && response != "yes" {
					fmt.Println("Aborted.")
					os.Exit(0)
				}
			} else {
				fmt.Printf("Rate limit looks sufficient for %d issues\n", issueCount)
			}
		}
	}

	// Process template and create issues
	for _, data := range dataMaps {
		// Render template with data
		processedContent, err := templateRenderer.Render(string(tmplContent), data)
		if err != nil {
			fmt.Printf("Failed to process template: %v\n", err)
			continue
		}

		// Parse issue template to get issue data
		issue, err := templateParser.ParseIssueTemplate(processedContent)
		if err != nil {
			fmt.Printf("Failed to parse issue template: %v\n", err)
			continue
		}

		if opts.dryRun {
			// Dry run: Show issue content
			fmt.Println("==== Issue Content ====")
			fmt.Printf("Title: %s\n", issue.Title)
			fmt.Printf("Labels: %v\n", issue.Labels)
			fmt.Printf("Assignees: %v\n", issue.Assignees)
			fmt.Printf("Milestone: %s\n", issue.Milestone)
			fmt.Printf("Body:\n%s\n", issue.Body)
			fmt.Println("=====================")
		} else {
			// Create issue
			response, err := githubClient.CreateIssue(issue, targetRepo)
			if err != nil {
				fmt.Printf("Failed to create issue: %v\n", err)
			} else {
				fmt.Printf("Issue #%d created: %s\n", response.Number, response.URL)
			}
		}
	}
}
