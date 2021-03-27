package main

import (
	prowConfig "k8s.io/test-infra/prow/config"
	"k8s.io/test-infra/prow/github"
	"k8s.io/test-infra/prow/pluginhelp"

	"github.com/opensourceways/yabot/gitee/gitee"
	"github.com/opensourceways/yabot/gitee/plugins"
	"github.com/opensourceways/yabot/gitee/plugins/cla"
	"github.com/opensourceways/yabot/gitee/plugins/lgtm"
	originp "github.com/opensourceways/yabot/prow/plugins"
)

func initPlugins(cfg prowConfig.Getter, agent *plugins.ConfigAgent, pm plugins.Plugins, cs *clients) error {
	gpc := func(name string) plugins.PluginConfig {
		return agent.Config().GetPluginConfig(name)
	}

	var v []plugins.Plugin
	v = append(v, cla.NewCLA(gpc, cs.giteeClient))
	v = append(v, lgtm.NewLGTM(gpc,agent.Config,cs.giteeClient,cs.ownersClient))

	for _, i := range v {
		name := i.PluginName()

		i.RegisterEventHandler(pm)
		pm.RegisterHelper(name, i.HelpProvider)

		agent.RegisterPluginConfigBuilder(name, i.NewPluginConfig)
	}

	return nil
}

func genHelpProvider(h plugins.HelpProvider) originp.HelpProvider {
	return func(_ *originp.Configuration, enabledRepos []prowConfig.OrgRepo) (*pluginhelp.PluginHelp, error) {
		return h(enabledRepos)
	}
}

func resetPluginHelper(pm plugins.Plugins) {
	handler := func(originp.Agent, github.GenericCommentEvent) error { return nil }

	hs := pm.HelpProviders()
	for k, h := range hs {
		if h == nil {
			continue
		}
		originp.RegisterGenericCommentHandler(k, handler, genHelpProvider(h))
	}

	//TODO: why is cla special?
	originp.RegisterGenericCommentHandler("cla", handler, claHelpProvider)
}

type pluginHelperAgent struct {
	agent *plugins.ConfigAgent
}

func (p pluginHelperAgent) Config() *originp.Configuration {
	c := p.agent.Config()
	return &originp.Configuration{
		Plugins: c.Plugins,
	}
}

type pluginHelperClient struct {
	c gitee.Client
}

func (p pluginHelperClient) GetRepos(org string, _ bool) ([]github.Repo, error) {
	repos, err := p.c.GetRepos(org)
	if err != nil {
		return nil, err
	}

	r := make([]github.Repo, 0, len(repos))
	for _, item := range repos {
		r = append(r, github.Repo{FullName: item.FullName})
	}
	return r, nil
}

func claHelpProvider(_ *originp.Configuration, enabledRepos []prowConfig.OrgRepo) (*pluginhelp.PluginHelp, error) {
	pluginHelp := &pluginhelp.PluginHelp{
		Description: "The check-cla plugin rechecks the CLA status of a pull request. If the author of pull request has already signed CLA, the label `openlookeng-cla/yes` will be added, otherwise, the label of `openlookeng-cla/no` will be added.",
	}
	pluginHelp.AddCommand(pluginhelp.Command{
		Usage:       "/check-cla",
		Description: "Check the CLA status of PR",
		Featured:    true,
		WhoCanUse:   "Anyone can use the command.",
		Examples:    []string{"/check-cla"},
	})
	return pluginHelp, nil
}
