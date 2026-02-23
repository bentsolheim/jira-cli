package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/bentsolheim/jira-cli/internal/formatter"
	"github.com/bentsolheim/jira-cli/internal/jira"
	"github.com/bentsolheim/jira-cli/internal/keychain"
	"github.com/spf13/cobra"
)

var (
	lsMine         bool
	lsStatus       string
	lsProject      string
	lsIncludeClosed bool
	lsMaxResults   int
)

const defaultClosedStatuses = "Lukket,Utført"

func getDefaultProject() string {
	return os.Getenv("JIRA_PROJECT")
}

func getClosedStatuses() []string {
	val := os.Getenv("JIRA_CLOSED_STATUSES")
	if val == "" {
		val = defaultClosedStatuses
	}
	parts := strings.Split(val, ",")
	var statuses []string
	for _, s := range parts {
		s = strings.TrimSpace(s)
		if s != "" {
			statuses = append(statuses, s)
		}
	}
	return statuses
}

func buildLsJQL(project, text, status string, mine, includeClosed bool) string {
	var conditions []string

	conditions = append(conditions, fmt.Sprintf("project = %s", project))

	if !includeClosed {
		statuses := getClosedStatuses()
		quoted := make([]string, len(statuses))
		for i, s := range statuses {
			quoted[i] = fmt.Sprintf("%q", s)
		}
		conditions = append(conditions, fmt.Sprintf("status NOT IN (%s)", strings.Join(quoted, ", ")))
	}

	if status != "" {
		conditions = append(conditions, fmt.Sprintf("status = %q", status))
	}

	if mine {
		conditions = append(conditions, "assignee = currentUser()")
	}

	if text != "" {
		conditions = append(conditions, fmt.Sprintf("(summary ~ %q OR description ~ %q)", text, text))
	}

	return strings.Join(conditions, " AND ") + " ORDER BY Rank ASC"
}

var lsCmd = &cobra.Command{
	Use:   "ls [text]",
	Short: "List open issues in the default project",
	Long: `List issues that are not closed in the default project.

The default project is read from the JIRA_PROJECT environment variable.
Closed statuses default to "Lukket, Utført" and can be overridden with
the JIRA_CLOSED_STATUSES environment variable (comma-separated).

An optional text argument searches in summary and description fields.

Examples:
  jira ls                              # All open issues
  jira ls "blueprint"                  # Text search in summary/description
  jira ls --mine                       # Only my issues
  jira ls "terraform" --mine           # My issues matching "terraform"
  jira ls --status "I gang"            # Only issues with status "I gang"
  jira ls --project OTHER              # Override default project
  jira ls --include-closed             # Include closed/resolved issues`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		project := lsProject
		if project == "" {
			project = getDefaultProject()
		}
		if project == "" {
			return fmt.Errorf("no project specified: set JIRA_PROJECT environment variable or use --project")
		}

		var text string
		if len(args) > 0 {
			text = args[0]
		}

		jql := buildLsJQL(project, text, lsStatus, lsMine, lsIncludeClosed)

		if verbose {
			fmt.Fprintf(os.Stderr, "JQL: %s\n", jql)
		}

		token, err := keychain.GetPAT(jiraURL)
		if err != nil {
			return err
		}

		client := jira.NewClient(jiraURL, token, verbose)
		result, err := client.Search(jql, lsMaxResults)
		if err != nil {
			return fmt.Errorf("search failed: %w", err)
		}

		f, err := formatter.New(outputFormat, jiraURL)
		if err != nil {
			return err
		}

		return f.FormatSearchResult(os.Stdout, result)
	},
}

func init() {
	lsCmd.Flags().BoolVar(&lsMine, "mine", false, "Only show issues assigned to me")
	lsCmd.Flags().StringVar(&lsStatus, "status", "", "Filter by specific status")
	lsCmd.Flags().StringVar(&lsProject, "project", "", "Override default project (env: JIRA_PROJECT)")
	lsCmd.Flags().BoolVar(&lsIncludeClosed, "include-closed", false, "Include closed/resolved issues")
	lsCmd.Flags().IntVar(&lsMaxResults, "max-results", 50, "Maximum number of results to return")
	rootCmd.AddCommand(lsCmd)
}
