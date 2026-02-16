package cindy

// Verdict represents the outcome of a review.
type Verdict string

const (
	Approve        Verdict = "approve"
	RequestChanges Verdict = "request_changes"
	Comment        Verdict = "comment"
)

// ReviewComment is a single piece of feedback within a review.
type ReviewComment struct {
	ID       string  `json:"id"`
	File     *string `json:"file"`
	Line     *int    `json:"line"`
	Body     string  `json:"body"`
	Resolved bool    `json:"resolved"`
}

// Review is the atomic unit of feedback in the Cindy protocol.
type Review struct {
	ID        string          `json:"id"`
	Branch    string          `json:"branch"`
	Revision  int             `json:"revision"`
	Actor     string          `json:"actor"`
	Verdict   Verdict         `json:"verdict"`
	Comments  []ReviewComment `json:"comments"`
	Timestamp string          `json:"timestamp"`
}

// AllResolved returns true if every comment in the review is marked resolved.
// Returns true for reviews with no comments.
func AllResolved(r *Review) bool {
	for _, c := range r.Comments {
		if !c.Resolved {
			return false
		}
	}
	return true
}

// UnresolvedComments returns the comments that are not yet resolved.
func UnresolvedComments(r *Review) []ReviewComment {
	var unresolved []ReviewComment
	for _, c := range r.Comments {
		if !c.Resolved {
			unresolved = append(unresolved, c)
		}
	}
	return unresolved
}

// IsBlocking returns true if the review's verdict blocks the pipeline
// (i.e., it is a request_changes verdict with unresolved comments).
func IsBlocking(r *Review) bool {
	return r.Verdict == RequestChanges && !AllResolved(r)
}
