package cla

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"k8s.io/test-infra/prow/config"
	"k8s.io/test-infra/prow/github"
	"k8s.io/test-infra/prow/pluginhelp"

	"github.com/opensourceways/yabot/prow/plugins"
)

var (
	checkCLARe = regexp.MustCompile(`(?mi)^/check-cla\s*$`)
	PluginName = "cla"
)

func init() {
	plugins.RegisterGenericCommentHandler(PluginName, handleGenericCommentEvent, helpProvider)
	plugins.RegisterPullRequestHandler(PluginName, handlePullRequestEvent, helpProvider)
	//plugins.RegisterReviewEventHandler(PluginName, handlePullRequestReviewEvent, helpProvider)
}

func helpProvider(config *plugins.Configuration, enabledRepos []config.OrgRepo) (*pluginhelp.PluginHelp, error) {
	pluginHelp := &pluginhelp.PluginHelp{
		Description: "The cla plugin manages the application and removal of the cla labels on pull requests. It is also responsible for warning unauthorized PR authors that they need to sign the cla before their PR will be merged.",
	}
	pluginHelp.AddCommand(pluginhelp.Command{
		Usage:       "/check-cla",
		Description: "Forces rechecking of the CLA status.",
		Featured:    true,
		WhoCanUse:   "Anyone",
		Examples:    []string{"/check-cla"},
	})
	return pluginHelp, nil
}

type githubClient interface {
	AddLabel(owner, repo string, number int, label string) error
	CreateComment(owner, repo string, number int, comment string) error
	RemoveLabel(owner, repo string, number int, label string) error
	ListPRCommits(org, repo string, number int) ([]github.RepositoryCommit, error)
	GetIssueLabels(org, repo string, number int) ([]github.Label, error)
}

func handleGenericCommentEvent(pc plugins.Agent, e github.GenericCommentEvent) error {
	return handleGenericComment(pc.GitHubClient, pc.PluginConfig, pc.Logger, e)
}

func handleGenericComment(gc githubClient, config *plugins.Configuration, log *logrus.Entry, e github.GenericCommentEvent) error {
	funcStart := time.Now()
	defer func() {
		log.WithField("duration", time.Since(funcStart).String()).Debug("Completed handleNoteEvent")
	}()

	// Only consider open PRs and new comments.
	if !e.IsPR || e.IssueState != "open" || e.Action != github.GenericCommentActionCreated {
		return nil
	}

	// Only consider "/check-cla" comments.
	if !checkCLARe.MatchString(e.Body) {
		return nil
	}

	org := e.Repo.Owner.Login
	repo := e.Repo.Name
	prNumber := e.Number

	cfg := config.CLAFor(org, repo)
	if cfg == nil {
		return fmt.Errorf("no config of cla for this repo")
	}

	labels, err := gc.GetIssueLabels(org, repo, prNumber)
	if err != nil {
		log.WithError(err).Error("Failed to get issue labels.")
	}
	hasCLAYes := github.HasLabel(cfg.CLALabelYes, labels)
	hasCLANo := github.HasLabel(cfg.CLALabelNo, labels)

	comment, err := getPrCommitsAbout(gc, org, repo, prNumber, cfg)
	if err != nil {
		return err
	}

	if comment == "" {
		if hasCLANo {
			if err := gc.RemoveLabel(org, repo, prNumber, cfg.CLALabelNo); err != nil {
				log.WithError(err).Warningf("Could not remove %s label.", cfg.CLALabelNo)
			}
		}

		if !hasCLAYes {
			if err := gc.AddLabel(org, repo, prNumber, cfg.CLALabelYes); err != nil {
				log.WithError(err).Warningf("Could not add %s label.", cfg.CLALabelYes)
			}
		}

		return gc.CreateComment(org, repo, prNumber, alreadySigned(e.IssueAuthor.Login))
	}

	if hasCLAYes {
		if err := gc.RemoveLabel(org, repo, prNumber, cfg.CLALabelYes); err != nil {
			log.WithError(err).Warningf("Could not remove %s label.", cfg.CLALabelYes)
		}
	}

	if !hasCLANo {
		if err := gc.AddLabel(org, repo, prNumber, cfg.CLALabelNo); err != nil {
			log.WithError(err).Warningf("Could not add %s label.", cfg.CLALabelNo)
		}
	}

	return gc.CreateComment(org, repo, prNumber, comment)
}

func handlePullRequestEvent(pc plugins.Agent, pre github.PullRequestEvent) error {
	return handlePullRequest(
		pc.Logger,
		pc.GitHubClient,
		pc.PluginConfig,
		&pre,
	)
}

