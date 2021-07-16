package plugins

import (
	"strings"

	sdk "github.com/google/go-github/v36/github"
)

// HasLabel checks if label is in the label set "issueLabels".
func HasLabel(label string, issueLabels []*sdk.Label) bool {
	for _, l := range issueLabels {
		if strings.ToLower(*l.Name) == strings.ToLower(label) {
			return true
		}
	}
	return false
}
