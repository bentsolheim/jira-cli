package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	outputFormat string
	jiraURL      string
)

var rootCmd = &cobra.Command{
	Use:   "jira",
	Short: "CLI for querying Jira issues, optimized for AI agent consumption",
	Long: `A command-line tool that queries your internal Jira installation
and presents issues in structured formats (JSON, Markdown, text)
suitable for AI/KI agent consumption.

Authentication uses a Personal Access Token stored in the macOS Keychain.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "markdown", "Output format: markdown, json, text")
	rootCmd.PersistentFlags().StringVar(&jiraURL, "url", "https://jira.sits.no", "Jira base URL")
}
