// SPDX-License-Identifier: MIT

package domain

// CapabilityKind enumerates every plug-in surface HarnessX manages through
// the Catalog. Kinds are frozen by spec p12 — adding a new kind requires a
// new spec entry.
type CapabilityKind string

const (
	KindAgent    CapabilityKind = "agent"
	KindMCP      CapabilityKind = "mcp"
	KindHook     CapabilityKind = "hook"
	KindSensor   CapabilityKind = "sensor"
	KindSkill    CapabilityKind = "skill"
	KindContext  CapabilityKind = "context"
	KindResource CapabilityKind = "resource"
	KindPlugin   CapabilityKind = "plugin"
)

// AllCapabilityKinds enumerates every supported kind in a deterministic
// order, used by listing surfaces (CLI table, HTTP filters).
func AllCapabilityKinds() []CapabilityKind {
	return []CapabilityKind{
		KindAgent, KindMCP, KindHook, KindSensor,
		KindSkill, KindContext, KindResource, KindPlugin,
	}
}

// CapabilityStatus describes the lifecycle stage of a capability for a
// given project.
type CapabilityStatus string

const (
	CapDetected     CapabilityStatus = "detected"
	CapNotInstalled CapabilityStatus = "not_installed"
	CapInstalled    CapabilityStatus = "installed"
	CapConfigured   CapabilityStatus = "configured"
	CapEnabled      CapabilityStatus = "enabled"
	CapDisabled     CapabilityStatus = "disabled"
	CapFailed       CapabilityStatus = "failed"
)

// Capability is the shared view of any item the Catalog manages.
type Capability struct {
	Kind         CapabilityKind   `json:"kind"`
	Name         string           `json:"name"`
	Version      string           `json:"version,omitempty"`
	Source       string           `json:"source,omitempty"` // bundled | user | external
	Status       CapabilityStatus `json:"status"`
	Description  string           `json:"description,omitempty"`
	ManifestPath string           `json:"manifest_path,omitempty"`
	ConfigPath   string           `json:"config_path,omitempty"`
	Tools        int              `json:"tools,omitempty"`
	Transport    string           `json:"transport,omitempty"`
	Scope        string           `json:"scope,omitempty"`
}

// FileOpKind enumerates the limited set of filesystem mutations a Capability
// plan can request. Anything outside this set is rejected by the Executor.
type FileOpKind string

const (
	FileCreate    FileOpKind = "create"
	FileOverwrite FileOpKind = "overwrite"
	FileAppend    FileOpKind = "append"
	FileDelete    FileOpKind = "delete"
	FileMkdir     FileOpKind = "mkdir"
)

// FileOp is a single declarative filesystem mutation. Plans return a slice
// of these; the Executor stages them in a temp dir and commits via rename.
type FileOp struct {
	Op   FileOpKind `json:"op"`
	Path string     `json:"path"` // absolute
	Body []byte     `json:"body,omitempty"`
}
