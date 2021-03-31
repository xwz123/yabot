package approve

import originp "github.com/opensourceways/yabot/prow/plugins"

type configuration struct {
	Approve []originp.Approve `json:"approve,omitempty"`
}

func (c *configuration) Validate() error {
	return nil
}

func (c *configuration) SetDefault() {
}
