package cindy

// MemoryLabeler is an in-memory Labeler implementation for testing.
type MemoryLabeler struct {
	labels map[string]Label
}

// NewMemoryLabeler creates a new MemoryLabeler.
func NewMemoryLabeler() *MemoryLabeler {
	return &MemoryLabeler{labels: make(map[string]Label)}
}

// GetLabel returns the label for a branch, or ("", nil) if not set.
func (ml *MemoryLabeler) GetLabel(branch string) (Label, error) {
	return ml.labels[branch], nil
}

// SetLabel sets the label for a branch.
func (ml *MemoryLabeler) SetLabel(branch string, label Label) error {
	ml.labels[branch] = label
	return nil
}

// AllLabels returns all labeled branches.
func (ml *MemoryLabeler) AllLabels() (map[string]Label, error) {
	result := make(map[string]Label, len(ml.labels))
	for k, v := range ml.labels {
		result[k] = v
	}
	return result, nil
}
