package cindy

import (
	"encoding/json"
	"testing"
)

func TestAllLabels(t *testing.T) {
	labels := AllLabels()
	if len(labels) != 10 {
		t.Errorf("expected 10 labels, got %d", len(labels))
	}
}

func TestCanTransition_ValidPaths(t *testing.T) {
	tests := []struct {
		from, to Label
	}{
		{Ready, Analyzing},
		{Analyzing, Approved},
		{Analyzing, Rejected},
		{Analyzing, HumanReview},
		{Analyzing, Blocked},
		{Analyzing, RevisionRequested},
		{Approved, Deploying},
		{Approved, Blocked},
		{Blocked, Approved},
		{Deploying, Deployed},
		{Deploying, Rollback},
		{HumanReview, Approved},
		{HumanReview, Rejected},
		{HumanReview, RevisionRequested},
		{RevisionRequested, Ready},
		{Deployed, Rollback},
	}

	for _, tt := range tests {
		if !CanTransition(tt.from, tt.to) {
			t.Errorf("expected %s → %s to be valid", tt.from, tt.to)
		}
	}
}

func TestCanTransition_InvalidPaths(t *testing.T) {
	tests := []struct {
		from, to Label
	}{
		{Ready, Deployed},
		{Ready, Approved},
		{Analyzing, Deploying},
		{Approved, Deployed},
		{Deployed, Approved},
		{Rejected, Ready},
		{Rollback, Ready},
		{Blocked, Deploying},
		{Deploying, Approved},
		{HumanReview, Deploying},
	}

	for _, tt := range tests {
		if CanTransition(tt.from, tt.to) {
			t.Errorf("expected %s → %s to be invalid", tt.from, tt.to)
		}
	}
}

func TestCanTransition_UnknownLabel(t *testing.T) {
	if CanTransition(Label("cindy:unknown"), Ready) {
		t.Error("expected unknown label to have no valid transitions")
	}
}

func TestValidTransitionsFrom(t *testing.T) {
	targets := ValidTransitionsFrom(Ready)
	if len(targets) != 1 || targets[0] != Analyzing {
		t.Errorf("expected [cindy:analyzing], got %v", targets)
	}

	targets = ValidTransitionsFrom(Analyzing)
	if len(targets) != 5 {
		t.Errorf("expected 5 transitions from analyzing, got %d", len(targets))
	}
}

func TestIsTerminal(t *testing.T) {
	if !IsTerminal(Rejected) {
		t.Error("expected Rejected to be terminal")
	}
	if !IsTerminal(Rollback) {
		t.Error("expected Rollback to be terminal")
	}
	if IsTerminal(Ready) {
		t.Error("expected Ready to not be terminal")
	}
	if IsTerminal(Deployed) {
		t.Error("expected Deployed to not be terminal (can transition to Rollback)")
	}
}

func TestParseManifest(t *testing.T) {
	data := `{
		"revision": 1,
		"responds_to": null,
		"subjects_affected": ["marketing.sale.completed"],
		"schema_changes": [{
			"subject": "marketing.sale.completed",
			"type": "extension",
			"fields_added": ["loyalty_tier"],
			"fields_removed": [],
			"fields_modified": []
		}],
		"consumers": ["aftersales", "analytics"],
		"risk_self_assessment": "medium",
		"depends_on": [],
		"description": "Add loyalty tier"
	}`

	m, err := ParseManifest([]byte(data))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if m.Revision != 1 {
		t.Errorf("expected revision 1, got %d", m.Revision)
	}
	if m.RespondsTo != nil {
		t.Errorf("expected responds_to nil, got %v", m.RespondsTo)
	}
	if len(m.SubjectsAffected) != 1 {
		t.Errorf("expected 1 subject, got %d", len(m.SubjectsAffected))
	}
	if len(m.SchemaChanges) != 1 {
		t.Errorf("expected 1 schema change, got %d", len(m.SchemaChanges))
	}
	if m.SchemaChanges[0].Type != SchemaExtension {
		t.Errorf("expected extension type, got %s", m.SchemaChanges[0].Type)
	}
	if len(m.Consumers) != 2 {
		t.Errorf("expected 2 consumers, got %d", len(m.Consumers))
	}
	if m.RiskSelfAssessment != "medium" {
		t.Errorf("expected medium risk, got %s", m.RiskSelfAssessment)
	}
	if m.Description != "Add loyalty tier" {
		t.Errorf("expected 'Add loyalty tier', got %s", m.Description)
	}
}

func TestParseManifest_WithRespondsTo(t *testing.T) {
	reviewID := "review-001"
	data, _ := json.Marshal(Manifest{
		Revision:          2,
		RespondsTo:        &reviewID,
		SubjectsAffected:  []string{"marketing.sale.completed"},
		SchemaChanges:     []SchemaChange{},
		Consumers:         []string{},
		RiskSelfAssessment: "low",
		DependsOn:         []string{},
		Description:       "Revision 2",
	})

	m, err := ParseManifest(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.RespondsTo == nil || *m.RespondsTo != "review-001" {
		t.Errorf("expected responds_to review-001, got %v", m.RespondsTo)
	}
}

func TestParseManifest_Invalid(t *testing.T) {
	_, err := ParseManifest([]byte("not json"))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestValidateSchemaChanges_Clean(t *testing.T) {
	m := &Manifest{
		SchemaChanges: []SchemaChange{
			{
				Subject:     "marketing.sale.completed",
				Type:        SchemaExtension,
				FieldsAdded: []string{"loyalty_tier"},
			},
		},
	}

	violations := ValidateSchemaChanges(m)
	if len(violations) != 0 {
		t.Errorf("expected no violations, got %d: %v", len(violations), violations)
	}
}

func TestValidateSchemaChanges_FieldRemoval(t *testing.T) {
	m := &Manifest{
		SchemaChanges: []SchemaChange{
			{
				Subject:       "marketing.sale.completed",
				Type:          SchemaExtension,
				FieldsAdded:   []string{"loyalty_tier"},
				FieldsRemoved: []string{"legacy_currency"},
			},
		},
	}

	violations := ValidateSchemaChanges(m)
	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(violations))
	}
	if violations[0].Field != "legacy_currency" {
		t.Errorf("expected legacy_currency violation, got %s", violations[0].Field)
	}
}

