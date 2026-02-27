package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/bentsolheim/jira-cli/internal/formatter"
	"github.com/bentsolheim/jira-cli/internal/jira"
	"github.com/bentsolheim/jira-cli/internal/keychain"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var updateIssueKey string

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update an existing Jira issue from YAML input",
	Long: `Update an existing Jira issue by providing YAML input via stdin.
Requires --issue-key flag to specify which issue to update.

Supported fields (all optional):
  summary:     Issue summary
  description: Issue description
  type:        Issue type name
  labels:      List of labels
  epicLink:    Epic issue key to link to
  epicName:    Epic short name
  parent:      Parent issue key (for subtasks)
  parentLink:  Parent Link for Epic → Del-leveranse hierarchy

Example YAML:
  summary: Updated summary
  labels:
    - priority
    - backend

Usage:
  echo 'summary: Updated task name
  labels:
    - urgent' | jira update --issue-key MUP-123`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if updateIssueKey == "" {
			return fmt.Errorf("--issue-key is required")
		}

		yamlData, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("reading stdin: %w", err)
		}

		var input jira.IssueInput
		if err := yaml.Unmarshal(yamlData, &input); err != nil {
			return fmt.Errorf("parsing YAML: %w", err)
		}

		req := &jira.IssueUpdateRequest{
			Fields: jira.IssueUpdateFields{},
		}

		if input.Summary != "" {
			req.Fields.Summary = &input.Summary
		}
		if input.Description != "" {
			req.Fields.Description = &input.Description
		}
		if input.Type != "" {
			req.Fields.IssueType = &jira.TypeRef{Name: input.Type}
		}
		if len(input.Labels) > 0 {
			req.Fields.Labels = &input.Labels
		}
		if input.EpicLink != "" {
			req.Fields.EpicLink = &input.EpicLink
		}
		if input.EpicName != "" {
			req.Fields.EpicName = &input.EpicName
		}
		if input.Parent != "" {
			req.Fields.Parent = &jira.IssueRef{Key: input.Parent}
		}
		if input.ParentLink != "" {
			req.Fields.ParentLink = &input.ParentLink
		}

		token, err := keychain.GetPAT(jiraURL)
		if err != nil {
			return err
		}

		client := jira.NewClient(jiraURL, token, verbose)
		issue, err := client.UpdateIssue(updateIssueKey, req)
		if err != nil {
			return fmt.Errorf("updating issue: %w", err)
		}

		f, err := formatter.New(outputFormat, jiraURL)
		if err != nil {
			return err
		}

		return f.FormatIssue(os.Stdout, issue)
	},
}

func init() {
	updateCmd.Flags().StringVar(&updateIssueKey, "issue-key", "", "Issue key to update (required)")
	updateCmd.MarkFlagRequired("issue-key")
	rootCmd.AddCommand(updateCmd)
}
