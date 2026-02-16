package cindy

import (
	"os/exec"
	"testing"
)

func TestShortLabel(t *testing.T) {
	tests := []struct {
		label Label
		want  string
	}{
		{Ready, "ready"},
		{Analyzing, "analyzing"},
		{Approved, "approved"},
		{Blocked, "blocked"},
		{Deploying, "deploying"},
		{Deployed, "deployed"},
		{Rejected, "rejected"},
		{Rollback, "rollback"},
		{HumanReview, "human-review"},
		{RevisionRequested, "revision-requested"},
	}

	for _, tt := range tests {
		got := ShortLabel(tt.label)
		if got != tt.want {
			t.Errorf("ShortLabel(%s) = %q, want %q", tt.label, got, tt.want)
		}
	}
}

func TestTagName(t *testing.T) {
	tests := []struct {
		label  Label
		branch string
		want   string
	}{
		{Approved, "feature/foo", "cindy/approved/feature/foo"},
		{RevisionRequested, "feature/add-loyalty-tier", "cindy/revision-requested/feature/add-loyalty-tier"},
		{Deployed, "main", "cindy/deployed/main"},
		{HumanReview, "feature/deep/nested/branch", "cindy/human-review/feature/deep/nested/branch"},
	}

	for _, tt := range tests {
		got := TagName(tt.label, tt.branch)
		if got != tt.want {
			t.Errorf("TagName(%s, %q) = %q, want %q", tt.label, tt.branch, got, tt.want)
		}
	}
}

func TestParseTag(t *testing.T) {
	tests := []struct {
		tag        string
		wantLabel  Label
		wantBranch string
		wantOK     bool
	}{
		{"cindy/approved/feature/foo", Approved, "feature/foo", true},
		{"cindy/revision-requested/feature/add-loyalty-tier", RevisionRequested, "feature/add-loyalty-tier", true},
		{"cindy/deployed/main", Deployed, "main", true},
		{"cindy/human-review/feature/deep/nested/branch", HumanReview, "feature/deep/nested/branch", true},
		{"cindy/ready/feature/cache-ttl-optimization", Ready, "feature/cache-ttl-optimization", true},
		{"cindy/rollback/hotfix/urgent", Rollback, "hotfix/urgent", true},
		// Not a cindy tag.
		{"v1.0.0", "", "", false},
		{"release/1.0", "", "", false},
		// Cindy prefix but no valid label.
		{"cindy/unknown/feature/foo", "", "", false},
		// Cindy prefix with label but no branch.
		{"cindy/approved/", "", "", false},
		{"cindy/approved", "", "", false},
	}

	for _, tt := range tests {
		label, branch, ok := ParseTag(tt.tag)
		if ok != tt.wantOK {
			t.Errorf("ParseTag(%q): ok = %v, want %v", tt.tag, ok, tt.wantOK)
			continue
		}
		if ok {
			if label != tt.wantLabel {
				t.Errorf("ParseTag(%q): label = %s, want %s", tt.tag, label, tt.wantLabel)
			}
			if branch != tt.wantBranch {
				t.Errorf("ParseTag(%q): branch = %q, want %q", tt.tag, branch, tt.wantBranch)
			}
		}
	}
}

func TestParseTag_RoundTrip(t *testing.T) {
	branches := []string{"feature/foo", "main", "feature/deep/nested/branch", "hotfix/urgent-fix"}
	for _, l := range AllLabels() {
		for _, b := range branches {
			tag := TagName(l, b)
			gotLabel, gotBranch, ok := ParseTag(tag)
			if !ok {
				t.Errorf("ParseTag(TagName(%s, %q)) failed", l, b)
				continue
			}
			if gotLabel != l {
				t.Errorf("round-trip label: got %s, want %s", gotLabel, l)
			}
			if gotBranch != b {
				t.Errorf("round-trip branch: got %q, want %q", gotBranch, b)
			}
		}
	}
}

// initGitRepo creates a bare-minimum git repo with one commit.
func initGitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	cmds := [][]string{
		{"git", "init"},
		{"git", "config", "user.email", "test@test.com"},
		{"git", "config", "user.name", "Test"},
		{"git", "commit", "--allow-empty", "-m", "init"},
	}
	for _, args := range cmds {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git setup %v: %s: %v", args, out, err)
		}
	}
	return dir
}