func handlePullRequest(log *logrus.Entry, gc githubClient, config *plugins.Configuration, pe *github.PullRequestEvent) error {
	funcStart := time.Now()
	defer func() {
		log.WithField("duration", time.Since(funcStart).String()).Debug("Completed handlePullRequest")
	}()

	if pe.Action != github.PullRequestActionOpened {
		log.Debug("Pull request state is not open, skipping...")
		return nil
	}

	org := pe.PullRequest.Base.Repo.Owner.Login
	repo := pe.PullRequest.Base.Repo.Name
	prNumber := pe.PullRequest.Number

	cfg := config.CLAFor(org, repo)
	if cfg == nil {
		return fmt.Errorf("no config of cla for this repo")
	}

	comment, err := getPrCommitsAbout(gc, org, repo, prNumber, cfg)
	if err != nil {
		return err
	}
	if comment == "" {
		if err := gc.AddLabel(org, repo, prNumber, cfg.CLALabelYes); err != nil {
			log.WithError(err).Warningf("Could not add %s label.", cfg.CLALabelYes)
		}
		return nil
	}

	if err := gc.AddLabel(org, repo, prNumber, cfg.CLALabelNo); err != nil {
		log.WithError(err).Warningf("Could not add %s label.", cfg.CLALabelNo)
	}
	return gc.CreateComment(org, repo, prNumber, comment)
}

func getPrCommitsAbout(gc githubClient, org, repo string, number int, cfg *plugins.CLA) (string, error) {
	commits, err := gc.ListPRCommits(org, repo, number)
	if err != nil {
		return "", err
	}

	cm := map[string]*github.RepositoryCommit{}
	for i := range commits {
		v := &commits[i]

		email := v.Author.Email
		if _, ok := cm[email]; !ok {
			cm[email] = v
		}
	}

	unSigned, err := checkCommitsSigned(cm, cfg.CheckURL)
	if err != nil {
		return "", err
	}

	if len(unSigned) == 0 {
		return "", nil
	}

	return signGuide(cfg.SignURL, "gitub", generateUnSignComment(unSigned, cm)), nil
}

func checkCommitsSigned(commits map[string]*github.RepositoryCommit, checkURL string) ([]string, error) {
	if len(commits) == 0 {
		return nil, fmt.Errorf("commits is empty, cla cannot be checked")
	}

	var unSigned []string
	for k := range commits {
		signed, err := isSigned(k, checkURL)
		if err != nil {
			return unSigned, err
		}
		if !signed {
			unSigned = append(unSigned, k)
		}
	}
	return unSigned, nil
}

func isSigned(email, url string) (bool, error) {
	endpoint := fmt.Sprintf("%s?email=%s", url, email)

	resp, err := http.Get(endpoint)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	rb, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return false, fmt.Errorf("response has status %q and body %q", resp.Status, string(rb))
	}

	type signingInfo struct {
		Signed bool `json:"signed"`
	}
	var v struct {
		Data signingInfo `json:"data"`
	}

	if err := json.Unmarshal(rb, &v); err != nil {
		return false, fmt.Errorf("unmarshal failed: %s", err.Error())
	}
	return v.Data.Signed, nil
}

func generateUnSignComment(unSigned []string, commits map[string]*github.RepositoryCommit) string {
	if len(unSigned) == 1 {
		return fmt.Sprintf(
			"The author(**%s**) of commit needs to sign cla.",
			commits[unSigned[0]].Author.Login)
	}

	cs := make([]string, 0, len(unSigned))
	for _, v := range unSigned {
		commit := commits[v]
		cs = append(cs, fmt.Sprintf(
			"The author(**%s**) of commit [%s](%s) need to sign cla.",
			commit.Author.Login, commit.SHA[:8], commit.HTMLURL))
	}
	return strings.Join(cs, "\n")

}

func signGuide(signURL, platform, cInfo string) string {
	s := `Thanks for your pull request. Before we can look at your pull request, you'll need to sign a Contributor License Agreement (CLA).

%s

:memo: **Please access [here](%s) to sign the CLA.**

It may take a couple minutes for the CLA signature to be fully registered; after that, please reply here with a new comment: **/check-cla** to verify. Thanks.

---

- If you've already signed a CLA, it's possible you're using a different email address for your %s account. Check your existing CLA data and verify the email. 
- If you signed the CLA as an employee or a member of an organization, please contact your corporation or organization to verify you have been activated to start contributing.
- If you have done the above and are still having issues with the CLA being reported as unsigned, please feel free to file an issue.
	`

	return fmt.Sprintf(s, cInfo, signURL, platform)
}

func alreadySigned(user string) string {
	s := `***@%s***, Thanks for your pull request. All the authors of commits have finished signinig CLA successfully. :wave: `
	return fmt.Sprintf(s, user)
}
