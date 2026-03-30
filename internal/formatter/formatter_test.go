package formatter

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Santiago1809/envforge/internal/auditor"
	"github.com/Santiago1809/envforge/internal/check"
	"github.com/Santiago1809/envforge/internal/differ"
	"github.com/Santiago1809/envforge/internal/parser"
)

func loadGolden(t *testing.T, name string) map[string]interface{} {
	t.Helper()
	path := filepath.Join("testdata", "fixtures", "expected_json", name+".json.golden")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read golden file %s: %v", path, err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("failed to parse golden JSON %s: %v", path, err)
	}
	return m
}

func normalizeTimestamps(m map[string]interface{}) {
	if ts, ok := m["timestamp"].(string); ok && strings.HasPrefix(ts, "2026") {
		m["timestamp"] = "2026-03-29T12:00:00Z"
	}
}

func renderToJSON(t *testing.T, data interface{}) string {
	t.Helper()
	// Use a buffer instead of os.Stdout by creating a custom formatter that writes to buffer.
	// Since JSONFormatter hardcodes os.Stdout, we'll temporarily capture it.
	r, w, _ := os.Pipe()
	oldStdout := os.Stdout
	os.Stdout = w
	defer func() {
		w.Close()
		os.Stdout = oldStdout
	}()

	f := &JSONFormatter{}
	if err := f.Render(data); err != nil {
		t.Fatalf("Render error: %v", err)
	}
	w.Close()
	out, _ := io.ReadAll(r)
	return string(out)
}

func TestJSONFormatter_Audit(t *testing.T) {
	golden := loadGolden(t, "audit")
	auditResult := &auditor.AuditResult{
		UsedNotDeclared: []auditor.EnvUsage{
			{Key: "DATABASE_URL", File: "src/db/connection.go", Lines: []int{15}},
			{Key: "JWT_SECRET", File: "src/auth/middleware.go", Lines: []int{8}},
		},
		DeclaredNotUsed: []string{"DEBUG_MODE"},
		DeclaredAndUsed: []string{"DB_HOST"},
	}

	out := renderToJSON(t, auditResult)
	var got map[string]interface{}
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}

	normalizeTimestamps(golden)
	normalizeTimestamps(got)

	// Remove empty arrays if not needed? We'll compare deep equality.
	if !equal(golden, got) {
		t.Errorf("JSON output does not match golden.\nGolden: %s\nGot: %s", mustMarshal(golden), mustMarshal(got))
	}
}

func TestJSONFormatter_Diff(t *testing.T) {
	golden := loadGolden(t, "diff")
	diffOutput := &differ.DiffOutput{
		Results: []differ.DiffResult{
			{Key: "API_KEY", DiffType: differ.DiffTypeMissing},
			{Key: "MY_LOCAL_VAR", DiffType: differ.DiffTypeExtra},
		},
		Summary: differ.DiffSummary{
			MissingCount: 1,
			ExtraCount:   1,
		},
		File1: ".env",
		File2: ".env.example",
	}

	out := renderToJSON(t, diffOutput)
	var got map[string]interface{}
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}

	normalizeTimestamps(golden)
	normalizeTimestamps(got)

	if !equal(golden, got) {
		t.Errorf("JSON output does not match golden.\nGolden: %s\nGot: %s", mustMarshal(golden), mustMarshal(got))
	}
}

func TestJSONFormatter_Check(t *testing.T) {
	golden := loadGolden(t, "check")
	checkResult := &check.CheckResult{
		Valid:         false,
		MissingKeys:   []string{"API_KEY"},
		EmptyKeys:     []string{},
		PresentKeys:   []string{"DATABASE_URL", "DB_HOST"},
		MissingCount:  1,
		EmptyCount:    0,
		TotalRequired: 0,
	}

	out := renderToJSON(t, checkResult)
	var got map[string]interface{}
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}

	normalizeTimestamps(golden)
	normalizeTimestamps(got)

	if !equal(golden, got) {
		t.Errorf("JSON output does not match golden.\nGolden: %s\nGot: %s", mustMarshal(golden), mustMarshal(got))
	}
}

func TestJSONFormatter_Info(t *testing.T) {
	golden := loadGolden(t, "info")
	env := parser.NewEnvFile()
	env.Set("DATABASE_URL", "postgres://localhost:5432/db")
	env.Set("API_KEY", "")

	out := renderToJSON(t, env)
	var got map[string]interface{}
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}

	normalizeTimestamps(golden)
	normalizeTimestamps(got)

	if !equal(golden, got) {
		t.Errorf("JSON output does not match golden.\nGolden: %s\nGot: %s", mustMarshal(golden), mustMarshal(got))
	}
}

func equal(a, b map[string]interface{}) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v1 := range a {
		v2, ok := b[k]
		if !ok {
			return false
		}
		if !deepEqual(v1, v2) {
			return false
		}
	}
	return true
}

func deepEqual(a, b interface{}) bool {
	// Handle common types
	switch va := a.(type) {
	case map[string]interface{}:
		if vb, ok := b.(map[string]interface{}); ok {
			return equal(va, vb)
		}
		return false
	case []interface{}:
		if vb, ok := b.([]interface{}); ok {
			if len(va) != len(vb) {
				return false
			}
			for i := range va {
				if !deepEqual(va[i], vb[i]) {
					return false
				}
			}
			return true
		}
		return false
	case float64:
		if vb, ok := b.(float64); ok {
			return va == vb
		}
		return false
	case string:
		if vb, ok := b.(string); ok {
			return va == vb
		}
		return false
	case bool:
		if vb, ok := b.(bool); ok {
			return va == vb
		}
		return false
	case nil:
		return b == nil
	default:
		return false
	}
}

func mustMarshal(v interface{}) string {
	b, _ := json.MarshalIndent(v, "", "  ")
	return string(b)
}
