package formatter

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/bentsolheim/jira-cli/internal/jira"
)

// TextFormatter outputs issues as human-readable plain text tables.
type TextFormatter struct {
	BaseURL string
}

func (f *TextFormatter) FormatIssue(w io.Writer, issue *jira.Issue) error {
	ai := toAgentIssue(issue)
	var b strings.Builder

	b.WriteString(fmt.Sprintf("%s: %s\n", ai.Key, ai.Summary))
	b.WriteString(strings.Repeat("=", len(ai.Key)+len(ai.Summary)+2) + "\n\n")

	tw := tabwriter.NewWriter(&b, 0, 0, 2, ' ', 0)
	fmt.Fprintf(tw, "Status:\t%s\n", ai.Status)
	fmt.Fprintf(tw, "Type:\t%s\n", ai.Type)
	fmt.Fprintf(tw, "Priority:\t%s\n", ai.Priority)
	fmt.Fprintf(tw, "Project:\t%s\n", ai.Project)
	if ai.Assignee != "" {
		fmt.Fprintf(tw, "Assignee:\t%s\n", ai.Assignee)
	}
	if ai.Reporter != "" {
		fmt.Fprintf(tw, "Reporter:\t%s\n", ai.Reporter)
	}
	if ai.Resolution != "" {
		fmt.Fprintf(tw, "Resolution:\t%s\n", ai.Resolution)
	}
	if ai.Parent != "" {
		fmt.Fprintf(tw, "Parent:\t%s\n", ai.Parent)
	}
	if len(ai.Labels) > 0 {
		fmt.Fprintf(tw, "Labels:\t%s\n", strings.Join(ai.Labels, ", "))
	}
	if len(ai.Components) > 0 {
		fmt.Fprintf(tw, "Components:\t%s\n", strings.Join(ai.Components, ", "))
	}
	fmt.Fprintf(tw, "Created:\t%s\n", ai.Created)
	fmt.Fprintf(tw, "Updated:\t%s\n", ai.Updated)
	tw.Flush()

	if ai.Description != "" {
		b.WriteString("\nDescription:\n")
		b.WriteString(strings.Repeat("-", 40) + "\n")
		b.WriteString(ai.Description + "\n")
	}

	if len(ai.Children) > 0 {
		b.WriteString("\nChildren:\n")
		tw2 := tabwriter.NewWriter(&b, 0, 0, 2, ' ', 0)
		fmt.Fprintln(tw2, "  KEY\tTYPE\tSTATUS\tSUMMARY")
		fmt.Fprintln(tw2, "  ---\t----\t------\t-------")
		for _, c := range ai.Children {
			fmt.Fprintf(tw2, "  %s\t%s\t%s\t%s\n", c.Key, c.Type, c.Status, c.Summary)
		}
		tw2.Flush()
	}

	if len(ai.Links) > 0 {
		b.WriteString("\nLinks:\n")
		for _, l := range ai.Links {
			b.WriteString(fmt.Sprintf("  - [%s] %s: %s\n", l.Type, l.IssueKey, l.Summary))
		}
	}

	if len(ai.Comments) > 0 {
		b.WriteString("\nComments:\n")
		for _, c := range ai.Comments {
			b.WriteString(fmt.Sprintf("\n  %s (%s):\n  %s\n", c.Author, c.Created, c.Body))
		}
	}

	_, err := io.WriteString(w, b.String())
	return err
}

func (f *TextFormatter) FormatSearchResult(w io.Writer, result *jira.SearchResult) error {
	fmt.Fprintf(w, "Results: %d of %d\n\n", len(result.Issues), result.Total)

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "KEY\tTYPE\tSTATUS\tASSIGNEE\tPARENT\tCREATED\tUPDATED\tSUMMARY")
	fmt.Fprintln(tw, "---\t----\t------\t--------\t------\t-------\t-------\t-------")

	for _, issue := range result.Issues {
		ai := toAgentIssue(&issue)
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			ai.Key, ai.Type, ai.Status, ai.Assignee, parentOrEpic(ai),
			formatShortDate(ai.Created), formatShortDate(ai.Updated), ai.Summary)
	}

	return tw.Flush()
}
