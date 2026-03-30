package formatter

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/Santiago1809/envforge/internal/auditor"
	"github.com/Santiago1809/envforge/internal/check"
	"github.com/Santiago1809/envforge/internal/differ"
	"github.com/Santiago1809/envforge/internal/parser"
)

type TextFormatter struct{}

func (f *TextFormatter) Render(data any) error {
	switch v := data.(type) {
	case *auditor.AuditResult:
		return f.renderAudit(v)
	case *differ.DiffOutput:
		return f.renderDiff(v)
	case *check.CheckResult:
		return f.renderCheck(v)
	case *parser.EnvFile:
		return f.renderInfo(v)
	default:
		return fmt.Errorf("unsupported data type for text formatter: %T", data)
	}
}

func (f *TextFormatter) renderAudit(result *auditor.AuditResult) error {
	fmt.Println()
	if len(result.UsedNotDeclared) > 0 {
		fmt.Printf("USED but NOT DECLARED (%d):\n", len(result.UsedNotDeclared))
		for _, u := range result.UsedNotDeclared {
			lines := joinInts(u.Lines)
			fmt.Printf("  + %s (%s:%s)\n", u.Key, u.File, lines)
		}
		fmt.Println()
	}

	if len(result.DeclaredNotUsed) > 0 {
		fmt.Printf("DECLARED but NOT USED (%d):\n", len(result.DeclaredNotUsed))
		for _, k := range result.DeclaredNotUsed {
			fmt.Printf("  - %s\n", k)
		}
		fmt.Println()
	}

	if len(result.DeclaredAndUsed) > 0 {
		fmt.Printf("DECLARED and USED (%d):\n", len(result.DeclaredAndUsed))
		for _, k := range result.DeclaredAndUsed {
			fmt.Printf("  = %s\n", k)
		}
	}
	return nil
}

func (f *TextFormatter) renderDiff(result *differ.DiffOutput) error {
	missing := []string{}
	extra := []string{}
	for _, r := range result.Results {
		if r.DiffType == differ.DiffTypeMissing {
			missing = append(missing, r.Key)
		} else if r.DiffType == differ.DiffTypeExtra {
			extra = append(extra, r.Key)
		}
	}

	if result.Summary.MissingCount > 0 || result.Summary.ExtraCount > 0 {
		fmt.Printf("MISSING in %s (%d):\n", result.File1, result.Summary.MissingCount)
		for _, k := range missing {
			fmt.Printf("  + %s\n", k)
		}
		fmt.Println()

		fmt.Printf("EXTRA in %s (%d):\n", result.File2, result.Summary.ExtraCount)
		for _, k := range extra {
			fmt.Printf("  - %s\n", k)
		}
		fmt.Println()
	} else {
		fmt.Println("✓ No differences found")
	}

	return nil
}

func (f *TextFormatter) renderCheck(result *check.CheckResult) error {
	if result.Valid {
		fmt.Println("All required environment variables are set")
		return nil
	}

	if len(result.MissingKeys) > 0 {
		fmt.Println("Missing required environment variables:")
		for _, k := range result.MissingKeys {
			fmt.Printf("  - %s\n", k)
		}
	}

	if len(result.EmptyKeys) > 0 {
		fmt.Println("Required environment variables with empty values:")
		for _, k := range result.EmptyKeys {
			fmt.Printf("  - %s\n", k)
		}
	}

	os.Exit(1)
	return nil
}

func (f *TextFormatter) renderInfo(env *parser.EnvFile) error {
	keys := env.Keys()
	fmt.Printf("Keys: %d\n", len(keys))
	for _, k := range keys {
		val, _ := env.Get(k)
		if val == "" {
			fmt.Printf("  %s (empty)\n", k)
		} else if len(val) > 50 {
			fmt.Printf("  %s = %s...\n", k, val[:50])
		} else {
			fmt.Printf("  %s = %s\n", k, val)
		}
	}
	return nil
}

func joinInts(nums []int) string {
	strs := make([]string, len(nums))
	for i, n := range nums {
		strs[i] = fmt.Sprintf("%d", n)
	}
	return strings.Join(strs, ",")
}

type JSONFormatter struct{}

func (f *JSONFormatter) Render(data any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(data)
}

func New(format OutputFormat) Formatter {
	switch format {
	case FormatJSON:
		return &JSONFormatter{}
	default:
		return &TextFormatter{}
	}
}
