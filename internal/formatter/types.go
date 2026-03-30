package formatter

import (
	"github.com/Santiago1809/envforge/internal/audittypes"
)

type OutputFormat string

const (
	FormatText OutputFormat = "text"
	FormatJSON OutputFormat = "json"
)

type Formatter interface {
	Render(data any) error
}

type AuditResultJSON struct {
	Timestamp       string              `json:"timestamp"`
	Directory       string              `json:"directory"`
	Language        string              `json:"language"`
	UsedNotDeclared []audittypes.VarRef `json:"used_not_declared"`
	DeclaredNotUsed []string            `json:"declared_not_used"`
	DeclaredAndUsed []string            `json:"declared_and_used"`
}

type DiffResultJSON struct {
	Timestamp string   `json:"timestamp"`
	File1     string   `json:"file1"`
	File2     string   `json:"file2"`
	Missing   []string `json:"missing"` // in file1 but not in file2
	Extra     []string `json:"extra"`   // in file2 but not in file1
}

type CheckResultJSON struct {
	Timestamp string   `json:"timestamp"`
	EnvFile   string   `json:"env_file"`
	Missing   []string `json:"missing"`
	Present   []string `json:"present"`
	Empty     []string `json:"empty"`
	Valid     bool     `json:"valid"`
}

type InfoResultJSON struct {
	Timestamp string                `json:"timestamp"`
	File      string                `json:"file"`
	Keys      []audittypes.KeyEntry `json:"keys"`
}
