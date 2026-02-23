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

var issueCmd = &cobra.Command{
	Use:   "issue [KEY...]",
	Short: "Get details of one or more Jira issues",
	Long: `Fetch full details of one or more Jira issues by key.
Keys can be provided as separate arguments or comma-separated.

Examples:
  jira issue PROJ-123
  jira issue PROJ-123 PROJ-456
  jira issue PROJ-123,PROJ-456,PROJ-789
  jira issue PROJ-123 -o markdown`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Support both "KEY1 KEY2" and "KEY1,KEY2" styles
		var keys []string
		for _, arg := range args {
			for _, k := range strings.Split(arg, ",") {
				k = strings.TrimSpace(k)
				if k != "" {
					keys = append(keys, k)
				}
			}
		}

		token, err := keychain.GetPAT(jiraURL)
		if err != nil {
			return err
		}

		client := jira.NewClient(jiraURL, token, verbose)
		f, err := formatter.New(outputFormat, jiraURL)
		if err != nil {
			return err
		}

		for i, key := range keys {
			issue, err := client.GetIssue(key)
			if err != nil {
				return fmt.Errorf("failed to get issue %s: %w", key, err)
			}

			if i > 0 {
				fmt.Fprintln(os.Stdout, "\n---\n")
			}

			if err := f.FormatIssue(os.Stdout, issue); err != nil {
				return err
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(issueCmd)
}
