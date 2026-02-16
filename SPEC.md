# Cindy Protocol Specification

Version: 0.1.0

## 1. Overview

Cindy is a label-based deployment protocol for agentic CI/CD workflows. It uses Git branches and labels as the state machine for tracking changes through a deployment pipeline.

## 2. Labels

All labels use the `cindy:` namespace. The following labels are defined:

| Label | Meaning |
|-------|---------|
| `cindy:ready` | Change is proposed and ready for deployment analysis |
| `cindy:analyzing` | Impact analysis in progress |
| `cindy:approved` | Passed analysis, queued for deployment |
| `cindy:blocked` | Held — a dependency hasn't completed yet |
| `cindy:deploying` | Actively rolling out |
| `cindy:deployed` | Successfully in production |
| `cindy:rejected` | Failed analysis (schema break, unsafe change, etc.) |
| `cindy:rollback` | Rolled back after deployment |
| `cindy:human-review` | Escalated, waiting for human decision |
| `cindy:revision-requested` | Feedback provided, changes needed |

## 3. State machine

### 3.1 Valid transitions

```
cindy:ready             → cindy:analyzing
cindy:analyzing         → cindy:approved | cindy:rejected | cindy:human-review | cindy:blocked | cindy:revision-requested
cindy:approved          → cindy:deploying | cindy:blocked
cindy:blocked           → cindy:approved
cindy:deploying         → cindy:deployed | cindy:rollback
cindy:human-review      → cindy:approved | cindy:rejected | cindy:revision-requested
cindy:revision-requested → cindy:ready
cindy:deployed          → cindy:rollback
```

### 3.2 Terminal states

- `cindy:rejected` — no outgoing transitions
- `cindy:rollback` — no outgoing transitions (unless re-submitted as a new revision via `cindy:revision-requested → cindy:ready`)

### 3.3 Label metadata

When a label is applied, the actor SHOULD include metadata containing at minimum:

- `actor` — who/what applied the label
- `reason` — why this transition happened
- `timestamp` — when

Optional metadata fields:

- `dependencies` — list of branch references this change depends on
- `risk_level` — low / medium / high as assessed by the analyzer

## 4. Change manifest

### 4.1 Location

The manifest MUST be placed at `.cindy/manifest.json` in the branch root.

### 4.2 Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `revision` | integer | yes | Revision counter, starts at 1 |
| `responds_to` | string or null | yes | Review ID this revision addresses |
| `subjects_affected` | string[] | yes | Event subjects affected by this change |
| `schema_changes` | SchemaChange[] | yes | Schema modifications (may be empty) |
| `consumers` | string[] | yes | Downstream consumers of affected subjects |
| `risk_self_assessment` | "low" \| "medium" \| "high" | yes | Author's risk assessment |
| `depends_on` | string[] | yes | Branch names this change depends on |
| `description` | string | yes | Human-readable description |

### 4.3 SchemaChange object

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `subject` | string | yes | Event subject being modified |
| `type` | "extension" \| "new" | yes | Extension of existing or new subject |
| `fields_added` | string[] | yes | New fields added |
| `fields_removed` | string[] | yes | Fields removed (must be empty for safe changes) |
| `fields_modified` | string[] | yes | Fields with type changes (must be empty for safe changes) |

## 5. Schema safety rules

These rules are non-negotiable and MUST be enforced by any conforming implementation:

1. **Adding fields**: ALLOWED
2. **Removing fields**: REJECTED — fields must be deprecated, not removed
3. **Renaming fields**: REJECTED — add the new name, deprecate the old
4. **Changing field types**: REJECTED — add a new field with the desired type

## 6. Revision tracking

- First submission: `revision: 1`, `responds_to: null`
- Resubmission after feedback: increment `revision`, set `responds_to` to the review ID
- Revision history is implicit in Git (each push with updated manifest is a revision)

## 7. Reviews

### 7.1 Review object

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | string | yes | Unique review identifier |
| `branch` | string | yes | Branch being reviewed |
| `revision` | integer | yes | Manifest revision this review applies to |
| `actor` | string | yes | Who authored the review |
| `verdict` | "approve" \| "request_changes" \| "comment" | yes | Review outcome |
| `comments` | Comment[] | yes | Review comments (may be empty) |
| `timestamp` | string | yes | ISO 8601 timestamp |

### 7.2 Comment object

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | string | yes | Unique comment identifier |
| `file` | string or null | yes | File path (null for general comments) |
| `line` | integer or null | yes | Line number (null for file-level or general) |
| `body` | string | yes | Feedback text |
| `resolved` | boolean | yes | Whether the author has addressed this |

### 7.3 Resolution rules

1. A `request_changes` verdict triggers `cindy:revision-requested`
2. The branch CANNOT return to `cindy:ready` until ALL unresolved comments from the latest `request_changes` review are marked `resolved: true`
3. An `approve` verdict allows transition to `cindy:approved`
4. A `comment` verdict is informational and does not block any transition

### 7.4 Multiple reviewers

A branch may receive reviews from multiple actors. The branch can only proceed when no unresolved `request_changes` reviews remain.

### 7.5 Storage

Cindy does not prescribe where reviews are stored. The protocol requires:

1. Reviews conform to the schema above
2. Reviews are retrievable by branch and revision
3. Comment resolution state is trackable and updatable

## 8. Integration patterns

Cindy does not prescribe how label events are detected. Common patterns:

1. **Webhook-based**: Watch Git for label events, translate to internal events
2. **Polling**: Periodically check for branches with `cindy:ready`
3. **Git hooks**: Server-side hooks fire on label changes

## 9. Conformance

An implementation is Cindy-conformant if it:

1. Uses only the defined labels in the `cindy:` namespace
2. Only performs valid transitions as defined in section 3.1
3. Enforces schema safety rules as defined in section 5
4. Uses the manifest format as defined in section 4
5. Uses the review format as defined in section 7 (if reviews are supported)
