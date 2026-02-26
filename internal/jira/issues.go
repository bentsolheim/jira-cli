package jira

import (
	"fmt"
	"net/url"
	"strconv"
)

// GetIssue fetches a single issue by key (e.g. "PROJ-123").
// If the issue is an Epic, it automatically fetches the issues in the epic.
func (c *Client) GetIssue(key string) (*Issue, error) {
	var issue Issue
	path := fmt.Sprintf("/rest/api/2/issue/%s", url.PathEscape(key))
	if err := c.do("GET", path, &issue); err != nil {
		return nil, err
	}

	// If the issue is an Epic, fetch its children via JQL
	if issue.Fields.IssueType != nil && isEpicType(issue.Fields.IssueType.Name) {
		jql := fmt.Sprintf(`"Epic Link" = %s ORDER BY status ASC, key ASC`, key)
		result, err := c.Search(jql, 100)
		if err == nil && len(result.Issues) > 0 {
			issue.EpicChildren = result.Issues
		}
	}

	return &issue, nil
}

// isEpicType checks if the issue type name represents an Epic.
// Handles both English ("Epic") and Norwegian ("Epos") names.
func isEpicType(name string) bool {
	switch name {
	case "Epic", "Epos":
		return true
	}
	return false
}

// Search executes a JQL query and returns matching issues.
func (c *Client) Search(jql string, maxResults int) (*SearchResult, error) {
	var result SearchResult
	params := url.Values{}
	params.Set("jql", jql)
	params.Set("maxResults", strconv.Itoa(maxResults))
	path := "/rest/api/2/search?" + params.Encode()
	if err := c.do("GET", path, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// CreateIssue creates a new Jira issue and returns the created issue.
func (c *Client) CreateIssue(req *IssueCreateRequest) (*Issue, error) {
	var response struct {
		Key string `json:"key"`
	}
	
	if err := c.doWithBody("POST", "/rest/api/2/issue", req, &response); err != nil {
		return nil, err
	}
	
	return c.GetIssue(response.Key)
}

// UpdateIssue updates an existing Jira issue and returns the updated issue.
func (c *Client) UpdateIssue(key string, req *IssueUpdateRequest) (*Issue, error) {
	path := fmt.Sprintf("/rest/api/2/issue/%s", url.PathEscape(key))
	
	if err := c.doWithBody("PUT", path, req, nil); err != nil {
		return nil, err
	}
	
	return c.GetIssue(key)
}
