package check

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCheckRequiredKeys(t *testing.T) {
	os.Setenv("TEST_KEY", "test_value")
	defer os.Unsetenv("TEST_KEY")

	result, err := CheckRequiredKeys([]string{"TEST_KEY"}, false)
	if err != nil {
		t.Fatalf("CheckRequiredKeys() error = %v", err)
	}

	if !result.Valid {
		t.Error("expected valid result")
	}

	if result.MissingCount != 0 {
		t.Errorf("MissingCount = %d, want 0", result.MissingCount)
	}
}

func TestCheckMissingKeys(t *testing.T) {
	os.Unsetenv("NONEXISTENT_KEY")

	result, err := CheckRequiredKeys([]string{"NONEXISTENT_KEY"}, false)
	if err != nil {
		t.Fatalf("CheckRequiredKeys() error = %v", err)
	}

	if result.Valid {
		t.Error("expected invalid result")
	}

	if result.MissingCount != 1 {
		t.Errorf("MissingCount = %d, want 1", result.MissingCount)
	}
}

func TestCheckEmptyKey(t *testing.T) {
	os.Setenv("EMPTY_KEY", "")
	defer os.Unsetenv("EMPTY_KEY")

	result, err := CheckRequiredKeys([]string{"EMPTY_KEY"}, false)
	if err != nil {
		t.Fatalf("CheckRequiredKeys() error = %v", err)
	}

	if result.Valid {
		t.Error("expected invalid result for empty key")
	}

	if result.EmptyCount != 1 {
		t.Errorf("EmptyCount = %d, want 1", result.EmptyCount)
	}
}

func TestCheckAllowEmpty(t *testing.T) {
	os.Setenv("EMPTY_KEY", "")
	defer os.Unsetenv("EMPTY_KEY")

	result, err := CheckRequiredKeys([]string{"EMPTY_KEY"}, true)
	if err != nil {
		t.Fatalf("CheckRequiredKeys() error = %v", err)
	}

	if !result.Valid {
		t.Error("expected valid result when allow empty")
	}
}

func TestCheckFromFile(t *testing.T) {
	dir := t.TempDir()
	envFile := filepath.Join(dir, ".env.example")

	content := "KEY1=\nKEY2=value2\nKEY3=\n"
	os.WriteFile(envFile, []byte(content), 0644)

	result, err := CheckFromFile(envFile, false, "")
	if err != nil {
		t.Fatalf("CheckFromFile() error = %v", err)
	}

	if result.Valid {
		t.Error("expected invalid result due to empty KEY1")
	}

	if result.EmptyCount != 2 {
		t.Errorf("EmptyCount = %d, want 2", result.EmptyCount)
	}
}

func TestCheckPrefix(t *testing.T) {
	os.Setenv("TEST_KEY1", "value1")
	os.Setenv("TEST_KEY2", "value2")
	os.Setenv("OTHER_KEY", "value3")
	defer os.Unsetenv("TEST_KEY1")
	defer os.Unsetenv("TEST_KEY2")
	defer os.Unsetenv("OTHER_KEY")

	opts := &Options{
		Required: []string{"TEST_KEY1", "TEST_KEY2", "OTHER_KEY"},
		Prefix:   "TEST_",
	}

	result, err := Check(opts)
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}

	if result.TotalRequired != 3 {
		t.Errorf("TotalRequired = %d, want 3", result.TotalRequired)
	}

	foundPrefix := false
	for _, key := range result.MissingKeys {
		if key == "OTHER_KEY" {
			foundPrefix = true
		}
	}
	if foundPrefix {
		t.Error("OTHER_KEY should have been filtered by prefix")
	}
}

func TestGetEnv(t *testing.T) {
	os.Setenv("TEST_GET", "test_value")
	defer os.Unsetenv("TEST_GET")

	value, exists := GetEnv("TEST_GET")
	if !exists {
		t.Error("expected key to exist")
	}
	if value != "test_value" {
		t.Errorf("value = %q, want %q", value, "test_value")
	}

	_, exists = GetEnv("NONEXISTENT")
	if exists {
		t.Error("expected key to not exist")
	}
}

func TestGetEnvOrDefault(t *testing.T) {
	os.Setenv("TEST_DEFAULT", "actual")
	defer os.Unsetenv("TEST_DEFAULT")

	value := GetEnvOrDefault("TEST_DEFAULT", "default")
	if value != "actual" {
		t.Errorf("value = %q, want %q", value, "actual")
	}

	value = GetEnvOrDefault("NONEXISTENT", "default")
	if value != "default" {
		t.Errorf("value = %q, want %q", value, "default")
	}
}

func TestGetAllWithPrefix(t *testing.T) {
	os.Setenv("PREFIX_KEY1", "value1")
	os.Setenv("PREFIX_KEY2", "value2")
	os.Setenv("OTHER_KEY", "value3")
	defer os.Unsetenv("PREFIX_KEY1")
	defer os.Unsetenv("PREFIX_KEY2")
	defer os.Unsetenv("OTHER_KEY")

	result := GetAllWithPrefix("PREFIX_")

	if len(result) != 2 {
		t.Errorf("len(result) = %d, want 2", len(result))
	}

	if _, ok := result["PREFIX_KEY1"]; !ok {
		t.Error("expected PREFIX_KEY1 in result")
	}

	if _, ok := result["OTHER_KEY"]; ok {
		t.Error("OTHER_KEY should not be in result")
	}
}

func TestCheckMultipleMissing(t *testing.T) {
	os.Setenv("EXISTING_KEY", "value")
	defer os.Unsetenv("EXISTING_KEY")

	result, err := CheckRequiredKeys([]string{"EXISTING_KEY", "MISSING1", "MISSING2"}, false)
	if err != nil {
		t.Fatalf("CheckRequiredKeys() error = %v", err)
	}

	if result.Valid {
		t.Error("expected invalid result")
	}

	if result.MissingCount != 2 {
		t.Errorf("MissingCount = %d, want 2", result.MissingCount)
	}
}
