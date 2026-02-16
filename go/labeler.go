package cindy

import "strings"

// TagPrefix is the prefix for all Cindy git tags.
const TagPrefix = "cindy/"

// Labeler manages Cindy labels for branches.
type Labeler interface {
	// GetLabel returns the current label for a branch, or ("", nil) if unlabeled.
	GetLabel(branch string) (Label, error)
	// SetLabel sets the label for a branch, replacing any existing label.
	SetLabel(branch string, label Label) error
	// AllLabels returns all currently labeled branches.
	AllLabels() (map[string]Label, error)
}

// ShortLabel strips the "cindy:" prefix from a label.
// For example, "cindy:approved" becomes "approved".
func ShortLabel(l Label) string {
	return strings.TrimPrefix(string(l), "cindy:")
}

// TagName returns the git tag name for a label and branch.
// For example, TagName(Approved, "feature/foo") returns "cindy/approved/feature/foo".
func TagName(label Label, branch string) string {
	return TagPrefix + ShortLabel(label) + "/" + branch
}

// ParseTag parses a git tag name into a label and branch.
// It tries each known label to find the matching prefix.
// Returns (label, branch, true) on success, or ("", "", false) if the tag is not a Cindy tag.
func ParseTag(tag string) (Label, string, bool) {
	if !strings.HasPrefix(tag, TagPrefix) {
		return "", "", false
	}
	rest := strings.TrimPrefix(tag, TagPrefix)

	for _, l := range allLabels() {
		short := ShortLabel(l)
		prefix := short + "/"
		if strings.HasPrefix(rest, prefix) {
			branch := strings.TrimPrefix(rest, prefix)
			if branch != "" {
				return l, branch, true
			}
		}
	}
	return "", "", false
}

// allLabels is a private alias to avoid shadowing the Labeler interface method name.
func allLabels() []Label {
	return AllLabels()
}
