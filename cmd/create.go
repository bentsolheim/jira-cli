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

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new Jira issue from YAML input",
	Long: `Create a new Jira issue by providing YAML input via stdin.

Supported fields:
  project:     Project key (e.g., "MUP")
  summary:     Issue summary (required)
  description: Issue description
  type:        Issue type name (e.g., "Task", "Bug", "Story", "Epos")
  labels:      List of labels
  epicLink:    Epic issue key to link to (for stories/tasks)
  epicName:    Epic short name (required when type is Epos/Epic)
  parent:      Parent issue key (for subtasks only)
  parentLink:  Parent Link for Epic → Del-leveranse hierarchy

Example YAML:
  project: MUP
  summary: Fix authentication bug
  description: Users cannot log in with SSO
  type: Bug
  labels:
    - security
    - urgent
  epicLink: MUP-123

Usage:
  echo 'project: MUP
  summary: New task
  type: Task' | jira create`,
	RunE: func(cmd *cobra.Command, args []string) error {
		yamlData, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("reading stdin: %w", err)
		}

		var input jira.IssueInput
		if err := yaml.Unmarshal(yamlData, &input); err != nil {
			return fmt.Errorf("parsing YAML: %w", err)
		}

		if input.Summary == "" {
			return fmt.Errorf("summary is required")
		}
		if input.Project == "" {
			return fmt.Errorf("project is required")
		}
		if input.Type == "" {
			return fmt.Errorf("type is required")
		}

		req := &jira.IssueCreateRequest{
			Fields: jira.IssueCreateFields{
				Project:    &jira.ProjectRef{Key: input.Project},
				Summary:    input.Summary,
				Description: input.Description,
				IssueType:  &jira.TypeRef{Name: input.Type},
				Labels:     input.Labels,
				EpicLink:   input.EpicLink,
				EpicName:   input.EpicName,
				ParentLink: input.ParentLink,
			},
		}

		if input.Parent != "" {
			req.Fields.Parent = &jira.IssueRef{Key: input.Parent}
		}

		token, err := keychain.GetPAT(jiraURL)
		if err != nil {
			return err
		}

		client := jira.NewClient(jiraURL, token, verbose)
		issue, err := client.CreateIssue(req)
		if err != nil {
			return fmt.Errorf("creating issue: %w", err)
		}

		f, err := formatter.New(outputFormat, jiraURL)
		if err != nil {
			return err
		}

		return f.FormatIssue(os.Stdout, issue)
	},
}

func init() {
	rootCmd.AddCommand(createCmd)
}
