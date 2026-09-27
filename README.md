# gh-issue-bulk-create

![](https://github.com/ntsk/gh-issue-bulk-create/actions/workflows/ci.yml/badge.svg)

A GitHub CLI extension to create multiple GitHub issues in bulk.

## Prerequisites

The extension requires the gh CLI to be installed and in the `PATH`. The extension also requires the user to have authenticated via `gh auth`.

## Installation

This project is a GitHub CLI extension. After installing the `gh` CLI, from a command-line run:

```bash
gh extension install ntsk/gh-issue-bulk-create
```

## Usage

This extension uses a template markdown file and a CSV file containing data to create multiple GitHub issues in bulk.

```bash
gh issue-bulk-create --template <template_file> --csv <csv_file> [--repo <owner/repo>] [--dry-run]
```

### Options

- `--template`: Path to the template markdown file (required)
- `--csv`: Path to the CSV file containing data (required)
- `--repo`: Target repository in the format of owner/repo (default: current repository)
- `--dry-run`: Only show the content of issues without creating them

### Template File

The template file follows the GitHub Issue template format with front matter metadata at the beginning of the file:

```markdown
---
title: "{{title}}"
labels: "{{label1}}, {{label2}}"
assignees: "{{assignee}}"
---

## Description
{{description}}

## Steps to Reproduce
{{steps}}
```

You can use mustache syntax (`{{variable_name}}`) in the template to embed data from the CSV file.

#### Labels and Assignees

`labels` and `assignees` are optional. Omit the keys to create issues without them:

```markdown
---
title: "{{title}}"
---
```

Both fields accept either a comma-separated string or a YAML list:

```markdown
labels: "bug, frontend"
assignees: "alice, bob"
```

```markdown
labels:
  - bug
  - frontend
assignees:
  - alice
  - bob
```

Empty entries are dropped. With `labels: "{{label1}}, {{label2}}"`, a row that fills in `label1` but leaves `label2` empty gets a single label rather than a blank one.

### CSV File

The CSV file **must contain a header row** with column names that match the variable names used in the template.
The first line is the header row, and each subsequent line will be used to create a separate issue.

```csv
title,label1,label2,assignee,description,steps
Login page error,bug,frontend,username,Error appears when clicking login button,Click login button
Search not working,bug,backend,username,"Results don't appear when searching,Enter ""test"" in search box and click search"
```

#### CSV Requirements

- The file must be in a standard comma-separated values (CSV) format following RFC 4180 specifications
- Headers are required and must be in the first row of the CSV file
- Each header (column name) must not be empty and should match variables used in the template
- Fields containing commas, newlines, or double quotes must be enclosed in double quotes
- Double quotes within a quoted field must be escaped by doubling them (e.g., `"` becomes `""`)
- Example of properly formatted CSV with special characters:
  ```csv
  title,description
  "Title with, comma","Description with ""quotes"""
  "Line breaks
  in text","Another field"
  ```

#### Warning Behaviors
- If there are CSV headers that aren't used in the template: A warning is displayed but the process continues
- If there are template variables that don't have corresponding CSV headers: A warning is displayed and you'll be prompted to confirm whether to continue. If you continue, those missing variables will be left empty in the generated issues

## Example

You can try it with the sample files included in the repository:

```bash
# From inside a git repository directory
gh issue-bulk-create --template sample-template.md --csv sample-data.csv --dry-run

# Or specify a different repository
gh issue-bulk-create --template sample-template.md --csv sample-data.csv --repo owner/repo-name
```