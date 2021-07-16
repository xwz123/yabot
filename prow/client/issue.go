package client

import (
	"context"

	"github.com/google/go-github/v36/github"
)

//CreateIssue create an issue under the designated owner/repo.
func (c *client) CreateIssue(owner, repo, title, body string, milestone int, labels, assignees []string) (int, error) {
	action := "CreateIssue"
	logDuration := c.log(action, owner, repo, title)
	defer logDuration()
	issue := &github.IssueRequest{
		Title:     github.String(title),
		Body:      github.String(body),
		Labels:    &labels,
		Milestone: github.Int(milestone),
		Assignees: &assignees,
	}
	is, _, err := c.Issues.Create(context.Background(), owner, repo, issue)
	if err != nil {
		return -1, c.wrapperError(err, action)
	}
	return int(*is.ID), nil
}

//GetIssueLabels list all labels on the specified issue.
func (c *client) GetIssueLabels(owner, repo string, number int) ([]*github.Label, error) {
	action := "GetIssueLabels"
	logDuration := c.log(action, owner, repo, number)
	defer logDuration()
	var labels []*github.Label
	err := doPaginatedRequest(
		func(page, perPage int) (interface{}, *github.Response, error) {
			opts := &github.ListOptions{Page: page, PerPage: perPage}
			return c.Issues.ListLabelsByIssue(context.Background(), owner, repo, number, opts)
		},
		func(v interface{}) {
			labels = append(labels, v.([]*github.Label)...)
		})
	return labels, err
}

//AssignIssue adds the provided GitHub users as assignees to the issue.
func (c *client) AssignIssue(owner, repo string, number int, assignees []string) error {
	action := "AssignIssue"
	logDuration := c.log(action, owner, repo, number)
	defer logDuration()
	_, _, err := c.Issues.AddAssignees(context.Background(), owner, repo, number, assignees)
	return c.wrapperError(err, action)
}

//ListIssueComments lists all comments on the specified issue.
//number of 0 will return all comments on all issues for the repository.
func (c *client) ListIssueComments(owner, repo string, number int) ([]*github.IssueComment, error) {
	action := "ListIssueComments"
	logDuration := c.log(action, owner, repo, number)
	defer logDuration()
	comments, _, err := c.Issues.ListComments(context.Background(), owner, repo, number, nil)
	return comments, c.wrapperError(err, action)
}

//AddLabel add label to an issue
func (c *client) AddLabel(owner, repo string, number int, label string) error {
	//TODO: modify this method to support add labels
	action := "AddLabel"
	c.log(action, owner, repo, number, label)
	labels := []string{label}
	_, _, err := c.Issues.AddLabelsToIssue(context.Background(), owner, repo, number, labels)
	return c.wrapperError(err, action)
}

//RemoveLabel removes a label for an issue.
func (c *client) RemoveLabel(owner, repo string, number int, label string) error {
	action := "RemoveLabel"
	logDuration := c.log(action, owner, repo, number, label)
	defer logDuration()
	_, err := c.Issues.RemoveLabelForIssue(context.Background(), owner, repo, number, label)
	return c.wrapperError(err, action)
}

//CreateComment creates a new comment on the specified issue.
func (c *client) CreateComment(owner, repo string, number int, comment string) error {
	action := "CreateComment"
	c.log(action, owner, repo, number, comment)
	cmt := &github.IssueComment{Body: github.String(comment)}
	_, _, err := c.Issues.CreateComment(context.Background(), owner, repo, number, cmt)
	return c.wrapperError(err, action)
}

//DeleteComment deletes an issue comment.
func (c *client) DeleteComment(owner, repo string, commentID int) error {
	action := "DeleteComment"
	c.log(action, owner, repo, commentID)
	_, err := c.Issues.DeleteComment(context.Background(), owner, repo, int64(commentID))
	return c.wrapperError(err, action)
}
