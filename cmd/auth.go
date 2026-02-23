package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/bentsolheim/jira-cli/internal/jira"
	"github.com/bentsolheim/jira-cli/internal/keychain"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage Jira authentication",
}

var authStoreCmd = &cobra.Command{
	Use:   "store",
	Short: "Store a Personal Access Token in the macOS Keychain",
	RunE: func(cmd *cobra.Command, args []string) error {
		var token string

		if term.IsTerminal(int(os.Stdin.Fd())) {
			fmt.Fprintf(os.Stderr, "Enter your Jira PAT for %s: ", jiraURL)
			raw, err := term.ReadPassword(int(os.Stdin.Fd()))
			if err != nil {
				return fmt.Errorf("reading token: %w", err)
			}
			fmt.Fprintln(os.Stderr) // newline after hidden input
			token = string(raw)
		} else {
			// Support piped input: echo "token" | jira auth store
			reader := bufio.NewReader(os.Stdin)
			raw, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("reading token from stdin: %w", err)
			}
			token = raw
		}

		token = strings.TrimSpace(token)
		if token == "" {
			return fmt.Errorf("token cannot be empty")
		}

		if err := keychain.StorePAT(jiraURL, token); err != nil {
			return err
		}
		fmt.Fprintln(os.Stderr, "PAT stored successfully in Keychain.")
		return nil
	},
}

var authTestCmd = &cobra.Command{
	Use:   "test",
	Short: "Verify that the stored PAT works against the Jira API",
	RunE: func(cmd *cobra.Command, args []string) error {
		token, err := keychain.GetPAT(jiraURL)
		if err != nil {
			return err
		}

		client := jira.NewClient(jiraURL, token, verbose)
		user, err := client.Myself()
		if err != nil {
			return fmt.Errorf("authentication failed: %w", err)
		}

		fmt.Fprintf(os.Stderr, "Authenticated as: %s (%s)\n", user.DisplayName, user.EmailAddress)
		return nil
	},
}

var authDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Remove the stored PAT from the macOS Keychain",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := keychain.DeletePAT(jiraURL); err != nil {
			return err
		}
		fmt.Fprintln(os.Stderr, "PAT deleted from Keychain.")
		return nil
	},
}

func init() {
	authCmd.AddCommand(authStoreCmd)
	authCmd.AddCommand(authTestCmd)
	authCmd.AddCommand(authDeleteCmd)
	rootCmd.AddCommand(authCmd)
}
