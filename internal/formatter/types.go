package formatter

type OutputFormat string

const (
	FormatText OutputFormat = "text"
	FormatJSON OutputFormat = "json"
)

type Formatter interface {
	Render(data any) error
}

type AuditResultJSON struct {
	Timestamp    string   `json:"timestamp"`
	Directory    string   `json:"directory"`
	Language     string   `json:"language"`
	Used         []VarRef `json:"used"`          // Used but not declared
	Declared     []string `json:"declared"`      // Declared but not used
	OnlyUsed     []string `json:"only_used"`     // Same as Used (keys only)
	OnlyDeclared []string `json:"only_declared"` // Same as Declared
}

type VarRef struct {
	Name string `json:"name"`
	File string `json:"file"`
	Line int    `json:"line"`
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
	Timestamp string     `json:"timestamp"`
	File      string     `json:"file"`
	Keys      []KeyEntry `json:"keys"`
}

type KeyEntry struct {
	Name     string `json:"name"`
	HasValue bool   `json:"has_value"`
	Length   int    `json:"length"`
}
