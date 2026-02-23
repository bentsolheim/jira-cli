package formatter

import (
	"fmt"
	"io"
	"strings"

	"github.com/bentsolheim/jira-cli/internal/jira"
)

// MarkdownFormatter outputs issues as Markdown, suitable for LLM/agent context.
type MarkdownFormatter struct{}

func (f *MarkdownFormatter) FormatIssue(w io.Writer, issue *jira.Issue) error {
	ai := toAgentIssue(issue)
	var b strings.Builder

	b.WriteString(fmt.Sprintf("# %s: %s\n\n", ai.Key, ai.Summary))

	b.WriteString("| Field | Value |\n|---|---|\n")
	b.WriteString(fmt.Sprintf("| Status | %s |\n", ai.Status))
	b.WriteString(fmt.Sprintf("| Type | %s |\n", ai.Type))
	b.WriteString(fmt.Sprintf("| Priority | %s |\n", ai.Priority))
	b.WriteString(fmt.Sprintf("| Project | %s |\n", ai.Project))
	if ai.Assignee != "" {
		b.WriteString(fmt.Sprintf("| Assignee | %s |\n", ai.Assignee))
	}
	if ai.Reporter != "" {
		b.WriteString(fmt.Sprintf("| Reporter | %s |\n", ai.Reporter))
	}
	if ai.Resolution != "" {
		b.WriteString(fmt.Sprintf("| Resolution | %s |\n", ai.Resolution))
	}
	if ai.Parent != "" {
		b.WriteString(fmt.Sprintf("| Parent | %s |\n", ai.Parent))
	}
	if len(ai.Labels) > 0 {
		b.WriteString(fmt.Sprintf("| Labels | %s |\n", strings.Join(ai.Labels, ", ")))
	}
	if len(ai.Components) > 0 {
		b.WriteString(fmt.Sprintf("| Components | %s |\n", strings.Join(ai.Components, ", ")))
	}
	b.WriteString(fmt.Sprintf("| Created | %s |\n", ai.Created))
	b.WriteString(fmt.Sprintf("| Updated | %s |\n", ai.Updated))

	if ai.Description != "" {
		b.WriteString("\n## Description\n\n")
		b.WriteString(ai.Description)
		b.WriteString("\n")
	}

	if len(ai.Children) > 0 {
		b.WriteString("\n## Children\n\n")
		b.WriteString("| Key | Type | Status | Summary |\n")
		b.WriteString("|-----|------|--------|---------|\n")
		for _, c := range ai.Children {
			b.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n", c.Key, c.Type, c.Status, c.Summary))
		}
	}

	if len(ai.Links) > 0 {
		b.WriteString("\n## Links\n\n")
		for _, l := range ai.Links {
			b.WriteString(fmt.Sprintf("- **%s** %s: %s\n", l.Type, l.IssueKey, l.Summary))
		}
	}

	if len(ai.Comments) > 0 {
		b.WriteString("\n## Comments\n\n")
		for _, c := range ai.Comments {
			b.WriteString(fmt.Sprintf("### %s (%s)\n\n%s\n\n", c.Author, c.Created, c.Body))
		}
	}

	_, err := io.WriteString(w, b.String())
	return err
}

func (f *MarkdownFormatter) FormatSearchResult(w io.Writer, result *jira.SearchResult) error {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("# Search Results (%d of %d)\n\n", len(result.Issues), result.Total))
	b.WriteString("| Key | Type | Status | Priority | Assignee | Summary |\n")
	b.WriteString("|-----|------|--------|----------|----------|---------|\n")

	for _, issue := range result.Issues {
		ai := toAgentIssue(&issue)
		b.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %s | %s |\n",
			ai.Key, ai.Type, ai.Status, ai.Priority, ai.Assignee, ai.Summary))
	}

	_, err := io.WriteString(w, b.String())
	return err
}
