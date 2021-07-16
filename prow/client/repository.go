package client

import (
	"context"
	"github.com/google/go-github/v36/github"
)

//IsCollaborator checks whether the specified GitHub user has collaborator
//access to the given repo.
//Note: This will return false if the user is not a collaborator OR the user
// is not a GitHub user.
func (c *client) IsCollaborator(owner, repo, login string) (bool, error) {
	action := "IsCollaborator"
	c.log(action, owner, repo)
	isCollaborator, _, err := c.Repositories.IsCollaborator(context.Background(), owner, repo, login)
	return isCollaborator, c.wrapperError(err, action)
}

//GetSingleCommit  fetches the specified commit, including all details about it.
func (c *client) GetSingleCommit(owner, repo, SHA string) (*github.RepositoryCommit, error) {
	action := "GetSingleCommit"
	c.log(action, owner, repo, SHA)
	commit, _, err := c.Repositories.GetCommit(context.Background(), owner, repo, SHA)
	return commit, c.wrapperError(err, action)
}

