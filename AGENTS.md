# AI Agent Guide for jira-cli

This tool is **designed for AI agents first, humans second**. All interface decisions prioritize programmatic consumption, structured I/O, and LLM context efficiency.

## Core Principles

1. **JSON output** — Always use `--output json` for reliable parsing
2. **YAML input** — Create/update operations accept structured YAML via stdin
3. **No interactive prompts** — All commands are fully automatable
4. **Predictable structure** — Consistent field names across all commands
5. **Context-optimized** — Output is deliberately flattened for token efficiency

## Authentication

The CLI uses macOS Keychain for credential storage. The PAT is already configured by the user — agents don't need to handle authentication.

If you encounter auth errors, inform the user to run:
```bash
jira-cli auth test
```

## Command Reference for Agents

### 1. List Issues (`ls`)

**Use case:** Find issues matching criteria, get project overview

```bash
jira-cli ls --project MUP --output json
jira-cli ls --mine --output json
jira-cli ls authentication --project MUP --output json
```

**Agent tips:**
- Default excludes closed issues (use `--include-closed` if needed)
- `--mine` filters to assigned issues for current user
- Text argument searches summary and description
- Returns array in `issues` field with `total` count

**Key flags:**
- `--project` — Filter by project key
- `--mine` — Show only my assigned issues
- `--status` — Filter by status name
- `--max-results` — Limit results (default: 50)
- `--include-closed` — Include closed/resolved issues
- `--output json` — Structured output

### 2. Get Issue Details (`issue`)

**Use case:** Fetch complete information about specific issue(s)

```bash
jira-cli issue MUP-123 --output json
jira-cli issue MUP-123 MUP-456 --output json
```

**Agent tips:**
- Accepts multiple issue keys (space-separated or comma-separated)
- Returns full issue object with comments, links, children
- Epic issues automatically include child issues in `epicChildren` field
- Use this when you need complete context about an issue

**Output structure:**
```json
{
  "key": "MUP-123",
  "summary": "Issue title",
  "status": "In Progress",
  "priority": "High",
  "type": "Task",
  "assignee": "User Name",
  "reporter": "User Name",
  "project": "MUP",
  "epic": "MUP-456",
  "labels": ["tag1", "tag2"],
  "created": "2026-01-15T10:30:00.000+0100",
  "updated": "2026-02-20T14:22:00.000+0100",
  "description": "Full description text",
  "comments": [
    {
      "author": "User Name",
      "created": "2026-02-18T09:00:00.000+0100",
      "body": "Comment text"
    }
  ],
  "links": [
    {
      "type": "Relates",
      "issueKey": "MUP-789",
      "summary": "Related issue"
    }
  ]
}
```

### 3. Create Issue (`create`)

**Use case:** Create new Jira issues programmatically

**Input format (YAML via stdin):**
```yaml
project: MUP
summary: Issue title (required)
description: Detailed description
type: Task
labels:
  - backend
  - urgent
epicLink: MUP-456
```

**Usage:**
```bash
echo 'project: MUP
summary: Implement user authentication
type: Task
description: Add JWT-based authentication
labels:
  - security
  - backend' | jira-cli create --output json
```

**Agent tips:**
- **Required fields:** `project`, `summary`, `type`
- **Optional fields:** `description`, `labels`, `epicLink`
- Type must match Jira's configured issue types (common: Task, Bug, Story, Epic, Forbedring, Epos)
- Returns the created issue object with assigned key
- Extract `key` field from response for subsequent operations

**Issue types vary by Jira instance** — use `ls` or `issue` to see what types are used:
```bash
jira-cli ls --project MUP --max-results 10 --output json | jq '.issues[].type' | sort -u
```

### 4. Update Issue (`update`)

**Use case:** Modify existing issues

**Input format (YAML via stdin, all fields optional):**
```yaml
summary: Updated title
description: Updated description
type: Bug
labels:
  - updated
  - priority
epicLink: MUP-789
```

