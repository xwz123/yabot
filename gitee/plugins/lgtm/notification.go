package lgtm

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	originl "github.com/opensourceways/yabot/prow/plugins/lgtm"
)

const (
	consentientDesc = "**LGTM**"
	opposedDesc     = "**NOT LGTM**"
	separator       = ", "
	dirSepa         = "\n- "
)

var (
	notificationStr   = "LGTM NOTIFIER: This PR is %s.\n\nReviewers added `/lgtm` are: %s.\n\nReviewers added `/lgtm cancel` are: %s.\n\nIt still needs review for the codes in each of these directoris:%s\n<details>Git tree hash: %s</details>"
	notificationStrRe = regexp.MustCompile(fmt.Sprintf(notificationStr, "(.*)", "(.*)", "(.*)", "([\\s\\S]*)", "(.*)"))
)

type notification struct {
	consentors map[string]bool
	opponents  map[string]bool
	dirs       []string
	treeHash   string
	commentID  int
}

func (not *notification) GetConsentors() map[string]bool {
	return not.consentors
}

func (not *notification) GetOpponents() map[string]bool {
	return not.opponents
}

func (not *notification) ResetConsentor() {
	not.consentors = map[string]bool{}
}

func (not *notification) ResetOpponents() {
	not.opponents = map[string]bool{}
}

func (not *notification) AddConsentor(consentor string, isReviewer bool) {
	not.consentors[consentor] = isReviewer
	if _, ok := not.opponents[consentor]; ok {
		delete(not.opponents, consentor)
	}
}

func (not *notification) AddOpponent(opponent string, isReviewer bool) {
	not.opponents[opponent] = isReviewer
	if _, ok := not.consentors[opponent]; ok {
		delete(not.consentors, opponent)
	}
}

func (not *notification) ResetDirs(s []string) {
	not.dirs = s
}

func (not *notification) GetDirs() []string {
	return not.dirs
}

func (not *notification) WriteComment(gc *ghclient, org, repo string, prNumber int, ok bool) error {
	r := consentientDesc
	if !ok {
		r = opposedDesc
	}

	s := ""
	if not.dirs != nil && len(not.dirs) > 0 {
		s = fmt.Sprintf("%s%s", dirSepa, strings.Join(not.dirs, dirSepa))
	}

	comment := fmt.Sprintf(
		notificationStr, r,
		reviewerToComment(not.consentors, separator),
		reviewerToComment(not.opponents, separator),
		s,
		not.treeHash,
	)

	if not.commentID == 0 {
		return gc.CreateComment(org, repo, prNumber, comment)
	}
	return gc.UpdatePRComment(org, repo, not.commentID, comment)
}

func loadLGTMNotification(gc *ghclient, org, repo string, prNumber int, sha string) (*notification, bool, error) {
	botname, err := gc.BotName()
	if err != nil {
		return nil, false, err
	}

	comments, err := gc.ListIssueComments(org, repo, prNumber)
	if err != nil {
		return nil, false, err
	}

	split := func(s, sep string) []string {
		if s != "" {
			return strings.Split(s, sep)
		}
		return nil
	}

	n := &notification{treeHash: sha}

	for _, comment := range comments {
		if comment.User.Login != botname {
			continue
		}

		m := notificationStrRe.FindStringSubmatch(comment.Body)
		if m != nil {
			n.commentID = comment.ID

			if m[5] == sha {
				n.consentors = commentToReviewer(m[2], separator)
				n.opponents = commentToReviewer(m[3], separator)
				n.dirs = split(m[4], dirSepa)

				return n, false, nil
			}
			break
		}
	}

	filenames, err := originl.GetChangedFiles(gc, org, repo, prNumber)
	if err != nil {
		return nil, false, err
	}

	n.ResetDirs(genDirs(filenames))
	n.ResetConsentor()
	n.ResetOpponents()
	return n, true, nil
}

func reviewerToComment(r map[string]bool, sep string) string {
	if r == nil || len(r) == 0 {
		return ""
	}

	s := make([]string, 0, len(r))
	for k, v := range r {
		if v {
			s = append(s, fmt.Sprintf("**%s**", k))
		} else {
			s = append(s, k)
		}
	}
	return strings.Join(s, sep)
}

func commentToReviewer(s, sep string) map[string]bool {
	if s != "" {
		a := strings.Split(s, sep)
		m := make(map[string]bool, len(a))

		for _, item := range a {
			r := strings.Trim(item, "**")
			m[r] = (item != r)
		}
		return m
	}

	return map[string]bool{}
}

func genDirs(filenames []string) []string {
	m := map[string]bool{}
	for _, n := range filenames {
		m[filepath.Dir(n)] = true
	}

	if m["."] {
		m["root directory"] = true
		delete(m, ".")
	}

	return mapKeys(m)
}

func mapKeys(m map[string]bool) []string {
	s := make([]string, 0, len(m))
	for k := range m {
		s = append(s, k)
	}
	return s
}
