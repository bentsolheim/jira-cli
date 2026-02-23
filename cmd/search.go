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

var maxResults int

var searchCmd = &cobra.Command{
	Use:   "search [JQL]",
	Short: "Search for Jira issues using JQL",
	Long: `Search for issues using Jira Query Language (JQL).

Examples:
  jira search "project = MYPROJ AND status = Open"
  jira search "assignee = currentUser() ORDER BY updated DESC"
  jira search "labels = backend AND sprint in openSprints()" --max-results 20 -o markdown`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		jql := strings.Join(args, " ")

		token, err := keychain.GetPAT(jiraURL)
		if err != nil {
			return err
		}

		client := jira.NewClient(jiraURL, token)
		result, err := client.Search(jql, maxResults)
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
	searchCmd.Flags().IntVar(&maxResults, "max-results", 50, "Maximum number of results to return")
	rootCmd.AddCommand(searchCmd)
}