func TestValidateSchemaChanges_FieldModification(t *testing.T) {
	m := &Manifest{
		SchemaChanges: []SchemaChange{
			{
				Subject:        "payments.order",
				Type:           SchemaExtension,
				FieldsModified: []string{"amount"},
			},
		},
	}

	violations := ValidateSchemaChanges(m)
	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(violations))
	}
	if violations[0].Field != "amount" {
		t.Errorf("expected amount violation, got %s", violations[0].Field)
	}
}

func TestValidateSchemaChanges_MultipleViolations(t *testing.T) {
	m := &Manifest{
		SchemaChanges: []SchemaChange{
			{
				Subject:        "a.b",
				Type:           SchemaExtension,
				FieldsRemoved:  []string{"x", "y"},
				FieldsModified: []string{"z"},
			},
		},
	}

	violations := ValidateSchemaChanges(m)
	if len(violations) != 3 {
		t.Errorf("expected 3 violations, got %d", len(violations))
	}
}

func TestValidateSchemaChanges_NewSubject(t *testing.T) {
	m := &Manifest{
		SchemaChanges: []SchemaChange{
			{
				Subject:     "payments.refund.completed",
				Type:        SchemaNew,
				FieldsAdded: []string{"refund_id", "amount", "reason"},
			},
		},
	}

	violations := ValidateSchemaChanges(m)
	if len(violations) != 0 {
		t.Errorf("expected no violations for new subject, got %d", len(violations))
	}
}

func TestHasSchemaChanges(t *testing.T) {
	empty := &Manifest{}
	if HasSchemaChanges(empty) {
		t.Error("expected false for empty schema changes")
	}

	with := &Manifest{SchemaChanges: []SchemaChange{{Subject: "a"}}}
	if !HasSchemaChanges(with) {
		t.Error("expected true for non-empty schema changes")
	}
}

func TestHasDependencies(t *testing.T) {
	empty := &Manifest{}
	if HasDependencies(empty) {
		t.Error("expected false for empty depends_on")
	}

	with := &Manifest{DependsOn: []string{"feature/other"}}
	if !HasDependencies(with) {
		t.Error("expected true for non-empty depends_on")
	}
}

func TestAllResolved_Empty(t *testing.T) {
	r := &Review{Comments: []ReviewComment{}}
	if !AllResolved(r) {
		t.Error("expected all resolved for empty comments")
	}
}

func TestAllResolved_AllTrue(t *testing.T) {
	r := &Review{
		Comments: []ReviewComment{
			{ID: "1", Body: "fix this", Resolved: true},
			{ID: "2", Body: "also this", Resolved: true},
		},
	}
	if !AllResolved(r) {
		t.Error("expected all resolved")
	}
}

func TestAllResolved_SomeFalse(t *testing.T) {
	r := &Review{
		Comments: []ReviewComment{
			{ID: "1", Body: "fix this", Resolved: true},
			{ID: "2", Body: "also this", Resolved: false},
		},
	}
	if AllResolved(r) {
		t.Error("expected not all resolved")
	}
}

func TestUnresolvedComments(t *testing.T) {
	r := &Review{
		Comments: []ReviewComment{
			{ID: "1", Body: "fix this", Resolved: true},
			{ID: "2", Body: "also this", Resolved: false},
			{ID: "3", Body: "and this", Resolved: false},
		},
	}
	unresolved := UnresolvedComments(r)
	if len(unresolved) != 2 {
		t.Errorf("expected 2 unresolved, got %d", len(unresolved))
	}
}

func TestIsBlocking(t *testing.T) {
	blocking := &Review{
		Verdict:  RequestChanges,
		Comments: []ReviewComment{{ID: "1", Body: "fix", Resolved: false}},
	}
	if !IsBlocking(blocking) {
		t.Error("expected blocking")
	}

	resolved := &Review{
		Verdict:  RequestChanges,
		Comments: []ReviewComment{{ID: "1", Body: "fix", Resolved: true}},
	}
	if IsBlocking(resolved) {
		t.Error("expected not blocking when all resolved")
	}

	commentOnly := &Review{
		Verdict:  Comment,
		Comments: []ReviewComment{{ID: "1", Body: "note", Resolved: false}},
	}
	if IsBlocking(commentOnly) {
		t.Error("expected comment verdict to not be blocking")
	}
}

func TestSchemaViolation_String(t *testing.T) {
	v := SchemaViolation{Subject: "a.b", Field: "x", Rule: "not allowed"}
	s := v.String()
	if s != `a.b: field "x" — not allowed` {
		t.Errorf("unexpected string: %s", s)
	}

	v2 := SchemaViolation{Subject: "a.b", Rule: "general issue"}
	s2 := v2.String()
	if s2 != "a.b: general issue" {
		t.Errorf("unexpected string: %s", s2)
	}
}
