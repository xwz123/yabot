package plugins

import (
	"strings"
	"time"

	"gitee.com/openeuler/go-gitee/gitee"
	"k8s.io/test-infra/prow/github"
)

func NoteEventToCommentEvent(e *gitee.NoteEvent) github.GenericCommentEvent {
	gc := github.GenericCommentEvent{
		Repo: github.Repo{
			Owner: github.User{
				Login: e.Repository.Namespace,
			},
			Name: e.Repository.Path,
		},
		User: github.User{
			Login: e.Comment.User.Login,
		},
		Action:  convertNoteEventAction(e),
		Body:    e.Comment.Body,
		HTMLURL: e.Comment.HtmlUrl,
		GUID:    "", //TODO
	}

	switch *(e.NoteableType) {
	case "PullRequest":
		setPullRequestInfo(e, &gc)
	}

	return gc
}

func convertNoteEventAction(e *gitee.NoteEvent) github.GenericCommentEventAction {
	var a github.GenericCommentEventAction

	switch *(e.Action) {
	case "comment":
		a = github.GenericCommentActionCreated
	}
	return a
}

func convertAssignees(assignees []gitee.UserHook) []github.User {
	r := make([]github.User, len(assignees))
	for i, item := range assignees {
		r[i] = github.User{Login: item.Login}
	}
	return r
}

func setPullRequestInfo(e *gitee.NoteEvent, gc *github.GenericCommentEvent) {
	pr := e.PullRequest
	gc.IsPR = true
	gc.IssueState = pr.State
	gc.IssueAuthor.Login = pr.Head.User.Login
	gc.Number = int(pr.Number)
	gc.IssueBody = pr.Body
	gc.IssueHTMLURL = pr.HtmlUrl
	gc.Assignees = convertAssignees(pr.Assignees)
}

func ConvertPullRequestEvent(e *gitee.PullRequestEvent) github.PullRequestEvent {
	epr := e.PullRequest
	pe := github.PullRequestEvent{
		Action: ConvertPullRequestAction(e),
		GUID:   "", //TODO
		PullRequest: github.PullRequest{
			Base: github.PullRequestBranch{
				Repo: github.Repo{
					Name: epr.Base.Repo.Path,
					Owner: github.User{
						Login: epr.Base.Repo.Namespace,
					},
					HTMLURL:  epr.Base.Repo.HtmlUrl,
					FullName: epr.Base.Repo.FullName,
				},
				Ref: epr.Base.Ref,
				SHA: epr.Base.Sha,
			},
			Head: github.PullRequestBranch{
				SHA: epr.Head.Sha,
			},
			User: github.User{
				Login:   epr.Head.User.Login,
				HTMLURL: epr.Head.User.HtmlUrl,
			},
			Number:   int(epr.Number),
			HTMLURL:  epr.HtmlUrl,
			State:    epr.State,
			Body:     epr.Body,
			Title:    epr.Title,
			Labels:   convertPullRequestLabel(e),
			ID:       int(epr.Id),
			Mergable: &(epr.Mergeable),
		},
		// Label:
	}

	return pe
}

func ConvertPushEvent(e *gitee.PushEvent) github.PushEvent {
	pe := github.PushEvent{
		GUID:    "", //TODO
		Ref:     *(e.Ref),
		Deleted: *(e.Deleted),
		After:   *(e.After),
		Repo: github.Repo{
			Owner: github.User{
				Login: e.Repository.Namespace,
			},
			Name: e.Repository.Path,
		},
		Commits: convertPushCommits(e),
		Compare: *(e.Compare),
	}
	return pe
}

func convertPushCommits(e *gitee.PushEvent) []github.Commit {
	r := make([]github.Commit, 0, len(e.Commits))
	for _, i := range e.Commits {
		r = append(r, github.Commit{
			Added:    i.Added,
			Removed:  i.Removed,
			Modified: i.Modified,
			ID:       i.Id,
			Message:  i.Message,
		})
	}
	return r
}

func ConvertPullRequestAction(e *gitee.PullRequestEvent) github.PullRequestEventAction {
	var a github.PullRequestEventAction

	switch strings.ToLower(*(e.Action)) {
	case "open":
		a = github.PullRequestActionOpened
	case "update":
		switch strings.ToLower(*(e.ActionDesc)) {
		case "source_branch_changed": // change the pr's commits
			a = github.PullRequestActionSynchronize
		case "target_branch_changed": // change the branch to which this pr will be merged
			a = github.PullRequestActionEdited
		case "update_label":
			a = github.PullRequestActionLabeled
		}
	case "close":
		a = github.PullRequestActionClosed
	}

	return a
}

func convertPullRequestLabel(e *gitee.PullRequestEvent) []github.Label {
	/*
		r := make([]github.Label, 0, len(e.PullRequest.Labels))

		for _, i := range e.PullRequest.Labels {
			r = append(r, github.Label{Name: i.Name})
		}
		return r
	*/
	return []github.Label{}
}

func ConvertGiteePR(v *gitee.PullRequest) *github.PullRequest  {
	r := github.PullRequest{
		Head: github.PullRequestBranch{
			SHA: v.Head.Sha,
			Ref: v.Head.Ref,
		},
		Base: github.PullRequestBranch{
			Ref: v.Base.Ref,
			SHA: v.Base.Sha,
			Repo: github.Repo{
				Name: v.Base.Repo.Name,
				Owner: github.User{
					Login: v.Base.Repo.Namespace.Path,
				},
				HTMLURL:  v.Base.Repo.HtmlUrl,
				FullName: v.Base.Repo.FullName,
			},
		},
		User: github.User{
			Login:   v.User.Login,
			HTMLURL: v.User.HtmlUrl,
		},

		Number:  int(v.Number),
		HTMLURL: v.HtmlUrl,
		State:   v.State,
		Body:    v.Body,
		Title:   v.Title,
		ID:      int(v.Id),
	}
	return &r
}

func ConvertGiteePRComment(i gitee.PullRequestComments) github.IssueComment {
	ct, _ := time.Parse(time.RFC3339, i.CreatedAt)
	ut, _ := time.Parse(time.RFC3339, i.UpdatedAt)

	return github.IssueComment{
		ID:        int(i.Id),
		Body:      i.Body,
		User:      github.User{Login: i.User.Login},
		HTMLURL:   i.HtmlUrl,
		CreatedAt: ct,
		UpdatedAt: ut,
	}
}