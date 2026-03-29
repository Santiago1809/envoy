package differ

import (
	"encoding/json"
	"fmt"

	"envoy/internal/parser"
)

type DiffType string

const (
	DiffTypeMissing DiffType = "MISSING"
	DiffTypeExtra   DiffType = "EXTRA"
	DiffTypeMatch   DiffType = "MATCH"
)

type DiffResult struct {
	Key      string
	DiffType DiffType
	File1Val string
	File2Val string
}

type DiffOutput struct {
	Results []DiffResult
	Summary DiffSummary
	Format  string
}

type DiffSummary struct {
	MissingCount int
	ExtraCount   int
	MatchCount   int
	TotalCount   int
}

type OutputFormat string

const (
	FormatTable  OutputFormat = "table"
	FormatJSON   OutputFormat = "json"
	FormatGitHub OutputFormat = "github"
)

type Differ struct {
	file1Path  string
	file2Path  string
	format     OutputFormat
	showValues bool
	verbose    bool
}

func New(file1Path, file2Path string) *Differ {
	return &Differ{
		file1Path:  file1Path,
		file2Path:  file2Path,
		format:     FormatTable,
		showValues: false,
		verbose:    false,
	}
}

func (d *Differ) SetFormat(format OutputFormat) {
	d.format = format
}

func (d *Differ) SetShowValues(show bool) {
	d.showValues = show
}

func (d *Differ) SetVerbose(verbose bool) {
	d.verbose = verbose
}

func (d *Differ) Diff() (*DiffOutput, error) {
	env1, err := parser.Load(d.file1Path)
	if err != nil {
		return nil, fmt.Errorf("failed to load %s: %w", d.file1Path, err)
	}

	env2, err := parser.Load(d.file2Path)
	if err != nil {
		return nil, fmt.Errorf("failed to load %s: %w", d.file2Path, err)
	}

	keys1 := env1.Keys()
	keys2 := env2.Keys()

	set1 := make(map[string]bool)
	set2 := make(map[string]bool)

	for _, k := range keys1 {
		set1[k] = true
	}
	for _, k := range keys2 {
		set2[k] = true
	}

	var results []DiffResult
	missingCount := 0
	extraCount := 0
	matchCount := 0

	for _, k := range keys2 {
		if !set1[k] {
			val2, _ := env2.Get(k)
			results = append(results, DiffResult{
				Key:      k,
				DiffType: DiffTypeMissing,
				File2Val: val2,
			})
			missingCount++
		}
	}

	for _, k := range keys1 {
		if !set2[k] {
			val1, _ := env1.Get(k)
			results = append(results, DiffResult{
				Key:      k,
				DiffType: DiffTypeExtra,
				File1Val: val1,
			})
			extraCount++
		}
	}

	for _, k := range keys1 {
		if set2[k] {
			val1, _ := env1.Get(k)
			val2, _ := env2.Get(k)
			if val1 == val2 {
				matchCount++
				if d.verbose {
					results = append(results, DiffResult{
						Key:      k,
						DiffType: DiffTypeMatch,
						File1Val: val1,
						File2Val: val2,
					})
				}
			}
		}
	}

	summary := DiffSummary{
		MissingCount: missingCount,
		ExtraCount:   extraCount,
		MatchCount:   matchCount,
		TotalCount:   len(keys1) + len(keys2),
	}

	return &DiffOutput{
		Results: results,
		Summary: summary,
		Format:  string(d.format),
	}, nil
}

func (d *Differ) Print(output *DiffOutput) error {
	switch d.format {
	case FormatJSON:
		return d.printJSON(output)
	case FormatGitHub:
		return d.printGitHub(output)
	default:
		return d.printTable(output)
	}
}

func (d *Differ) printJSON(output *DiffOutput) error {
	data := struct {
		Results []DiffResult `json:"results"`
		Summary DiffSummary  `json:"summary"`
	}{
		Results: output.Results,
		Summary: output.Summary,
	}
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	fmt.Println(string(b))
	return nil
}

func (d *Differ) printGitHub(output *DiffOutput) error {
	for _, r := range output.Results {
		switch r.DiffType {
		case DiffTypeMissing:
			fmt.Printf("::warning file=%s,line=1::Missing env var: %s\n", d.file2Path, r.Key)
		case DiffTypeExtra:
			fmt.Printf("::warning file=%s,line=1::Extra env var (undocumented): %s\n", d.file1Path, r.Key)
		}
	}
	return nil
}

func (d *Differ) printTable(output *DiffOutput) error {
	hasDiffs := output.Summary.MissingCount > 0 || output.Summary.ExtraCount > 0

	if !hasDiffs {
		fmt.Println("✓ No differences found")
		return nil
	}

	fmt.Println()
	if output.Summary.MissingCount > 0 {
		fmt.Printf("MISSING in %s (%d):\n", d.file1Path, output.Summary.MissingCount)
		for _, r := range output.Results {
			if r.DiffType == DiffTypeMissing {
				fmt.Printf("  + %s", r.Key)
				if d.showValues && r.File2Val != "" {
					fmt.Printf(" = %s", r.File2Val)
				}
				fmt.Println()
			}
		}
	}

	if output.Summary.ExtraCount > 0 {
		fmt.Printf("\nEXTRA in %s (%d):\n", d.file1Path, output.Summary.ExtraCount)
		for _, r := range output.Results {
			if r.DiffType == DiffTypeExtra {
				fmt.Printf("  - %s", r.Key)
				if d.showValues && r.File1Val != "" {
					fmt.Printf(" = %s", r.File1Val)
				}
				fmt.Println()
			}
		}
	}

	if d.verbose && output.Summary.MatchCount > 0 {
		fmt.Printf("\nMATCH (%d):\n", output.Summary.MatchCount)
		for _, r := range output.Results {
			if r.DiffType == DiffTypeMatch {
				fmt.Printf("  = %s", r.Key)
				if d.showValues {
					fmt.Printf(" = %s", r.File1Val)
				}
				fmt.Println()
			}
		}
	}

	return nil
}

func DiffFiles(file1, file2 string, format OutputFormat, showValues, verbose bool) (bool, error) {
	d := New(file1, file2)
	d.SetFormat(format)
	d.SetShowValues(showValues)
	d.SetVerbose(verbose)

	output, err := d.Diff()
	if err != nil {
		return false, err
	}

	err = d.Print(output)
	if err != nil {
		return false, err
	}

	hasDiffs := output.Summary.MissingCount > 0 || output.Summary.ExtraCount > 0
	return hasDiffs, nil
}
