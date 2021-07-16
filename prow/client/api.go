package client

import "github.com/google/go-github/v36/github"

// IssueAPI interface for issue related API actions
type IssueAPI interface {
	CreateIssue(owner, repo, title, body string, milestone int, labels, assignees []string) (int, error)
	GetIssueLabels(owner, repo string, number int) ([]*github.Label, error)
	AssignIssue(owner, repo string, number int, assignees []string) error
	ListIssueComments(owner, repo string, number int) ([]*github.IssueComment, error)
	AddLabel(owner, repo string, number int, label string) error
	RemoveLabel(owner, repo string, number int, label string) error
	CreateComment(owner, repo string, number int, comment string) error
	DeleteComment(owner, repo string, commentID int) error
}

// RepositoryAPI interface for repository related API actions
type RepositoryAPI interface {
	IsCollaborator(owner, repo, login string) (bool, error)
	GetSingleCommit(owner, repo, SHA string) (*github.RepositoryCommit, error)
}

// PullRequestAPI interface for pull request related API actions
type PullRequestAPI interface {
	ListPRCommits(owner, repo string, number int) ([]*github.RepositoryCommit, error)
	GetPullRequest(owner, repo string, number int) (*github.PullRequest, error)
	GetPullRequestChanges(owner, repo string, number int) ([]*github.CommitFile, error)
}

// UserAPI interface for user related API actions
type UserAPI interface {
	BotName() (string, error)
	BotUser() (*github.User, error)
}

// OrganizationAPI interface for organisation related API actions
type OrganizationAPI interface {
	IsMember(owner, user string) (bool, error)
}

// TeamAPI interface for team related API actions
type TeamAPI interface {
	ListTeams(owner string) ([]*github.Team, error)
	ListTeamMembers(org, slug, role string) ([]*github.User, error)
}
