package parser

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadReader(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]string
		wantErr  bool
	}{
		{
			name:  "basic key-value",
			input: "KEY=value",
			expected: map[string]string{
				"KEY": "value",
			},
			wantErr: false,
		},
		{
			name:  "multiple keys",
			input: "KEY1=value1\nKEY2=value2",
			expected: map[string]string{
				"KEY1": "value1",
				"KEY2": "value2",
			},
			wantErr: false,
		},
		{
			name:  "single quoted value",
			input: "KEY='hello world'",
			expected: map[string]string{
				"KEY": "hello world",
			},
			wantErr: false,
		},
		{
			name:  "double quoted value",
			input: "KEY=\"hello world\"",
			expected: map[string]string{
				"KEY": "hello world",
			},
			wantErr: false,
		},
		{
			name:  "double quoted with escape",
			input: "KEY=\"hello\\nworld\"",
			expected: map[string]string{
				"KEY": "hello\nworld",
			},
			wantErr: false,
		},
		{
			name:  "inline comment",
			input: "KEY=value # comment",
			expected: map[string]string{
				"KEY": "value",
			},
			wantErr: false,
		},
		{
			name:  "empty value",
			input: "KEY=",
			expected: map[string]string{
				"KEY": "",
			},
			wantErr: false,
		},
		{
			name:  "blank line",
			input: "KEY=value\n\nKEY2=value2",
			expected: map[string]string{
				"KEY":  "value",
				"KEY2": "value2",
			},
			wantErr: false,
		},
		{
			name:  "full line comment",
			input: "# this is a comment\nKEY=value",
			expected: map[string]string{
				"KEY": "value",
			},
			wantErr: false,
		},
		{
			name:  "key with underscore",
			input: "MY_KEY=value",
			expected: map[string]string{
				"MY_KEY": "value",
			},
			wantErr: false,
		},
		{
			name:  "key with numbers",
			input: "KEY123=value",
			expected: map[string]string{
				"KEY123": "value",
			},
			wantErr: false,
		},
		{
			name:  "value with equals",
			input: "KEY=value=equals",
			expected: map[string]string{
				"KEY": "value=equals",
			},
			wantErr: false,
		},
		{
			name:  "value with colon",
			input: "KEY=value:colon",
			expected: map[string]string{
				"KEY": "value:colon",
			},
			wantErr: false,
		},
		{
			name:  "backtick quoted value",
			input: "KEY=`value`",
			expected: map[string]string{
				"KEY": "value",
			},
			wantErr: false,
		},
		{
			name:  "trailing spaces",
			input: "KEY=value   ",
			expected: map[string]string{
				"KEY": "value",
			},
			wantErr: false,
		},
		{
			name:  "hash in double quotes not comment",
			input: "KEY=\"value#notcomment\"",
			expected: map[string]string{
				"KEY": "value#notcomment",
			},
			wantErr: false,
		},
		{
			name:  "empty lines and comments",
			input: "# comment\n\nKEY=value\n\n# another comment\nKEY2=value2",
			expected: map[string]string{
				"KEY":  "value",
				"KEY2": "value2",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env, err := LoadReader(bytes.NewReader([]byte(tt.input)))
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadReader() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				for key, expectedVal := range tt.expected {
					val, ok := env.Get(key)
					if !ok {
						t.Errorf("key %q not found", key)
						continue
					}
					if val != expectedVal {
						t.Errorf("key %q = %q, want %q", key, val, expectedVal)
					}
				}
			}
		})
	}
}

func TestEnvFile_Keys(t *testing.T) {
	env := NewEnvFile()
	env.Set("KEY1", "value1")
	env.Set("KEY2", "value2")
	env.Set("KEY3", "value3")

	keys := env.Keys()
	if len(keys) != 3 {
		t.Errorf("Keys() = %d, want 3", len(keys))
	}
	if keys[0] != "KEY1" || keys[1] != "KEY2" || keys[2] != "KEY3" {
		t.Errorf("Keys() = %v, want [KEY1 KEY2 KEY3]", keys)
	}
}

func TestEnvFile_Get(t *testing.T) {
	env := NewEnvFile()
	env.Set("KEY", "value")

	val, ok := env.Get("KEY")
	if !ok {
		t.Error("Get() = false, want true")
	}
	if val != "value" {
		t.Errorf("Get() = %q, want %q", val, "value")
	}

	_, ok = env.Get("NONEXISTENT")
	if ok {
		t.Error("Get() for nonexistent key = true, want false")
	}
}

func TestEnvFile_Set(t *testing.T) {
	env := NewEnvFile()
	env.Set("KEY", "value1")
	env.Set("KEY", "value2")

	val, _ := env.Get("KEY")
	if val != "value2" {
		t.Errorf("Set() = %q, want %q", val, "value2")
	}

	keys := env.Keys()
	if len(keys) != 1 {
		t.Errorf("Keys() after Set() = %d, want 1", len(keys))
	}
}

