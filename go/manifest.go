package cindy

import (
	"encoding/json"
	"fmt"
	"os"
)

// SchemaChangeType describes the kind of schema change.
type SchemaChangeType string

const (
	SchemaExtension SchemaChangeType = "extension"
	SchemaNew       SchemaChangeType = "new"
)

// SchemaChange describes a single schema modification within a manifest.
type SchemaChange struct {
	Subject        string           `json:"subject"`
	Type           SchemaChangeType `json:"type"`
	FieldsAdded    []string         `json:"fields_added"`
	FieldsRemoved  []string         `json:"fields_removed"`
	FieldsModified []string         `json:"fields_modified"`
}

// Manifest is the Cindy change manifest placed at .cindy/manifest.json in a branch.
type Manifest struct {
	Revision          int            `json:"revision"`
	RespondsTo        *string        `json:"responds_to"`
	SubjectsAffected  []string       `json:"subjects_affected"`
	SchemaChanges     []SchemaChange `json:"schema_changes"`
	Consumers         []string       `json:"consumers"`
	RiskSelfAssessment string        `json:"risk_self_assessment"`
	DependsOn         []string       `json:"depends_on"`
	Description       string         `json:"description"`
}

// ParseManifest parses a Cindy manifest from JSON bytes.
func ParseManifest(data []byte) (*Manifest, error) {
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parsing manifest: %w", err)
	}
	return &m, nil
}

// LoadManifest reads and parses a Cindy manifest from a file path.
func LoadManifest(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading manifest: %w", err)
	}
	return ParseManifest(data)
}

// SchemaViolation describes a schema safety rule violation.
type SchemaViolation struct {
	Subject string
	Field   string
	Rule    string
}

func (v SchemaViolation) String() string {
	if v.Field != "" {
		return fmt.Sprintf("%s: field %q — %s", v.Subject, v.Field, v.Rule)
	}
	return fmt.Sprintf("%s: %s", v.Subject, v.Rule)
}

// ValidateSchemaChanges checks that all schema changes in the manifest comply
// with Cindy's schema safety rules:
//   - Adding fields: allowed
//   - Removing fields: rejected
//   - Modifying field types: rejected
//
// Returns a list of violations. An empty list means the manifest is safe.
func ValidateSchemaChanges(m *Manifest) []SchemaViolation {
	var violations []SchemaViolation

	for _, sc := range m.SchemaChanges {
		for _, f := range sc.FieldsRemoved {
			violations = append(violations, SchemaViolation{
				Subject: sc.Subject,
				Field:   f,
				Rule:    "field removal not allowed (deprecate instead)",
			})
		}
		for _, f := range sc.FieldsModified {
			violations = append(violations, SchemaViolation{
				Subject: sc.Subject,
				Field:   f,
				Rule:    "field type modification not allowed (add new field instead)",
			})
		}
	}

	return violations
}

// HasSchemaChanges returns true if the manifest declares any schema changes.
func HasSchemaChanges(m *Manifest) bool {
	return len(m.SchemaChanges) > 0
}

// HasDependencies returns true if the manifest declares dependencies on other branches.
func HasDependencies(m *Manifest) bool {
	return len(m.DependsOn) > 0
}
