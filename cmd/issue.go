package cmd

import (
	"fmt"
	"os"

	"github.com/bentsolheim/jira-cli/internal/formatter"
	"github.com/bentsolheim/jira-cli/internal/jira"
	"github.com/bentsolheim/jira-cli/internal/keychain"
	"github.com/spf13/cobra"
)

var issueCmd = &cobra.Command{
	Use:   "issue [KEY]",
	Short: "Get details of a Jira issue",
	Long: `Fetch full details of a Jira issue by its key.

Examples:
  jira issue PROJ-123
  jira issue PROJ-123 -o markdown
  jira issue PROJ-123 -o text`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]

		token, err := keychain.GetPAT(jiraURL)
		if err != nil {
			return err
		}

		client := jira.NewClient(jiraURL, token, verbose)
		issue, err := client.GetIssue(key)
		if err != nil {
			return fmt.Errorf("failed to get issue %s: %w", key, err)
		}

		f, err := formatter.New(outputFormat, jiraURL)
		if err != nil {
			return err
		}

		return f.FormatIssue(os.Stdout, issue)
	},
}

func init() {
	rootCmd.AddCommand(issueCmd)
}
