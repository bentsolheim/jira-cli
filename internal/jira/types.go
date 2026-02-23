package jira

// User represents a Jira user.
type User struct {
	Key          string `json:"key"`
	Name         string `json:"name"`
	DisplayName  string `json:"displayName"`
	EmailAddress string `json:"emailAddress"`
}

// Issue represents a Jira issue.
type Issue struct {
	Key    string      `json:"key"`
	Self   string      `json:"self"`
	Fields IssueFields `json:"fields"`
}

// IssueFields contains the fields of a Jira issue.
type IssueFields struct {
	Summary     string      `json:"summary"`
	Description string      `json:"description"`
	Status      *Status     `json:"status"`
	Priority    *Priority   `json:"priority"`
	Assignee    *User       `json:"assignee"`
	Reporter    *User       `json:"reporter"`
	IssueType   *IssueType  `json:"issuetype"`
	Project     *Project    `json:"project"`
	Created     string      `json:"created"`
	Updated     string      `json:"updated"`
	Labels      []string    `json:"labels"`
	Components  []Component `json:"components"`
	Comment     *Comments   `json:"comment"`
	IssueLinks  []IssueLink `json:"issuelinks"`
	Subtasks    []Issue     `json:"subtasks"`
	Parent      *Issue      `json:"parent"`
	Resolution  *Resolution `json:"resolution"`
}

// Status represents an issue status.
type Status struct {
	Name string `json:"name"`
}

// Priority represents an issue priority.
type Priority struct {
	Name string `json:"name"`
}

// IssueType represents the type of an issue.
type IssueType struct {
	Name    string `json:"name"`
	Subtask bool   `json:"subtask"`
}

// Project represents a Jira project.
type Project struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

// Component represents a project component.
type Component struct {
	Name string `json:"name"`
}

// Comments wraps a list of comments.
type Comments struct {
	Comments []Comment `json:"comments"`
	Total    int       `json:"total"`
}

// Comment represents an issue comment.
type Comment struct {
	Author  *User  `json:"author"`
	Body    string `json:"body"`
	Created string `json:"created"`
	Updated string `json:"updated"`
}

// IssueLink represents a link between issues.
type IssueLink struct {
	Type         IssueLinkType `json:"type"`
	InwardIssue  *Issue        `json:"inwardIssue"`
	OutwardIssue *Issue        `json:"outwardIssue"`
}

// IssueLinkType represents the type of link.
type IssueLinkType struct {
	Name    string `json:"name"`
	Inward  string `json:"inward"`
	Outward string `json:"outward"`
}

// Resolution represents an issue resolution.
type Resolution struct {
	Name string `json:"name"`
}

// SearchResult represents the result of a JQL search.
type SearchResult struct {
	StartAt    int     `json:"startAt"`
	MaxResults int     `json:"maxResults"`
	Total      int     `json:"total"`
	Issues     []Issue `json:"issues"`
}
