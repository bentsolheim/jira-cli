# jira-cli

A command-line tool for querying your internal Jira installation, optimized for AI/KI agent consumption.

## Features

- **macOS Keychain integration** — PAT stored securely, no config files
- **JQL search** — Full Jira Query Language support
- **Issue details** — Fetch complete issue data including comments, links, children
- **Agent-friendly output** — Markdown (default), JSON, and plain text formats
- **Epic awareness** — "Issues in Epic" links are treated as children alongside subtasks
- **Flattened structure** — Output is deliberately simplified for LLM context windows

## Installation

```bash
go install github.com/bentsolheim/jira-cli@latest
```

Or build from source:

```bash
git clone https://github.com/bentsolheim/jira-cli.git
cd jira-cli
go build -o jira-cli .
```

## Setup

### 1. Store your PAT

Generate a Personal Access Token in Jira (Profile → Personal Access Tokens), then:

```bash
jira-cli auth store
```

### 2. Verify authentication

```bash
jira-cli auth test
```

## Usage

### Search for issues

```bash
# Default markdown output
jira-cli search "project = MYPROJ AND status = Open"

# JSON output (for programmatic use)
jira-cli search "assignee = currentUser() ORDER BY updated DESC" -o json

# Limit results
jira-cli search "labels = backend" --max-results 10
```

### Get issue details

```bash
# Markdown (default)
jira-cli issue PROJ-123

# JSON
jira-cli issue PROJ-123 -o json

# Plain text
jira-cli issue PROJ-123 -o text
```

### Use with a different Jira instance

```bash
jira-cli --url https://other-jira.example.com search "project = FOO"
```

## Output Formats

| Format | Flag | Best for |
|--------|------|----------|
| Markdown | `-o markdown` | Default. Human-readable, LLM context windows |
| JSON | `-o json` | AI agents, piping to `jq`, programmatic use |
| Text | `-o text` | Human terminal use |

## Example JSON Output

```json
{
  "key": "PROJ-123",
  "summary": "Implement user authentication",
  "status": "In Progress",
  "priority": "High",
  "type": "Story",
  "assignee": "Jane Doe",
  "project": "PROJ",
  "labels": ["backend", "security"],
  "created": "2026-01-15T10:30:00.000+0100",
  "updated": "2026-02-20T14:22:00.000+0100",
  "description": "As a user I want to...",
  "children": [
    {"key": "PROJ-124", "summary": "Add login endpoint", "status": "Done", "type": "Sub-task"},
    {"key": "PROJ-125", "summary": "Add OAuth support", "status": "Open", "type": "Story"}
  ],
  "comments": [
    {
      "author": "John Smith",
      "created": "2026-02-18T09:00:00.000+0100",
      "body": "Ready for review"
    }
  ]
}
```

## Keychain Management

```bash
jira-cli auth store    # Store/update PAT
jira-cli auth test     # Verify PAT works
jira-cli auth delete   # Remove PAT from Keychain
```

The PAT is stored in macOS Keychain under service name `jira-cli` with the Jira URL as the account identifier.
