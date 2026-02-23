package formatter

import (
	"encoding/json"
	"io"

	"github.com/bentsolheim/jira-cli/internal/jira"
)

// JSONFormatter outputs issues as structured JSON.
type JSONFormatter struct{}

// agentIssue is a flattened, agent-friendly representation of a Jira issue.
type agentIssue struct {
	Key         string           `json:"key"`
	Summary     string           `json:"summary"`
	Status      string           `json:"status"`
	Priority    string           `json:"priority,omitempty"`
	Type        string           `json:"type"`
	Assignee    string           `json:"assignee,omitempty"`
	Reporter    string           `json:"reporter,omitempty"`
	Project     string           `json:"project"`
	Labels      []string         `json:"labels,omitempty"`
	Components  []string         `json:"components,omitempty"`
	Created     string           `json:"created"`
	Updated     string           `json:"updated"`
	Resolution  string           `json:"resolution,omitempty"`
	Description string           `json:"description,omitempty"`
	Parent      string           `json:"parent,omitempty"`
	Children    []agentChildIssue `json:"children,omitempty"`
	Links       []agentLink      `json:"links,omitempty"`
	Comments    []agentComment   `json:"comments,omitempty"`
}

type agentChildIssue struct {
	Key     string `json:"key"`
	Summary string `json:"summary"`
	Status  string `json:"status"`
	Type    string `json:"type,omitempty"`
}

type agentLink struct {
	Type     string `json:"type"`
	IssueKey string `json:"issueKey"`
	Summary  string `json:"summary"`
}

type agentComment struct {
	Author  string `json:"author"`
	Created string `json:"created"`
	Body    string `json:"body"`
}

type agentSearchResult struct {
	Total  int          `json:"total"`
	Count  int          `json:"count"`
	Issues []agentIssue `json:"issues"`
}

func toAgentIssue(issue *jira.Issue) agentIssue {
	ai := agentIssue{
		Key:         issue.Key,
		Summary:     issue.Fields.Summary,
		Description: issue.Fields.Description,
		Created:     issue.Fields.Created,
		Updated:     issue.Fields.Updated,
	}

	if issue.Fields.Status != nil {
		ai.Status = issue.Fields.Status.Name
	}
	if issue.Fields.Priority != nil {
		ai.Priority = issue.Fields.Priority.Name
	}
	if issue.Fields.IssueType != nil {
		ai.Type = issue.Fields.IssueType.Name
	}
	if issue.Fields.Assignee != nil {
		ai.Assignee = issue.Fields.Assignee.DisplayName
	}
	if issue.Fields.Reporter != nil {
		ai.Reporter = issue.Fields.Reporter.DisplayName
	}
	if issue.Fields.Project != nil {
		ai.Project = issue.Fields.Project.Key
	}
	if issue.Fields.Resolution != nil {
		ai.Resolution = issue.Fields.Resolution.Name
	}
	if issue.Fields.Parent != nil {
		ai.Parent = issue.Fields.Parent.Key
	}

	ai.Labels = issue.Fields.Labels

	for _, c := range issue.Fields.Components {
		ai.Components = append(ai.Components, c.Name)
	}

	for _, sub := range issue.Fields.Subtasks {
		child := agentChildIssue{
			Key:     sub.Key,
			Summary: sub.Fields.Summary,
		}
		if sub.Fields.Status != nil {
			child.Status = sub.Fields.Status.Name
		}
		if sub.Fields.IssueType != nil {
			child.Type = sub.Fields.IssueType.Name
		}
		ai.Children = append(ai.Children, child)
	}

	for _, link := range issue.Fields.IssueLinks {
		// Treat "Epic" links (Issues in Epic) as children
		isEpicChild := link.Type.Name == "Epic" && link.InwardIssue != nil
		if isEpicChild {
			child := agentChildIssue{
				Key:     link.InwardIssue.Key,
				Summary: link.InwardIssue.Fields.Summary,
			}
			if link.InwardIssue.Fields.Status != nil {
				child.Status = link.InwardIssue.Fields.Status.Name
			}
			if link.InwardIssue.Fields.IssueType != nil {
				child.Type = link.InwardIssue.Fields.IssueType.Name
			}
			ai.Children = append(ai.Children, child)
			continue
		}

		al := agentLink{Type: link.Type.Name}
		if link.OutwardIssue != nil {
			al.IssueKey = link.OutwardIssue.Key
			al.Summary = link.OutwardIssue.Fields.Summary
		} else if link.InwardIssue != nil {
			al.IssueKey = link.InwardIssue.Key
			al.Summary = link.InwardIssue.Fields.Summary
		}
		ai.Links = append(ai.Links, al)
	}

	if issue.Fields.Comment != nil {
		for _, c := range issue.Fields.Comment.Comments {
			ac := agentComment{
				Body:    c.Body,
				Created: c.Created,
			}
			if c.Author != nil {
				ac.Author = c.Author.DisplayName
			}
			ai.Comments = append(ai.Comments, ac)
		}
	}

	return ai
}

func (f *JSONFormatter) FormatIssue(w io.Writer, issue *jira.Issue) error {
	ai := toAgentIssue(issue)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(ai)
}

func (f *JSONFormatter) FormatSearchResult(w io.Writer, result *jira.SearchResult) error {
	ar := agentSearchResult{
		Total: result.Total,
		Count: len(result.Issues),
	}
	for _, issue := range result.Issues {
		ar.Issues = append(ar.Issues, toAgentIssue(&issue))
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(ar)
}