func TestGitLabeler(t *testing.T) {
	repo := initGitRepo(t)

	gl, err := NewGitLabeler(repo)
	if err != nil {
		t.Fatalf("NewGitLabeler: %v", err)
	}

	// Initially no labels.
	label, err := gl.GetLabel("feature/test")
	if err != nil {
		t.Fatalf("GetLabel: %v", err)
	}
	if label != "" {
		t.Errorf("expected empty label, got %s", label)
	}

	all, err := gl.AllLabels()
	if err != nil {
		t.Fatalf("AllLabels: %v", err)
	}
	if len(all) != 0 {
		t.Errorf("expected empty labels, got %v", all)
	}

	// Set a label.
	if err := gl.SetLabel("feature/test", Approved); err != nil {
		t.Fatalf("SetLabel: %v", err)
	}

	label, err = gl.GetLabel("feature/test")
	if err != nil {
		t.Fatalf("GetLabel: %v", err)
	}
	if label != Approved {
		t.Errorf("expected approved, got %s", label)
	}

	// Set another branch.
	if err := gl.SetLabel("feature/other", Blocked); err != nil {
		t.Fatalf("SetLabel: %v", err)
	}

	all, err = gl.AllLabels()
	if err != nil {
		t.Fatalf("AllLabels: %v", err)
	}
	if len(all) != 2 {
		t.Errorf("expected 2 labels, got %d", len(all))
	}
	if all["feature/test"] != Approved {
		t.Errorf("expected approved for feature/test, got %s", all["feature/test"])
	}
	if all["feature/other"] != Blocked {
		t.Errorf("expected blocked for feature/other, got %s", all["feature/other"])
	}

	// Transition: change label (old tag should be deleted).
	if err := gl.SetLabel("feature/test", Deploying); err != nil {
		t.Fatalf("SetLabel transition: %v", err)
	}

	label, err = gl.GetLabel("feature/test")
	if err != nil {
		t.Fatalf("GetLabel: %v", err)
	}
	if label != Deploying {
		t.Errorf("expected deploying after transition, got %s", label)
	}

	// Verify old tag is gone by checking tag count.
	all, err = gl.AllLabels()
	if err != nil {
		t.Fatalf("AllLabels: %v", err)
	}
	if len(all) != 2 {
		t.Errorf("expected 2 labels after transition (not 3), got %d: %v", len(all), all)
	}
}

func TestGitLabeler_BranchWithSlashes(t *testing.T) {
	repo := initGitRepo(t)

	gl, err := NewGitLabeler(repo)
	if err != nil {
		t.Fatalf("NewGitLabeler: %v", err)
	}

	branch := "feature/deep/nested/branch-name"
	if err := gl.SetLabel(branch, HumanReview); err != nil {
		t.Fatalf("SetLabel: %v", err)
	}

	label, err := gl.GetLabel(branch)
	if err != nil {
		t.Fatalf("GetLabel: %v", err)
	}
	if label != HumanReview {
		t.Errorf("expected human-review, got %s", label)
	}
}

func TestGitLabeler_NotARepo(t *testing.T) {
	dir := t.TempDir()
	_, err := NewGitLabeler(dir)
	if err == nil {
		t.Error("expected error for non-git directory")
	}
}

func TestGitLabeler_NoRemote(t *testing.T) {
	repo := initGitRepo(t)

	gl, err := NewGitLabeler(repo)
	if err != nil {
		t.Fatalf("NewGitLabeler: %v", err)
	}

	// SetLabel should succeed even without a remote — push is best-effort.
	if err := gl.SetLabel("feature/test", Ready); err != nil {
		t.Fatalf("SetLabel without remote: %v", err)
	}

	// Verify the tag exists locally.
	cmd := exec.Command("git", "-C", repo, "tag", "-l", "cindy/*")
	out, _ := cmd.Output()
	if got := string(out); got == "" {
		t.Error("expected tag to exist locally")
	}
}

func TestMemoryLabeler(t *testing.T) {
	ml := NewMemoryLabeler()

	// Initially empty.
	label, err := ml.GetLabel("feature/test")
	if err != nil {
		t.Fatalf("GetLabel: %v", err)
	}
	if label != "" {
		t.Errorf("expected empty, got %s", label)
	}

	all, err := ml.AllLabels()
	if err != nil {
		t.Fatalf("AllLabels: %v", err)
	}
	if len(all) != 0 {
		t.Errorf("expected empty, got %v", all)
	}

	// Set and get.
	if err := ml.SetLabel("feature/test", Approved); err != nil {
		t.Fatalf("SetLabel: %v", err)
	}

	label, err = ml.GetLabel("feature/test")
	if err != nil {
		t.Fatalf("GetLabel: %v", err)
	}
	if label != Approved {
		t.Errorf("expected approved, got %s", label)
	}

	// Overwrite.
	if err := ml.SetLabel("feature/test", Deployed); err != nil {
		t.Fatalf("SetLabel: %v", err)
	}

	label, _ = ml.GetLabel("feature/test")
	if label != Deployed {
		t.Errorf("expected deployed, got %s", label)
	}

	// AllLabels returns copy.
	ml.SetLabel("feature/other", Blocked)
	all, _ = ml.AllLabels()
	if len(all) != 2 {
		t.Errorf("expected 2, got %d", len(all))
	}

	// Mutating returned map doesn't affect internal state.
	all["feature/test"] = Ready
	label, _ = ml.GetLabel("feature/test")
	if label != Deployed {
		t.Errorf("expected internal state unchanged, got %s", label)
	}
}

func TestGitLabeler_TagPrefix(t *testing.T) {
	// Verify the constant value.
	if TagPrefix != "cindy/" {
		t.Errorf("expected TagPrefix = %q, got %q", "cindy/", TagPrefix)
	}
}

