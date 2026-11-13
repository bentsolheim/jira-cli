package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/bentsolheim/jira-cli/internal/formatter"
	"github.com/bentsolheim/jira-cli/internal/jira"
	"github.com/bentsolheim/jira-cli/internal/keychain"
	"github.com/spf13/cobra"
)

var (
	commentFile   string
	commentBody   string
	commentDryRun bool
)

type commentAdder interface {
	AddComment(key, body string) (*jira.Comment, error)
}

// newCommentClient is a variable so tests can substitute a fake client.
var newCommentClient = func() (commentAdder, error) {
	token, err := keychain.GetPAT(jiraURL)
	if err != nil {
		return nil, err
	}
	return jira.NewClient(jiraURL, token, verbose), nil
}

var commentCmd = &cobra.Command{
	Use:   "comment KEY",
	Short: "Add a comment to a Jira issue",
	Long: `Add a comment to a Jira issue. The comment body is sent as-is,
so write it in Jira wiki markup (h3., *bold*, {code}, [text|url], * bullets).

The body is read from exactly one of:
  --file PATH   Read body from a file ("-" means stdin)
  --body TEXT   Use the given text
  stdin         Used when neither --file nor --body is given

Trailing whitespace is trimmed. An empty body is an error.
Use --dry-run to validate and preview without contacting Jira.

Examples:
  jira comment MUP-123 --file comment.txt
  jira comment MUP-123 --file comment.txt --dry-run
  cat comment.txt | jira comment MUP-123 -o json
  jira comment MUP-123 --body "Deployed to test."`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := strings.TrimSpace(args[0])
		if key == "" {
			return fmt.Errorf("issue key is required")
		}

		body, err := readCommentBody(commentFile, commentBody, cmd.Flags().Changed("body"), cmd.InOrStdin())
		if err != nil {
			return err
		}

		f, err := formatter.New(outputFormat, jiraURL)
		if err != nil {
			return err
		}

		if commentDryRun {
			return f.FormatComment(cmd.OutOrStdout(), &formatter.CommentResult{
				IssueKey: key,
				Comment:  &jira.Comment{Body: body},
				DryRun:   true,
			})
		}

		client, err := newCommentClient()
		if err != nil {
			return err
		}

		comment, err := client.AddComment(key, body)
		if err != nil {
			return fmt.Errorf("adding comment to %s: %w", key, err)
		}

		return f.FormatComment(cmd.OutOrStdout(), &formatter.CommentResult{
			IssueKey: key,
			Comment:  comment,
		})
	},
}

// readCommentBody resolves the comment body from --file, --body or stdin.
func readCommentBody(file, body string, bodySet bool, stdin io.Reader) (string, error) {
	if file != "" && bodySet {
		return "", errors.New("use either --file or --body, not both")
	}

	var text string
	switch {
	case bodySet:
		text = body
	case file != "" && file != "-":
		data, err := os.ReadFile(file)
		if err != nil {
			return "", fmt.Errorf("reading comment file: %w", err)
		}
		text = string(data)
	default:
		if f, ok := stdin.(*os.File); ok && file == "" {
			if fi, err := f.Stat(); err == nil && fi.Mode()&os.ModeCharDevice != 0 {
				return "", errors.New("no comment body: use --file PATH, --body TEXT, or pipe text on stdin")
			}
		}
		data, err := io.ReadAll(stdin)
		if err != nil {
			return "", fmt.Errorf("reading stdin: %w", err)
		}
		text = string(data)
	}

	text = strings.TrimRight(text, " \t\r\n")
	if strings.TrimSpace(text) == "" {
		return "", errors.New("comment body is empty")
	}
	return text, nil
}

func init() {
	commentCmd.Flags().StringVarP(&commentFile, "file", "f", "", `Read comment body from file ("-" for stdin)`)
	commentCmd.Flags().StringVarP(&commentBody, "body", "b", "", "Comment body text")
	commentCmd.Flags().BoolVar(&commentDryRun, "dry-run", false, "Show what would be posted without contacting Jira")
	rootCmd.AddCommand(commentCmd)
}
