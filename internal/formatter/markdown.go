package formatter

import (
	"fmt"
	"io"
	"strings"

	"github.com/bentsolheim/jira-cli/internal/jira"
)

// MarkdownFormatter outputs issues as Markdown, suitable for LLM/agent context.
type MarkdownFormatter struct {
	BaseURL string
}

func (f *MarkdownFormatter) FormatIssue(w io.Writer, issue *jira.Issue) error {
	ai := toAgentIssue(issue)
	var b strings.Builder

	b.WriteString(fmt.Sprintf("# [%s](%s/browse/%s): %s\n\n", ai.Key, f.BaseURL, ai.Key, ai.Summary))

	b.WriteString(fmt.Sprintf("- **Status:** %s\n", ai.Status))
	b.WriteString(fmt.Sprintf("- **Type:** %s\n", ai.Type))
	b.WriteString(fmt.Sprintf("- **Priority:** %s\n", ai.Priority))
	b.WriteString(fmt.Sprintf("- **Project:** %s\n", ai.Project))
	if ai.Assignee != "" {
		b.WriteString(fmt.Sprintf("- **Assignee:** %s\n", ai.Assignee))
	}
	if ai.Reporter != "" {
		b.WriteString(fmt.Sprintf("- **Reporter:** %s\n", ai.Reporter))
	}
	if ai.Resolution != "" {
		b.WriteString(fmt.Sprintf("- **Resolution:** %s\n", ai.Resolution))
	}
	if ai.Parent != "" {
		b.WriteString(fmt.Sprintf("- **Parent:** %s\n", ai.Parent))
	}
	if ai.Epic != "" {
		b.WriteString(fmt.Sprintf("- **Epic:** [%s](%s/browse/%s)\n", ai.Epic, f.BaseURL, ai.Epic))
	}
	if len(ai.Labels) > 0 {
		b.WriteString(fmt.Sprintf("- **Labels:** %s\n", strings.Join(ai.Labels, ", ")))
	}
	if len(ai.Components) > 0 {
		b.WriteString(fmt.Sprintf("- **Components:** %s\n", strings.Join(ai.Components, ", ")))
	}
	b.WriteString(fmt.Sprintf("- **Created:** %s\n", ai.Created))
	b.WriteString(fmt.Sprintf("- **Updated:** %s\n", ai.Updated))

	if ai.Description != "" {
		b.WriteString("\n## Description\n\n")
		b.WriteString(ai.Description)
		b.WriteString("\n")
	}

	if len(ai.Children) > 0 {
		b.WriteString("\n## Children\n\n")

		headers := []string{"Key", "Type", "Status", "Assignee", "Summary"}
		var rows [][]string
		for _, c := range ai.Children {
			rows = append(rows, []string{
				fmt.Sprintf("[%s](%s/browse/%s)", c.Key, f.BaseURL, c.Key),
				c.Type, c.Status, c.Assignee, c.Summary,
			})
		}
		writeAlignedTable(&b, headers, rows)

		b.WriteString("\n## Child Details\n")
		for i, c := range ai.Children {
			b.WriteString(fmt.Sprintf("\n### [%s](%s/browse/%s): %s\n\n", c.Key, f.BaseURL, c.Key, c.Summary))
			b.WriteString(fmt.Sprintf("- **Status:** %s\n", c.Status))
			b.WriteString(fmt.Sprintf("- **Type:** %s\n", c.Type))
			if c.Assignee != "" {
				b.WriteString(fmt.Sprintf("- **Assignee:** %s\n", c.Assignee))
			}
			if c.Description != "" {
				b.WriteString(fmt.Sprintf("\n%s\n", c.Description))
			}
			if i < len(ai.Children)-1 {
				b.WriteString("\n\n\n")
			}
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

// writeAlignedTable writes a markdown table with columns padded to equal width.
func writeAlignedTable(b *strings.Builder, headers []string, rows [][]string) {
	// Calculate max width for each column
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	// Write header row
	b.WriteString("|")
	for i, h := range headers {
		b.WriteString(fmt.Sprintf(" %-*s |", widths[i], h))
	}
	b.WriteString("\n")

	// Write separator row
	b.WriteString("|")
	for _, w := range widths {
		b.WriteString("-")
		b.WriteString(strings.Repeat("-", w))
		b.WriteString("-|")
	}
	b.WriteString("\n")

	// Write data rows
	for _, row := range rows {
		b.WriteString("|")
		for i := range headers {
			cell := ""
			if i < len(row) {
				cell = row[i]
			}
			b.WriteString(fmt.Sprintf(" %-*s |", widths[i], cell))
		}
		b.WriteString("\n")
	}
}

func (f *MarkdownFormatter) FormatSearchResult(w io.Writer, result *jira.SearchResult) error {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("# Search Results (%d of %d)\n\n", len(result.Issues), result.Total))

	headers := []string{"Key", "Type", "Status", "Priority", "Assignee", "Summary"}
	var rows [][]string
	for _, issue := range result.Issues {
		ai := toAgentIssue(&issue)
		rows = append(rows, []string{ai.Key, ai.Type, ai.Status, ai.Priority, ai.Assignee, ai.Summary})
	}
	writeAlignedTable(&b, headers, rows)

	_, err := io.WriteString(w, b.String())
	return err
}
