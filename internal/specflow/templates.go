// SPDX-License-Identifier: MIT

package specflow

// Template is a recurring-pattern recipe used by the guided spec
// author. Each template appends extra clarifying questions and merges
// a markdown skeleton into the first draft so the model has a starting
// shape instead of guessing structure from scratch.
type Template struct {
	ID             string
	Description    string
	ExtraQuestions []Question
	Skeleton       string
}

// Modes is the closed set the wizard offers — mirrors
// internal/domain/session.go Mode constants but stays here so
// specflow does not import the domain package.
var Modes = []string{
	"feature",
	"bugfix",
	"design-to-product",
	"optimization",
	"audit",
	"review",
	"setup",
}

// Templates is the bundled catalogue of recurring patterns. Order
// matches the prompt the wizard displays.
var Templates = []Template{
	{
		ID:          "none",
		Description: "no specific pattern (open-ended spec)",
	},
	{
		ID:          "auth",
		Description: "authentication (JWT, refresh, sessions)",
		ExtraQuestions: []Question{
			{Key: "auth_token_lifetime", Prompt: "Access token TTL? (e.g. 15m)", Required: true},
			{Key: "auth_refresh", Prompt: "Refresh-token flow? rotation policy?"},
			{Key: "auth_roles", Prompt: "What roles or scopes ship in the token?"},
		},
		Skeleton: "## Auth contract\n- Token format: \n- Issuer / audience: \n- Token lifetime: \n- Refresh policy: \n- Role/scope claims: \n",
	},
	{
		ID:          "authz",
		Description: "authorisation (RBAC, policies, ACLs)",
		ExtraQuestions: []Question{
			{Key: "authz_actors", Prompt: "Who acts on what? (subject × resource matrix)", Required: true},
			{Key: "authz_policy_store", Prompt: "Where do policies live? (DB, OPA, in-code)"},
			{Key: "authz_deny_default", Prompt: "Deny-by-default or allow-by-default?"},
		},
		Skeleton: "## Authorisation model\n- Subjects: \n- Resources: \n- Actions: \n- Policy store: \n- Default decision: \n",
	},
	{
		ID:          "pagination",
		Description: "list pagination (cursor / offset)",
		ExtraQuestions: []Question{
			{Key: "pag_strategy", Prompt: "Cursor or offset? Why?", Required: true},
			{Key: "pag_page_size", Prompt: "Default page size + max page size?", Required: true},
			{Key: "pag_sort", Prompt: "Sort key + stability (ties broken how?)"},
		},
		Skeleton: "## Pagination contract\n- Strategy: \n- Default size / max: \n- Sort key: \n- Tie-breaker: \n- Response shape: { items: [...], next_cursor: \"...\" }\n",
	},
	{
		ID:          "rate-limit",
		Description: "rate limiting (token-bucket / sliding window)",
		ExtraQuestions: []Question{
			{Key: "rl_algorithm", Prompt: "Token-bucket, leaky-bucket or sliding-window?", Required: true},
			{Key: "rl_key", Prompt: "What key bounds the rate? (ip, api_key, user_id)", Required: true},
			{Key: "rl_response", Prompt: "How does the API signal exceeded? (429 + Retry-After?)"},
		},
		Skeleton: "## Rate-limit contract\n- Algorithm: \n- Key: \n- Quota: \n- Window: \n- Over-limit response: \n",
	},
	{
		ID:          "caching",
		Description: "caching layer (TTL + invalidation)",
		ExtraQuestions: []Question{
			{Key: "cache_backend", Prompt: "Backend? (in-process LRU, redis, CDN)", Required: true},
			{Key: "cache_ttl", Prompt: "Default TTL + per-key overrides?"},
			{Key: "cache_invalidation", Prompt: "What triggers invalidation? (write-through, pub/sub, manual)"},
		},
		Skeleton: "## Caching contract\n- Backend: \n- TTL: \n- Key shape: \n- Invalidation strategy: \n- Cache-miss behaviour: \n",
	},
	{
		ID:          "audit-log",
		Description: "audit log (append-only, structured events)",
		ExtraQuestions: []Question{
			{Key: "audit_storage", Prompt: "Where do events land? (sqlite, postgres, kafka)", Required: true},
			{Key: "audit_schema", Prompt: "Event schema fields? (actor, action, resource, timestamp, …)"},
			{Key: "audit_retention", Prompt: "Retention window + compliance flags?"},
		},
		Skeleton: "## Audit event schema\n- Storage: \n- Retention: \n- Required fields: actor, action, resource, timestamp\n- PII handling: \n",
	},
	{
		ID:          "algorithm-custom",
		Description: "custom algorithm (e.g. round-robin, smart routing)",
		ExtraQuestions: []Question{
			{Key: "algo_invariants", Prompt: "What invariants must hold every iteration?", Required: true},
			{Key: "algo_complexity", Prompt: "Target time/space complexity?"},
			{Key: "algo_benchmarks", Prompt: "Benchmark inputs + comparison baseline?"},
		},
		Skeleton: "## Algorithm\n- Inputs: \n- Outputs: \n- Invariants: \n- Complexity target: \n- Benchmark plan: \n",
	},
}

// LookupTemplate returns the Template matching id, or the "none"
// entry if id is unknown / empty.
func LookupTemplate(id string) Template {
	for _, t := range Templates {
		if t.ID == id {
			return t
		}
	}
	return Templates[0]
}

// QuestionsFor returns the baseline + template-specific questions in
// the order they should be asked.
func QuestionsFor(templateID string) []Question {
	tpl := LookupTemplate(templateID)
	out := make([]Question, 0, len(BaselineQuestions)+len(tpl.ExtraQuestions))
	out = append(out, BaselineQuestions...)
	out = append(out, tpl.ExtraQuestions...)
	return out
}

// SkeletonFor returns the markdown skeleton fragment for the given
// template id (empty when the template has no skeleton).
func SkeletonFor(templateID string) string {
	return LookupTemplate(templateID).Skeleton
}
