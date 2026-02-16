package cindy

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// GitLabeler manages Cindy labels as git tags.
type GitLabeler struct {
	repoPath string
}

// NewGitLabeler creates a new GitLabeler for the given repository path.
// Returns an error if the path is not a git repository.
func NewGitLabeler(repoPath string) (*GitLabeler, error) {
	cmd := exec.Command("git", "-C", repoPath, "rev-parse", "--git-dir")
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("not a git repository: %s", repoPath)
	}
	return &GitLabeler{repoPath: repoPath}, nil
}

// GetLabel returns the current Cindy label for a branch by searching git tags.
func (gl *GitLabeler) GetLabel(branch string) (Label, error) {
	tags, err := gl.listTags()
	if err != nil {
		return "", err
	}

	for _, tag := range tags {
		label, tagBranch, ok := ParseTag(tag)
		if ok && tagBranch == branch {
			return label, nil
		}
	}
	return "", nil
}

// SetLabel sets the Cindy label for a branch by creating a git tag.
// Any existing Cindy tag for the branch is deleted first.
// Pushes to remote best-effort (failure is logged to stderr but not returned).
func (gl *GitLabeler) SetLabel(branch string, label Label) error {
	// Delete any existing tag for this branch.
	tags, err := gl.listTags()
	if err != nil {
		return err
	}
	for _, tag := range tags {
		_, tagBranch, ok := ParseTag(tag)
		if ok && tagBranch == branch {
			if err := gl.deleteTag(tag); err != nil {
				return fmt.Errorf("deleting old tag %s: %w", tag, err)
			}
			gl.pushDeleteTag(tag)
		}
	}

	// Create new tag at HEAD.
	newTag := TagName(label, branch)
	cmd := exec.Command("git", "-C", gl.repoPath, "tag", newTag, "HEAD")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("creating tag %s: %s", newTag, strings.TrimSpace(string(out)))
	}

	gl.pushTag(newTag)
	return nil
}

// AllLabels returns all branches with Cindy labels by parsing git tags.
func (gl *GitLabeler) AllLabels() (map[string]Label, error) {
	tags, err := gl.listTags()
	if err != nil {
		return nil, err
	}

	result := make(map[string]Label)
	for _, tag := range tags {
		label, branch, ok := ParseTag(tag)
		if ok {
			result[branch] = label
		}
	}
	return result, nil
}

func (gl *GitLabeler) listTags() ([]string, error) {
	cmd := exec.Command("git", "-C", gl.repoPath, "tag", "-l", TagPrefix+"*")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("listing tags: %w", err)
	}
	raw := strings.TrimSpace(string(out))
	if raw == "" {
		return nil, nil
	}
	return strings.Split(raw, "\n"), nil
}

func (gl *GitLabeler) deleteTag(tag string) error {
	cmd := exec.Command("git", "-C", gl.repoPath, "tag", "-d", tag)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}
	return nil
}

func (gl *GitLabeler) hasRemote() bool {
	cmd := exec.Command("git", "-C", gl.repoPath, "remote", "get-url", "origin")
	return cmd.Run() == nil
}

func (gl *GitLabeler) pushTag(tag string) {
	if !gl.hasRemote() {
		return
	}
	cmd := exec.Command("git", "-C", gl.repoPath, "push", "origin", tag)
	if out, err := cmd.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to push tag %s: %s\n", tag, strings.TrimSpace(string(out)))
	}
}

func (gl *GitLabeler) pushDeleteTag(tag string) {
	if !gl.hasRemote() {
		return
	}
	cmd := exec.Command("git", "-C", gl.repoPath, "push", "origin", "--delete", tag)
	if out, err := cmd.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to delete remote tag %s: %s\n", tag, strings.TrimSpace(string(out)))
	}
}
