package formatter

import (
	"fmt"
	"io"
	"strings"
	"time"

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

// formatShortDate converts a Jira timestamp to "jan. 02 15:04" format.
func formatShortDate(jiraDate string) string {
	// Jira uses ISO 8601: "2026-02-23T12:39:11.408+0100"
	formats := []string{
		"2006-01-02T15:04:05.000-0700",
		"2006-01-02T15:04:05.000+0000",
		"2006-01-02T15:04:05.000Z",
		time.RFC3339,
	}
	months := []string{
		"jan.", "feb.", "mar.", "apr.", "mai", "jun.",
		"jul.", "aug.", "sep.", "okt.", "nov.", "des.",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, jiraDate); err == nil {
			return fmt.Sprintf("%s %02d %02d:%02d", months[t.Month()-1], t.Day(), t.Hour(), t.Minute())
		}
	}
	// Fallback: truncate if long enough
	if len(jiraDate) >= 16 {
		return jiraDate[:10] + " " + jiraDate[11:16]
	}
	return jiraDate
}

// shortenName converts "Lastname, Firstname Middle" to "Firstname ML".
// E.g. "Tørå Hagli, Andreas" -> "Andreas HT", "Christensen, Kristoffer Waage" -> "Kristoffer WC".
func shortenName(name string) string {
	if name == "" {
		return ""
	}
	parts := strings.SplitN(name, ", ", 2)
	if len(parts) != 2 {
		// Not in "Last, First" format — return as-is
		return name
	}
	lastParts := strings.Fields(parts[0])
	firstParts := strings.Fields(parts[1])
	if len(firstParts) == 0 {
		return name
	}
	firstName := firstParts[0]
	// Build initials from remaining first-name parts + last-name parts (reversed)
	var initials []rune
	for _, p := range firstParts[1:] {
		for _, r := range p {
			initials = append(initials, []rune(strings.ToUpper(string(r)))[0])
			break
		}
	}
	for _, p := range lastParts {
		for _, r := range p {
			initials = append(initials, []rune(strings.ToUpper(string(r)))[0])
			break
		}
	}
	if len(initials) > 0 {
		return firstName + " " + string(initials)
	}
	return firstName
}

// parentOrEpic returns the parent key or epic key for display in tables.
func parentOrEpic(ai agentIssue) string {
	if ai.Parent != "" {
		return ai.Parent
	}
	return ai.Epic
}

func (f *MarkdownFormatter) FormatSearchResult(w io.Writer, result *jira.SearchResult) error {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("# Search Results (%d of %d)\n\n", len(result.Issues), result.Total))

	headers := []string{"Created", "Updated", "Key", "Parent", "Type", "Status", "Assignee", "Summary"}
	var rows [][]string
	for _, issue := range result.Issues {
		ai := toAgentIssue(&issue)
		rows = append(rows, []string{
			formatShortDate(ai.Created), formatShortDate(ai.Updated),
			ai.Key, parentOrEpic(ai), ai.Type, ai.Status, shortenName(ai.Assignee), ai.Summary,
		})
	}
	writeAlignedTable(&b, headers, rows)

	_, err := io.WriteString(w, b.String())
	return err
}