**Usage:**
```bash
echo 'summary: Updated task name
labels:
  - completed
  - backend' | jira-cli update --issue-key MUP-123 --output json
```

**Agent tips:**
- **Required flag:** `--issue-key` (the issue to update)
- All YAML fields are optional — only provided fields are modified
- Returns the updated issue object after changes
- Labels are **replaced** not merged — include all desired labels

### 5. Search with JQL (`search`)

**Use case:** Advanced queries using Jira Query Language

```bash
jira-cli search 'project = MUP AND status = "In Progress" ORDER BY updated DESC' --output json
jira-cli search 'assignee = currentUser() AND updated >= -7d' --max-results 20 --output json
```

**Agent tips:**
- Use JQL for complex queries beyond `ls` capabilities
- Common JQL functions: `currentUser()`, `-7d` (last 7 days)
- Returns same structure as `ls` command
- See [JQL documentation](https://confluence.atlassian.com/jirasoftwareserver/advanced-searching-939938733.html) for syntax

## Agent Workflows

### Workflow 1: Create a task with epic link

```bash
# 1. Find the epic
EPIC_KEY=$(jira-cli ls "platform migration" --project MUP --output json | jq -r '.issues[] | select(.type == "Epos") | .key' | head -1)

# 2. Create subtask linked to epic
echo "project: MUP
summary: Migrate authentication service
type: Oppgave
description: Move auth service to new platform
epicLink: $EPIC_KEY
labels:
  - migration
  - backend" | jira-cli create --output json
```

### Workflow 2: Update multiple issues

```bash
# Find all open tasks with specific label
jira-cli ls --project MUP --output json | \
  jq -r '.issues[] | select(.labels[]? == "needs-review") | .key' | \
while read key; do
  echo "labels:
  - reviewed
  - completed" | jira-cli update --issue-key "$key" --output json
done
```

### Workflow 3: Report on epic progress

```bash
# Get epic with all children
jira-cli issue MUP-456 --output json | \
  jq '{
    epic: .key,
    summary: .summary,
    total_children: (.epicChildren | length),
    completed: [.epicChildren[] | select(.status == "Ferdig")] | length,
    in_progress: [.epicChildren[] | select(.status == "I gang")] | length
  }'
```

## Output Format Selection

| Format | Use Case |
|--------|----------|
| `--output json` | **Default for agents** — Structured, parseable, complete data |
| `--output markdown` | Human-readable reports, documentation generation |
| `--output text` | Terminal display, logs |

**Always use `--output json`** for reliable programmatic parsing.

## Error Handling

### Common errors and solutions:

**Authentication failed:**
```
Error: keychain error: The specified item could not be found in the keychain
```
→ User needs to run `jira-cli auth store`

**Invalid issue type:**
```
Error: creating issue: API error (HTTP 400): {"errors":{"issuetype":"The issue type selected is invalid."}}
```
→ Check valid types with `jira-cli ls --project X --output json | jq '.issues[].type' | sort -u`

**Issue not found:**
```
Error: updating issue: API error (HTTP 404): ...
```
→ Verify issue key exists with `jira-cli issue KEY --output json`

**Missing required field:**
```
Error: summary is required
Error: project is required
Error: type is required
```
→ Include all required fields in YAML input

## Best Practices for Agents

1. **Always use `--output json`** for programmatic consumption
2. **Parse JSON with `jq`** or language-native JSON parsers
3. **Extract issue keys** from responses for chaining commands
4. **Check issue types** before creating (types vary by Jira instance)
5. **Include all labels** when updating (labels are replaced, not merged)
6. **Use `ls` for discovery**, `issue` for details, `search` for complex queries
7. **Handle errors gracefully** — check exit codes and parse error messages
8. **Verify operations** by fetching the issue after create/update
9. **Respect rate limits** — avoid rapid-fire requests
10. **Test with dry-run mindset** — validate YAML structure before execution

## Environment Variables

- `JIRA_PROJECT` — Default project for `ls` command
- `JIRA_CLOSED_STATUSES` — Comma-separated list of statuses to exclude (default: "Ferdig,Avvist,Avbrutt")

## Field Mapping Reference

| YAML Field | Jira API Field | Notes |
|------------|---------------|-------|
| `project` | `fields.project.key` | Project key (e.g., "MUP") |
| `summary` | `fields.summary` | Issue title |
| `description` | `fields.description` | Full description text |
| `type` | `fields.issuetype.name` | Issue type name |
| `labels` | `fields.labels` | Array of strings |
| `epicLink` | `fields.customfield_10761` | Epic issue key |

## Limitations

- **macOS only** — Uses macOS Keychain for credential storage
- **Epic link field** — Custom field ID (10761) may vary by Jira instance
- **Issue types** — Available types are instance-specific
- **No assignee modification** — Not currently supported in create/update
- **No status transitions** — Cannot change workflow status (use Jira UI)
- **No file attachments** — Not supported
- **No comments via create/update** — Use Jira UI or API directly

## Examples for Common Tasks

### Get all open issues assigned to me
```bash
jira-cli ls --mine --output json
```

### Create a bug with description
```bash
echo 'project: MUP
summary: Login page returns 500 error
type: Bug
description: |
  When attempting to log in with valid credentials,
  the server returns HTTP 500.
  
  Steps to reproduce:
  1. Navigate to /login
  2. Enter credentials
  3. Click submit
labels:
  - critical
  - backend' | jira-cli create --output json
```

### Update issue description only
```bash
echo 'description: Updated description with new information' | \
  jira-cli update --issue-key MUP-123 --output json
```

### Find all issues updated in last 7 days
```bash
jira-cli search 'project = MUP AND updated >= -7d ORDER BY updated DESC' \
  --max-results 50 --output json
```

### Get full context for an issue
```bash
jira-cli issue MUP-123 --output json | jq '{
  key,
  summary,
  status,
  assignee,
  description,
  comment_count: (.comments | length),
  latest_comment: .comments[-1]
}'
```

## Integration with Agent Frameworks

### LangChain/LangGraph Tool Definition

```python
from langchain.tools import StructuredTool
import subprocess
import json

def jira_create_issue(project: str, summary: str, type: str, 
                      description: str = "", labels: list = None, 
                      epic_link: str = None) -> dict:
    """Create a Jira issue"""
    yaml_input = f"project: {project}\nsummary: {summary}\ntype: {type}\n"
    if description:
        yaml_input += f"description: {description}\n"
    if labels:
        yaml_input += "labels:\n" + "\n".join(f"  - {l}" for l in labels)
    if epic_link:
        yaml_input += f"epicLink: {epic_link}\n"
    
    result = subprocess.run(
        ["jira-cli", "create", "--output", "json"],
        input=yaml_input.encode(),
        capture_output=True
    )
    return json.loads(result.stdout)

jira_create_tool = StructuredTool.from_function(
    func=jira_create_issue,
    name="jira_create_issue",
    description="Create a new Jira issue with specified fields"
)
```

### OpenAI Function Calling

```json
{
  "name": "jira_list_issues",
  "description": "List Jira issues with optional filters",
  "parameters": {
    "type": "object",
    "properties": {
      "project": {
        "type": "string",
        "description": "Project key to filter by"
      },
      "mine": {
        "type": "boolean",
        "description": "Show only issues assigned to me"
      },
      "search_text": {
        "type": "string",
        "description": "Search text for summary/description"
      },
      "max_results": {
        "type": "integer",
        "description": "Maximum number of results to return"
      }
    }
  }
}
```

## Support

For issues or questions:
- Check error messages first — they include actionable guidance
- Verify authentication: `jira-cli auth test`
- Inspect available issue types: `jira-cli ls --project X --output json | jq '.issues[].type' | sort -u`
- Review this guide for agent-optimized patterns
