# Cindy — Agentic CI/CD Release Protocol

Cindy is an open, label-based deployment protocol designed for agentic CI/CD workflows. It defines how code changes move through a deployment pipeline using Git labels as the state machine — no coupling to any specific orchestrator, platform, or agent framework.

Any agent, nim, bot, or human that understands the Cindy label convention can participate in the pipeline.

## Why Cindy exists

Traditional CI/CD assumes humans open PRs and review code. In agentic environments where autonomous agents propose changes at scale, the pipeline needs a shared protocol that:

- Is **decoupled** from any specific orchestrator
- Uses **Git as the source of truth** (branches + labels = state)
- Is **event-driven** (label changes are signals)
- Supports **sequencing and dependency management** between concurrent changes
- Is readable by both humans and machines

## Cindy replaces the PR

Cindy does not use pull requests. The branch + label is the primitive.

- A branch with `cindy:ready` is the proposal
- The orchestrator merges the branch when it reaches `cindy:deployed`
- For `cindy:human-review`, the human reviews the branch diff directly, then applies `cindy:approved` or `cindy:rejected`

## The label state machine

| Label | Meaning |
|-------|---------|
| `cindy:ready` | Change proposed and ready for analysis |
| `cindy:analyzing` | Impact analysis in progress |
| `cindy:approved` | Passed analysis, queued for deployment |
| `cindy:blocked` | Held — dependency hasn't completed |
| `cindy:deploying` | Actively rolling out |
| `cindy:deployed` | Successfully in production |
| `cindy:rejected` | Failed analysis |
| `cindy:rollback` | Rolled back after deployment |
| `cindy:human-review` | Escalated for human decision |
| `cindy:revision-requested` | Feedback provided, changes needed |

### Valid transitions

```
cindy:ready → cindy:analyzing
cindy:analyzing → cindy:approved | cindy:rejected | cindy:human-review | cindy:blocked | cindy:revision-requested
cindy:approved → cindy:deploying | cindy:blocked
cindy:blocked → cindy:approved
cindy:deploying → cindy:deployed | cindy:rollback
cindy:human-review → cindy:approved | cindy:rejected | cindy:revision-requested
cindy:revision-requested → cindy:ready
cindy:deployed → cindy:rollback
```

## Change manifest

Branches include a `.cindy/manifest.json` describing the change:

```json
{
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
  "description": "Add loyalty tier to sale completed events"
}
```

## Schema safety rules

Cindy enforces one hard rule: **schemas can only be extended, never broken.**

- Adding fields: allowed
- Removing fields: rejected (deprecate first)
- Renaming fields: rejected (add new, deprecate old)
- Changing field types: rejected

## Go package

Import the Go package for label constants, transition validation, manifest parsing, review types, and schema safety enforcement:

```go
import "github.com/nimsforest/cindy/go"
```

```go
// Check transitions
cindy.CanTransition(cindy.Ready, cindy.Analyzing) // true
cindy.CanTransition(cindy.Ready, cindy.Deployed)   // false

// Validate schema safety
violations := cindy.ValidateSchemaChanges(manifest)

// Check review resolution
cindy.AllResolved(review) // true if all comments resolved
```

The Go package has zero external dependencies — stdlib only.

## Resources

- [SPEC.md](SPEC.md) — Formal protocol specification
- [schema/](schema/) — JSON Schema for manifest validation
- [examples/](examples/) — Example manifests

## License

MIT
