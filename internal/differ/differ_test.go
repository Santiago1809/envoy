package differ

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Santiago1809/envoy/internal/parser"
)

func TestDiff(t *testing.T) {
	dir := t.TempDir()

	file1 := filepath.Join(dir, ".env")
	file2 := filepath.Join(dir, ".env.example")

	os.WriteFile(file1, []byte("KEY1=value1\nKEY2=value2\nKEY3=value3\n"), 0644)
	os.WriteFile(file2, []byte("KEY1=value1\nKEY2=different\nKEY4=\n"), 0644)

	d := New(file1, file2)
	d.SetVerbose(false)

	output, err := d.Diff()
	if err != nil {
		t.Fatalf("Diff() error = %v", err)
	}

	missingCount := 0
	extraCount := 0

	for _, r := range output.Results {
		if r.DiffType == DiffTypeMissing {
			missingCount++
		}
		if r.DiffType == DiffTypeExtra {
			extraCount++
		}
	}

	if missingCount != 1 {
		t.Errorf("Missing count = %d, want 1", missingCount)
	}
	if extraCount != 1 {
		t.Errorf("Extra count = %d, want 1", extraCount)
	}
}

func TestDiffNoDifferences(t *testing.T) {
	dir := t.TempDir()

	file1 := filepath.Join(dir, ".env")
	file2 := filepath.Join(dir, ".env.example")

	os.WriteFile(file1, []byte("KEY1=value1\nKEY2=value2\n"), 0644)
	os.WriteFile(file2, []byte("KEY1=value1\nKEY2=value2\n"), 0644)

	d := New(file1, file2)
	d.SetVerbose(false)

	output, err := d.Diff()
	if err != nil {
		t.Fatalf("Diff() error = %v", err)
	}

	hasDiffs := output.Summary.MissingCount > 0 || output.Summary.ExtraCount > 0
	if hasDiffs {
		t.Error("expected no differences")
	}
}

func TestDiffVerbose(t *testing.T) {
	dir := t.TempDir()

	file1 := filepath.Join(dir, ".env")
	file2 := filepath.Join(dir, ".env.example")

	os.WriteFile(file1, []byte("KEY1=value1\nKEY2=value2\n"), 0644)
	os.WriteFile(file2, []byte("KEY1=value1\nKEY2=value2\n"), 0644)

	d := New(file1, file2)
	d.SetVerbose(true)

	output, err := d.Diff()
	if err != nil {
		t.Fatalf("Diff() error = %v", err)
	}

	matchCount := 0
	for _, r := range output.Results {
		if r.DiffType == DiffTypeMatch {
			matchCount++
		}
	}

	if matchCount != 2 {
		t.Errorf("Match count = %d, want 2", matchCount)
	}
}

func TestDiffJSONFormat(t *testing.T) {
	dir := t.TempDir()

	file1 := filepath.Join(dir, ".env")
	file2 := filepath.Join(dir, ".env.example")

	os.WriteFile(file1, []byte("KEY1=value1\n"), 0644)
	os.WriteFile(file2, []byte("KEY1=value1\nKEY2=\n"), 0644)

	d := New(file1, file2)
	d.SetFormat(FormatJSON)

	output, err := d.Diff()
	if err != nil {
		t.Fatalf("Diff() error = %v", err)
	}

	if output.Format != "json" {
		t.Errorf("Format = %s, want json", output.Format)
	}
}

func TestDiffGitHubFormat(t *testing.T) {
	dir := t.TempDir()

	file1 := filepath.Join(dir, ".env")
	file2 := filepath.Join(dir, ".env.example")

	os.WriteFile(file1, []byte("KEY1=value1\n"), 0644)
	os.WriteFile(file2, []byte("KEY1=value1\nKEY2=\n"), 0644)

	d := New(file1, file2)
	d.SetFormat(FormatGitHub)

	_, err := d.Diff()
	if err != nil {
		t.Fatalf("Diff() error = %v", err)
	}
}

func TestDiffFiles(t *testing.T) {
	dir := t.TempDir()

	file1 := filepath.Join(dir, "env1")
	file2 := filepath.Join(dir, "env2")

	os.WriteFile(file1, []byte("A=1\nB=2\n"), 0644)
	os.WriteFile(file2, []byte("A=1\nC=3\n"), 0644)

	hasDiffs, err := DiffFiles(file1, file2, FormatTable, false, false)
	if err != nil {
		t.Fatalf("DiffFiles() error = %v", err)
	}

	if !hasDiffs {
		t.Error("expected differences")
	}
}

func TestDiffFilesNoDiffs(t *testing.T) {
	dir := t.TempDir()

	file1 := filepath.Join(dir, "env1")
	file2 := filepath.Join(dir, "env2")

	os.WriteFile(file1, []byte("A=1\n"), 0644)
	os.WriteFile(file2, []byte("A=1\n"), 0644)

	hasDiffs, err := DiffFiles(file1, file2, FormatTable, false, false)
	if err != nil {
		t.Fatalf("DiffFiles() error = %v", err)
	}

	if hasDiffs {
		t.Error("expected no differences")
	}
}

func TestDiffMissingKeys(t *testing.T) {
	dir := t.TempDir()

	file1 := filepath.Join(dir, ".env")
	file2 := filepath.Join(dir, ".env.example")

	env1 := parser.NewEnvFile()
	env1.Set("KEY1", "value1")

	env1File, _ := os.Create(file1)
	env1.Write(env1File.Name())
	env1File.Close()

	env2File, _ := os.Create(file2)
	env2File.WriteString("KEY1=\nKEY2=\n")
	env2File.Close()

	d := New(file1, file2)
	output, _ := d.Diff()

	if output.Summary.MissingCount != 1 {
		t.Errorf("MissingCount = %d, want 1", output.Summary.MissingCount)
	}
}

func TestDiffExtraKeys(t *testing.T) {
	dir := t.TempDir()

	file1 := filepath.Join(dir, ".env")
	file2 := filepath.Join(dir, ".env.example")

	env1 := parser.NewEnvFile()
	env1.Set("KEY1", "value1")
	env1.Set("KEY2", "value2")

	env1File, _ := os.Create(file1)
	env1.Write(env1File.Name())
	env1File.Close()

	env2File, _ := os.Create(file2)
	env2File.WriteString("KEY1=\n")
	env2File.Close()

	d := New(file1, file2)
	output, _ := d.Diff()

	if output.Summary.ExtraCount != 1 {
		t.Errorf("ExtraCount = %d, want 1", output.Summary.ExtraCount)
	}
}
