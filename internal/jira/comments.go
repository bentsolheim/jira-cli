package jira

import (
	"fmt"
	"net/url"
)

// AddComment adds a comment (Jira wiki markup) to an issue and returns the created comment.
func (c *Client) AddComment(key, body string) (*Comment, error) {
	path := fmt.Sprintf("/rest/api/2/issue/%s/comment", url.PathEscape(key))
	var comment Comment
	if err := c.doWithBody("POST", path, &CommentRequest{Body: body}, &comment); err != nil {
		return nil, err
	}
	return &comment, nil
}