func TestEnvFile_Write(t *testing.T) {
	env := NewEnvFile()
	env.Set("KEY", "value")
	env.Set("KEY2", "value with space")

	tmpfile, err := os.CreateTemp("", "env")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	err = env.Write(tmpfile.Name())
	if err != nil {
		t.Errorf("Write() error = %v", err)
	}

	content, err := os.ReadFile(tmpfile.Name())
	if err != nil {
		t.Fatal(err)
	}

	expected := `KEY=value
KEY2="value with space"
`
	if string(content) != expected {
		t.Errorf("Write() = %q, want %q", string(content), expected)
	}
}

func TestEnvFile_Delete(t *testing.T) {
	env := NewEnvFile()
	env.Set("KEY", "value")

	err := env.Delete("KEY")
	if err != nil {
		t.Errorf("Delete() error = %v", err)
	}

	_, ok := env.Get("KEY")
	if ok {
		t.Error("Get() after Delete() = true, want false")
	}

	err = env.Delete("NONEXISTENT")
	if err != ErrKeyNotFound {
		t.Errorf("Delete() nonexistent = %v, want ErrKeyNotFound", err)
	}
}

func TestLoad(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "env")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	content := "KEY=value\nKEY2=value2"
	if _, err := tmpfile.WriteString(content); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	env, err := Load(tmpfile.Name())
	if err != nil {
		t.Errorf("Load() error = %v", err)
	}

	val, ok := env.Get("KEY")
	if !ok || val != "value" {
		t.Errorf("Load() key = %q, want %q", val, "value")
	}
}

func TestLoad_fileNotFound(t *testing.T) {
	_, err := Load("/nonexistent/path/.env")
	if err == nil {
		t.Error("Load() for nonexistent file = nil, want error")
	}
}

func TestVariableExpansion(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]string
	}{
		{
			name:  "simple variable",
			input: "HOST=localhost\nURL=postgres://${HOST}/db",
			expected: map[string]string{
				"HOST": "localhost",
				"URL":  "postgres://localhost/db",
			},
		},
		{
			name:  "multiple variables",
			input: "HOST=localhost\nPORT=5432\nURL=postgres://${HOST}:${PORT}/db",
			expected: map[string]string{
				"HOST": "localhost",
				"PORT": "5432",
				"URL":  "postgres://localhost:5432/db",
			},
		},
		{
			name:  "dollar sign variable",
			input: "PRICE=100\nMSG=Price is $PRICE",
			expected: map[string]string{
				"PRICE": "100",
				"MSG":   "Price is $PRICE",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env, err := LoadReader(bytes.NewReader([]byte(tt.input)))
			if err != nil {
				t.Fatalf("LoadReader() error = %v", err)
			}
			env.Expand()
			for key, expectedVal := range tt.expected {
				val, ok := env.Get(key)
				if !ok {
					t.Errorf("key %q not found", key)
					continue
				}
				if val != expectedVal {
					t.Errorf("key %q = %q, want %q", key, val, expectedVal)
				}
			}
		})
	}
}

func TestParseError(t *testing.T) {
	err := &ParseError{Line: 10, Column: 5, msg: "test error"}
	expected := "parse error at line 10, column 5: test error"
	if err.Error() != expected {
		t.Errorf("ParseError.Error() = %q, want %q", err.Error(), expected)
	}
}

func TestPreserveOrder(t *testing.T) {
	input := "Z_KEY=value1\nA_KEY=value2\nM_KEY=value3"
	env, err := LoadReader(bytes.NewReader([]byte(input)))
	if err != nil {
		t.Fatalf("LoadReader() error = %v", err)
	}

	keys := env.Keys()
	if len(keys) != 3 {
		t.Fatalf("Keys() = %d, want 3", len(keys))
	}
	if keys[0] != "Z_KEY" || keys[1] != "A_KEY" || keys[2] != "M_KEY" {
		t.Errorf("Keys() order = %v, want [Z_KEY A_KEY M_KEY]", keys)
	}
}

func TestMultilineValue(t *testing.T) {
	input := `KEY="line1
line2
line3"`
	env, err := LoadReader(bytes.NewReader([]byte(input)))
	if err != nil {
		t.Fatalf("LoadReader() error = %v", err)
	}

	val, ok := env.Get("KEY")
	if !ok {
		t.Fatal("key KEY not found")
	}
	expected := "line1\nline2\nline3"
	if val != expected {
		t.Errorf("KEY = %q, want %q", val, expected)
	}
}

func TestSampleEnvFile(t *testing.T) {
	env, err := Load(filepath.Join("..", "..", "testdata", "sample.env"))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	keys := env.Keys()
	if len(keys) == 0 {
		t.Error("no keys found in sample.env")
	}

	val, ok := env.Get("DATABASE_URL")
	if !ok {
		t.Error("DATABASE_URL key not found")
	}
	if val != "postgres://localhost:5432/db" {
		t.Errorf("DATABASE_URL = %q, want %q", val, "postgres://localhost:5432/db")
	}

	val, ok = env.Get("SINGLE_QUOTED")
	if !ok {
		t.Error("SINGLE_QUOTED key not found")
	}
	if val != "hello world" {
		t.Errorf("SINGLE_QUOTED = %q, want %q", val, "hello world")
	}
}
