//Package client is an encapsulation of the go-github library.
package client

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-github/v36/github"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"k8s.io/test-infra/prow/config/secret"
)

var lre = regexp.MustCompile(`<([^>]*)>; *rel="([^"]*)"`)

type Client interface {
	IssueAPI
	RepositoryAPI
	PullRequestAPI
	UserAPI
	OrganizationAPI
	TeamAPI
	WithFields(fields logrus.Fields) Client
	ForPlugin(plugin string) Client
}

type client struct {
	*github.Client
	logger *logrus.Entry
}

func (c *client) BotName() (string, error) {
	user, err := c.BotUser()
	if err != nil {
		return "", err
	}
	return *user.Login, nil
}

//BotUser returns the user data of the authenticated identity.
func (c *client) BotUser() (*github.User, error) {
	action := "BotUser"
	c.log(action)
	user, _, err := c.Users.Get(context.Background(), "")
	return user, c.wrapperError(err, action)
}

//IsMember checks if a user is a member of an organization.
func (c *client) IsMember(owner, user string) (bool, error) {
	action := "IsMember"
	c.log(action, owner, user)
	b, _, err := c.Organizations.IsMember(context.Background(), owner, user)
	return b, c.wrapperError(err, action)
}

//ListTeams lists all of the teams for an organization.
func (c *client) ListTeams(owner string) ([]*github.Team, error) {
	action := "ListTeams"
	logDuration := c.log(action, owner)
	defer logDuration()
	var teams []*github.Team
	err := doPaginatedRequest(
		func(page, perPage int) (interface{}, *github.Response, error) {
			opt := &github.ListOptions{Page: page, PerPage: perPage}
			return c.Teams.ListTeams(context.Background(), owner, opt)
		},
		func(obj interface{}) {
			if v, ok := obj.([]*github.Team); ok {
				teams = append(teams, v...)
			}
		})
	return teams, c.wrapperError(err, action)
}

//ListTeamMembers lists all of the users who are members of a team, given a specified
// organization name, by team slug.
// role is filters members returned by their role in the team. Can be one of:
// member,maintainer and all. default is all.
func (c *client) ListTeamMembers(org, slug, role string) ([]*github.User, error) {
	action := "ListTeamMembers"
	logDuration := c.log(action, org, slug, role)
	defer logDuration()
	var tms []*github.User
	err := doPaginatedRequest(
		func(page, perPage int) (interface{}, *github.Response, error) {
			opt := &github.TeamListTeamMembersOptions{
				Role: role,
				ListOptions: github.ListOptions{
					Page:    page,
					PerPage: perPage,
				},
			}
			return c.Teams.ListTeamMembersBySlug(context.Background(), org, slug, opt)
		},
		func(obj interface{}) {
			if v, ok := obj.([]*github.User); ok {
				tms = append(tms, v...)
			}
		})
	return tms, c.wrapperError(err, action)
}

// ForPlugin clones the client, keeping the underlying delegate the same but adding
// a plugin identifier and log field
func (c *client) ForPlugin(plugin string) Client {
	return c.forKeyValue("plugin", plugin)
}

// WithFields clones the client, keeping the underlying delegate the same but adding
// fields to the logging context
func (c *client) WithFields(fields logrus.Fields) Client {
	return &client{Client: c.Client, logger: c.logger.WithFields(fields)}
}

func (c *client) forKeyValue(key, value string) Client {
	return &client{
		Client: c.Client,
		logger: c.logger.WithField(key, value),
	}
}

func (c *client) log(methodName string, args ...interface{}) (logDuration func()) {
	if c.logger == nil {
		return func() {}
	}
	var as []string
	for _, arg := range args {
		as = append(as, fmt.Sprintf("%v", arg))
	}
	start := time.Now()
	c.logger.Infof("%s(%s)", methodName, strings.Join(as, ", "))
	return func() {
		c.logger.WithField("duration", time.Since(start).String()).Debugf("%s(%s) finished", methodName, strings.Join(as, ", "))
	}
}

func (c *client) wrapperError(err error, action string) error {
	if err == nil {
		return err
	}
	return fmt.Errorf("faild to %s: %s", action, err.Error())
}

//NewClientWithFields create a github client.with added logging fields.
//`getToken` is a generator for the GitHub access token to use.
func NewClientWithFields(fields logrus.Fields, getToken func() []byte) Client {
	acToken := string(getToken())
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: acToken})
	tc := oauth2.NewClient(context.Background(), ts)
	gc := github.NewClient(tc)
	return &client{
		Client: gc,
		logger: logrus.WithFields(fields).WithField("client", "github"),
	}
}

//NewClient create a github client.
func NewClient(getToken func() []byte) Client {
	return NewClientWithFields(logrus.Fields{}, getToken)
}

//NewClientWithSecretAndLogFields create a github client.
//'secretAgent' a agent for get token by file path
//'tokenPath' the path of token storage location
func NewClientWithSecretAndLogFields(secretAgent *secret.Agent, fields logrus.Fields, tokenPath string) (Client, error) {
	var generator *func() []byte
	if tokenPath == "" {
		logrus.Warn("empty -github-token-path, will use anonymous github client")
		generatorFunc := func() []byte {
			return []byte{}
		}
		generator = &generatorFunc
	} else {
		if secretAgent == nil {
			return nil, fmt.Errorf("cannot store token from %q without a secret agent", tokenPath)
		}
		generatorFunc := secretAgent.GetTokenGenerator(tokenPath)
		generator = &generatorFunc
	}
	return NewClientWithFields(fields, *generator), nil
}

//doPaginatedRequest iterates over all objects in the paginated result indicated by the 'call' method.
//'call()'  implementation the single page request.
//'accumulate()' should accept that populated  for each page of results.
func doPaginatedRequest(call func(page, perPage int) (interface{}, *github.Response, error), accumulate func(interface{})) error {
	perPage := 20
	page := 1
	for {
		v, response, err := call(page, perPage)
		if err != nil {
			return err
		}
		accumulate(v)
		link := parseLinks(response.Header.Get("Link"))["next"]
		if link == "" {
			break
		}
		pagePath, err := url.Parse(link)
		if err != nil {
			return fmt.Errorf("failed to parse 'next' link: %v", err)
		}
		p := pagePath.Query().Get("page")
		if p == "" {
			return fmt.Errorf("failed to get 'page' on link: %s", p)
		}
		page, err = strconv.Atoi(p)
		if err != nil {
			return err
		}
	}
	return nil
}

// Parse Link headers, returning a map from Rel to URL.
// Only understands the URI and "rel" parameter. Very limited.
// See https://tools.ietf.org/html/rfc5988#section-5
func parseLinks(h string) map[string]string {
	links := map[string]string{}
	for _, m := range lre.FindAllStringSubmatch(h, 10) {
		if len(m) != 3 {
			continue
		}
		links[m[2]] = m[1]
	}
	return links
}
