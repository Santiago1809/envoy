package auditor

import (
	"path/filepath"
	"testing"
)

func TestAuditGoFile(t *testing.T) {
	auditor := New(".")
	results, err := auditor.auditGoFile(filepath.Join("..", "..", "testdata", "fixtures", "go", "main.go"))
	if err != nil {
		t.Fatalf("auditGoFile() error = %v", err)
	}

	found := make(map[string]bool)
	for _, r := range results {
		found[r.Key] = true
	}

	expected := []string{"DB_HOST", "DB_PORT", "API_KEY", "DATABASE_URL", "SECRET_KEY"}
	for _, key := range expected {
		if !found[key] {
			t.Errorf("expected to find key %q", key)
		}
	}
}

func TestAuditJSFile(t *testing.T) {
	auditor := New(".")
	results, err := auditor.auditJSFile(filepath.Join("..", "..", "testdata", "fixtures", "js", "app.js"))
	if err != nil {
		t.Fatalf("auditJSFile() error = %v", err)
	}

	found := make(map[string]bool)
	for _, r := range results {
		found[r.Key] = true
	}

	expected := []string{"DB_HOST", "DB_PORT", "API_KEY"}
	for _, key := range expected {
		if !found[key] {
			t.Errorf("expected to find key %q", key)
		}
	}
}

func TestAuditPythonFile(t *testing.T) {
	auditor := New(".")
	results, err := auditor.auditPythonFile(filepath.Join("..", "..", "testdata", "fixtures", "py", "app.py"))
	if err != nil {
		t.Fatalf("auditPythonFile() error = %v", err)
	}

	found := make(map[string]bool)
	for _, r := range results {
		found[r.Key] = true
	}

	expected := []string{"DB_HOST", "DB_PORT", "API_KEY"}
	for _, key := range expected {
		if !found[key] {
			t.Errorf("expected to find key %q", key)
		}
	}
}

func TestAuditShellFile(t *testing.T) {
	auditor := New(".")
	results, err := auditor.auditShellFile(filepath.Join("..", "..", "testdata", "fixtures", "sh", "script.sh"))
	if err != nil {
		t.Fatalf("auditShellFile() error = %v", err)
	}

	found := make(map[string]bool)
	for _, r := range results {
		found[r.Key] = true
	}

	expected := []string{"DB_HOST", "DB_PORT", "API_KEY"}
	for _, key := range expected {
		if !found[key] {
			t.Errorf("expected to find key %q", key)
		}
	}
}

func TestAuditDir(t *testing.T) {
	envFile := filepath.Join("..", "..", "testdata", "fixtures", ".env.example")
	result, err := AuditDir(
		filepath.Join("..", "..", "testdata", "fixtures"),
		envFile,
		[]Language{LangGo, LangJS, LangPython, LangShell},
		[]string{"node_modules", "vendor"},
		false,
	)
	if err != nil {
		t.Fatalf("AuditDir() error = %v", err)
	}

	if len(result.UsedNotDeclared) == 0 {
		t.Error("expected some used but not declared vars")
	}

	if len(result.DeclaredNotUsed) == 0 {
		t.Error("expected some declared but not used vars")
	}
}

func TestAuditDirNoEnvFile(t *testing.T) {
	result, err := AuditDir(
		filepath.Join("..", "..", "testdata", "fixtures"),
		"",
		[]Language{LangGo},
		[]string{},
		false,
	)
	if err != nil {
		t.Fatalf("AuditDir() error = %v", err)
	}

	if len(result.UsedNotDeclared) == 0 {
		t.Error("expected some used vars")
	}
}

func TestLanguageFromExt(t *testing.T) {
	tests := []struct {
		ext      string
		expected Language
	}{
		{".go", LangGo},
		{".js", LangJS},
		{".ts", LangJS},
		{".jsx", LangJS},
		{".py", LangPython},
		{".sh", LangShell},
		{".txt", ""},
		{".md", ""},
	}

	a := New(".")
	for _, tt := range tests {
		result := a.languageFromExt(tt.ext)
		if result != tt.expected {
			t.Errorf("languageFromExt(%q) = %v, want %v", tt.ext, result, tt.expected)
		}
	}
}

func TestCollectFiles(t *testing.T) {
	auditor := New(filepath.Join("..", "..", "testdata", "fixtures"))
	auditor.SetLanguages([]Language{LangGo})

	files, err := auditor.collectFiles()
	if err != nil {
		t.Fatalf("collectFiles() error = %v", err)
	}

	if len(files) == 0 {
		t.Error("expected to find some Go files")
	}
}

func TestAuditWithLanguageFilter(t *testing.T) {
	result, err := AuditDir(
		filepath.Join("..", "..", "testdata", "fixtures"),
		"",
		[]Language{LangGo},
		[]string{},
		false,
	)
	if err != nil {
		t.Fatalf("AuditDir() error = %v", err)
	}

	found := make(map[Language]bool)
	for _, r := range result.UsedNotDeclared {
		found[r.Language] = true
	}

	if found[LangJS] || found[LangPython] || found[LangShell] {
		t.Error("should only find Go language vars")
	}
}
