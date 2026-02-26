# AI Agent Guide for jira-cli

This tool is **designed for AI agents first, humans second**. Interface design prioritizes agent ease of use above all else.

## Critical Rules for Agents

1. **NO DATA MODIFICATION WITHOUT PERMISSION** — Never use `create` or `update` commands during testing or autonomous operation without explicit user approval. This is strictly forbidden.

2. **Always use `--output json`** — Required for reliable parsing and programmatic use.

3. **YAML input for mutations** — `create` and `update` accept structured YAML via stdin.

4. **No interactive prompts** — All commands are fully automatable.

## Commands

### Read Operations (safe for testing)
- `ls` — List issues with filters
- `issue KEY` — Get issue details  
- `search "JQL"` — Advanced JQL queries

### Write Operations (require user approval)
- `create` — Create issue from YAML stdin
- `update --issue-key KEY` — Update issue from YAML stdin

## Key Points

- Auth via macOS Keychain (pre-configured by user)
- Output flattened for LLM context efficiency
- Epic issues include children in `epicChildren` field
- Labels are replaced on update, not merged
- Issue types vary by instance (use `ls` to discover)

## Basic Examples

```bash
# List issues (safe)
jira-cli ls --project MUP --mine --output json

# Get details (safe)
jira-cli issue MUP-123 --output json

# Create (REQUIRES USER APPROVAL)
echo 'project: MUP
summary: Task title
type: Task' | jira-cli create --output json

# Update (REQUIRES USER APPROVAL)
echo 'summary: New title' | jira-cli update --issue-key MUP-123 --output json
```
