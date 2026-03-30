package audittypes

type VarRef struct {
	Key        string          `json:"key"`
	References []VarOccurrence `json:"references"`
}

type VarOccurrence struct {
	File     string `json:"file"`
	Lines    []int  `json:"lines"`
	Language string `json:"language"`
}

// AuditResultJSON represents the JSON output for the audit command.
type AuditResultJSON struct {
	Timestamp       string   `json:"timestamp"`
	Directory       string   `json:"directory"`
	Language        string   `json:"language"`
	UsedNotDeclared []VarRef `json:"used_not_declared"`
	DeclaredNotUsed []string `json:"declared_not_used"`
	DeclaredAndUsed []string `json:"declared_and_used"`
}

// DiffResultJSON represents the JSON output for the diff command.
type DiffResultJSON struct {
	Timestamp string   `json:"timestamp"`
	File1     string   `json:"file1"`
	File2     string   `json:"file2"`
	Missing   []string `json:"missing"`
	Extra     []string `json:"extra"`
}

// CheckResultJSON represents the JSON output for the check command.
type CheckResultJSON struct {
	Timestamp string   `json:"timestamp"`
	EnvFile   string   `json:"env_file"`
	Missing   []string `json:"missing"`
	Present   []string `json:"present"`
	Empty     []string `json:"empty"`
	Valid     bool     `json:"valid"`
}

// InfoResultJSON represents the JSON output for the info command.
type InfoResultJSON struct {
	Timestamp string     `json:"timestamp"`
	File      string     `json:"file"`
	Keys      []KeyEntry `json:"keys"`
}

type KeyEntry struct {
	Name     string `json:"name"`
	HasValue bool   `json:"has_value"`
	Length   int    `json:"length"`
}
