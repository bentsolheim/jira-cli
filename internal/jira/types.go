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

	// EpicChildren holds issues belonging to this epic.
	// Not populated from JSON — filled by a separate API call.
	EpicChildren []Issue `json:"-"`
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
	EpicLink    string      `json:"customfield_10761"`
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

// IssueInput is the user-friendly YAML input format.
type IssueInput struct {
	Project     string   `yaml:"project"`
	Summary     string   `yaml:"summary"`
	Description string   `yaml:"description"`
	Type        string   `yaml:"type"`
	Labels      []string `yaml:"labels"`
	EpicLink    string   `yaml:"epicLink"`
}

// IssueCreateRequest represents the payload for creating a Jira issue.
type IssueCreateRequest struct {
	Fields IssueCreateFields `json:"fields"`
}

// IssueCreateFields contains fields for creating an issue.
type IssueCreateFields struct {
	Project     *ProjectRef `json:"project"`
	Summary     string      `json:"summary"`
	Description string      `json:"description,omitempty"`
	IssueType   *TypeRef    `json:"issuetype"`
	Labels      []string    `json:"labels,omitempty"`
	EpicLink    string      `json:"customfield_10761,omitempty"`
}

// IssueUpdateRequest represents the payload for updating a Jira issue.
type IssueUpdateRequest struct {
	Fields IssueUpdateFields `json:"fields"`
}

// IssueUpdateFields contains fields for updating an issue.
type IssueUpdateFields struct {
	Summary     *string   `json:"summary,omitempty"`
	Description *string   `json:"description,omitempty"`
	IssueType   *TypeRef  `json:"issuetype,omitempty"`
	Labels      *[]string `json:"labels,omitempty"`
	EpicLink    *string   `json:"customfield_10761,omitempty"`
}

// ProjectRef is a reference to a project by key.
type ProjectRef struct {
	Key string `json:"key"`
}

// TypeRef is a reference to an issue type by name.
type TypeRef struct {
	Name string `json:"name"`
}
