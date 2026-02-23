package jira

import (
	"fmt"
	"net/url"
	"strconv"
)

// GetIssue fetches a single issue by key (e.g. "PROJ-123").
func (c *Client) GetIssue(key string) (*Issue, error) {
	var issue Issue
	path := fmt.Sprintf("/rest/api/2/issue/%s", url.PathEscape(key))
	if err := c.do("GET", path, &issue); err != nil {
		return nil, err
	}
	return &issue, nil
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
