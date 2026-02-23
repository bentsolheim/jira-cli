package formatter

import (
	"fmt"
	"io"

	"github.com/bentsolheim/jira-cli/internal/jira"
)

// Formatter defines the interface for output formatting.
type Formatter interface {
	FormatIssue(w io.Writer, issue *jira.Issue) error
	FormatSearchResult(w io.Writer, result *jira.SearchResult) error
}

// New creates a formatter for the given format name.
func New(format string) (Formatter, error) {
	switch format {
	case "json":
		return &JSONFormatter{}, nil
	case "markdown", "md":
		return &MarkdownFormatter{}, nil
	case "text":
		return &TextFormatter{}, nil
	default:
		return nil, fmt.Errorf("unknown output format: %q (use json, markdown, or text)", format)
	}
}
