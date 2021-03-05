package plugins

import (
	originp "github.com/opensourceways/yabot/prow/plugins"
)

func init() {
	originp.AboutThisBot = originp.GetBotDesc("https://gitee.com/")
}
