package client

import (
	"context"
	"github.com/google/go-github/v36/github"
)

// ListPRCommits lists the all of commits in a pull request.
func (c *client) ListPRCommits(owner, repo string, number int) ([]*github.RepositoryCommit, error) {
	action := "ListPRCommits"
	logDuration := c.log(action, owner, repo, number)
	defer logDuration()
	var commits []*github.RepositoryCommit
	err := doPaginatedRequest(
		func(page, perPage int) (interface{}, *github.Response, error) {
			opt := &github.ListOptions{Page: page, PerPage: perPage}
			return c.PullRequests.ListCommits(context.Background(), owner, repo, number, opt)
		},
		func(obj interface{}) {
			if v, ok := obj.([]*github.RepositoryCommit); ok {
				commits = append(commits, v...)
			}
		})
	return commits, c.wrapperError(err, action)
}

//GetPullRequest a single pull request.
func (c *client) GetPullRequest(owner, repo string, number int) (*github.PullRequest, error) {
	action := "GetPullRequest"
	c.log(action, owner, repo, number)
	pr, _, err := c.PullRequests.Get(context.Background(), owner, repo, number)
	return pr, c.wrapperError(err, action)
}

//GetPullRequestChanges lists the change files in a pull request.
func (c *client) GetPullRequestChanges(owner, repo string, number int) ([]*github.CommitFile, error) {
	action := "GetPullRequestChanges"
	logDuration := c.log(action, owner, repo, number)
	defer logDuration()
	var cfs []*github.CommitFile
	err := doPaginatedRequest(
		func(page, perPage int) (interface{}, *github.Response, error) {
			opt := &github.ListOptions{Page: page, PerPage: perPage}
			return c.PullRequests.ListFiles(context.Background(), owner, repo, number, opt)
		},
		func(obj interface{}) {
			if v, ok := obj.([]*github.CommitFile); ok {
				cfs = append(cfs, v...)
			}
		})
	return cfs, c.wrapperError(err, action)
}

