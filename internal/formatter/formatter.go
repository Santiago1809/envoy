package formatter

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

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
		return fmt.Errorf("unsupported data type for text formatter")
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
	timestamp := time.Now().UTC().Format(time.RFC3339)
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")

	switch v := data.(type) {
	case *auditor.AuditResult:
		usedRefs := toVarRefs(v.UsedNotDeclared)
		declaredKeys := emptyIfNil(v.DeclaredNotUsed)
		out := AuditResultJSON{
			Timestamp:    timestamp,
			Language:     "all",
			Used:         usedRefs,
			Declared:     declaredKeys,
			OnlyUsed:     toStringKeys(usedRefs),
			OnlyDeclared: declaredKeys,
		}
		return enc.Encode(out)

	case *differ.DiffOutput:
		missing, extra := extractKeys(v.Results)
		out := DiffResultJSON{
			Timestamp: timestamp,
			File1:     v.File1,
			File2:     v.File2,
			Missing:   missing,
			Extra:     extra,
		}
		return enc.Encode(out)

	case *check.CheckResult:
		out := CheckResultJSON{
			Timestamp: timestamp,
			EnvFile:   "", // will be set by caller if needed
			Missing:   emptyIfNil(v.MissingKeys),
			Present:   emptyIfNil(v.PresentKeys),
			Empty:     emptyIfNil(v.EmptyKeys),
			Valid:     v.Valid,
		}
		return enc.Encode(out)

	case *parser.EnvFile:
		keys := v.Keys()
		entries := make([]KeyEntry, len(keys))
		for i, k := range keys {
			val, _ := v.Get(k)
			entries[i] = KeyEntry{
				Name:     k,
				HasValue: val != "",
				Length:   len(val),
			}
		}
		out := InfoResultJSON{
			Timestamp: timestamp,
			File:      "", // will be set by caller if needed
			Keys:      entries,
		}
		return enc.Encode(out)

	default:
		return fmt.Errorf("unsupported data type for JSON formatter")
	}
}

func emptyIfNil(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}

func toVarRefs(usage []auditor.EnvUsage) []VarRef {
	refs := make([]VarRef, len(usage))
	for i, u := range usage {
		line := 0
		if len(u.Lines) > 0 {
			line = u.Lines[0]
		}
		refs[i] = VarRef{Name: u.Key, File: u.File, Line: line}
	}
	return refs
}

func toStringKeys(refs []VarRef) []string {
	keys := make([]string, len(refs))
	for i, r := range refs {
		keys[i] = r.Name
	}
	return keys
}

func extractKeys(results []differ.DiffResult) ([]string, []string) {
	missing := []string{}
	extra := []string{}
	for _, r := range results {
		if r.DiffType == differ.DiffTypeMissing {
			missing = append(missing, r.Key)
		} else if r.DiffType == differ.DiffTypeExtra {
			extra = append(extra, r.Key)
		}
	}
	return missing, extra
}

func New(format OutputFormat) Formatter {
	switch format {
	case FormatJSON:
		return &JSONFormatter{}
	default:
		return &TextFormatter{}
	}
}
