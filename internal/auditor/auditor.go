package auditor

import (
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

type Language string

const (
	LangGo     Language = "go"
	LangJS     Language = "js"
	LangPython Language = "py"
	LangShell  Language = "sh"
	LangAll    Language = "all"
)

type EnvUsage struct {
	Key      string
	File     string
	Lines    []int
	Language Language
}

type AuditResult struct {
	UsedNotDeclared []EnvUsage
	DeclaredNotUsed []string
	DeclaredAndUsed []string
}

type Auditor struct {
	rootDir   string
	languages []Language
	exclude   []string
	envFile   string
	results   []EnvUsage
	mu        sync.Mutex
	declared  map[string]bool
	wg        sync.WaitGroup
	ctx       context.Context
}

var DefaultExclusions = []string{"testdata", "vendor", "node_modules", ".git", "dist", "build", "bin"}

func New(rootDir string) *Auditor {
	return &Auditor{
		rootDir:   rootDir,
		languages: []Language{LangGo, LangJS, LangPython, LangShell},
		exclude:   DefaultExclusions,
		declared:  make(map[string]bool),
	}
}

func (a *Auditor) SetLanguages(langs []Language) {
	a.languages = langs
}

func (a *Auditor) SetExclude(exclude []string) {
	seen := make(map[string]bool)
	for _, e := range a.exclude {
		seen[e] = true
	}
	for _, e := range exclude {
		if !seen[e] {
			a.exclude = append(a.exclude, e)
		}
	}
}

func (a *Auditor) SetEnvFile(envFile string) {
	a.envFile = envFile
}

func (a *Auditor) SetContext(ctx context.Context) {
	a.ctx = ctx
}

func (a *Auditor) Run() (*AuditResult, error) {
	if a.envFile != "" {
		if err := a.loadDeclaredVars(); err != nil {
			return nil, fmt.Errorf("failed to load env file: %w", err)
		}
	}

	files, err := a.collectFiles()
	if err != nil {
		return nil, err
	}

	a.wg.Add(len(files))
	for _, file := range files {
		go a.auditFile(file)
	}

	a.wg.Wait()

	return a.buildResult()
}

func (a *Auditor) loadDeclaredVars() error {
	data, err := os.ReadFile(a.envFile)
	if err != nil {
		return err
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if idx := strings.Index(line, "="); idx > 0 {
			key := strings.TrimSpace(line[:idx])
			if key != "" {
				a.declared[key] = true
			}
		}
	}
	return nil
}

func (a *Auditor) collectFiles() ([]string, error) {
	var files []string

	err := filepath.WalkDir(a.rootDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if a.ctx != nil && a.ctx.Err() != nil {
			return a.ctx.Err()
		}

		if d.IsDir() {
			name := d.Name()
			for _, ex := range a.exclude {
				if name == ex {
					return filepath.SkipDir
				}
			}
			return nil
		}

		ext := filepath.Ext(path)
		lang := a.languageFromExt(ext)
		if lang == "" {
			return nil
		}

		for _, l := range a.languages {
			if l == LangAll || l == lang {
				files = append(files, path)
				break
			}
		}

		return nil
	})

	return files, err
}

func (a *Auditor) languageFromExt(ext string) Language {
	switch ext {
	case ".go":
		return LangGo
	case ".js", ".jsx", ".ts", ".tsx", ".mjs", ".cjs":
		return LangJS
	case ".py":
		return LangPython
	case ".sh", ".bash", ".zsh":
		return LangShell
	}
	return ""
}

func (a *Auditor) auditFile(path string) {
	defer a.wg.Done()

	if a.ctx != nil && a.ctx.Err() != nil {
		return
	}

	ext := filepath.Ext(path)
	lang := a.languageFromExt(ext)

	var vars []EnvUsage
	var err error

	switch lang {
	case LangGo:
		vars, err = a.auditGoFile(path)
	case LangJS:
		vars, err = a.auditJSFile(path)
	case LangPython:
		vars, err = a.auditPythonFile(path)
	case LangShell:
		vars, err = a.auditShellFile(path)
	}

	if err != nil {
		return
	}

	a.mu.Lock()
	defer a.mu.Unlock()
	a.results = append(a.results, vars...)
}

func (a *Auditor) auditGoFile(path string) ([]EnvUsage, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var vars []EnvUsage

	ast.Inspect(node, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}

		ident, ok := sel.X.(*ast.Ident)
		if !ok {
			return true
		}

		if ident.Name == "os" && (sel.Sel.Name == "Getenv" || sel.Sel.Name == "LookupEnv") {
			if len(call.Args) > 0 {
				if lit, ok := call.Args[0].(*ast.BasicLit); ok && lit.Kind == token.STRING {
					key := strings.Trim(lit.Value, `"`)
					pos := fset.Position(call.Pos())
					vars = append(vars, EnvUsage{
						Key:      key,
						File:     path,
						Lines:    []int{pos.Line},
						Language: LangGo,
					})
				}
			}
		}

		if ident.Name == "viper" && (sel.Sel.Name == "Get" || sel.Sel.Name == "Set") {
			if len(call.Args) > 0 {
				if lit, ok := call.Args[0].(*ast.BasicLit); ok && lit.Kind == token.STRING {
					key := strings.Trim(lit.Value, `"`)
					pos := fset.Position(call.Pos())
					vars = append(vars, EnvUsage{
						Key:      key,
						File:     path,
						Lines:    []int{pos.Line},
						Language: LangGo,
					})
				}
			}
		}

		return true
	})

	return vars, nil
}

var jsProcessEnvPattern = regexp.MustCompile(`process\.env\.([A-Z_][A-Z0-9_]*)`)
var jsDotEnvPattern = regexp.MustCompile(`\.env\.([A-Z_][A-Z0-9_]*)`)
var jsProcessEnvBracketPattern = regexp.MustCompile(`process\.env\[["']([A-Z_][A-Z0-9_]*)["']\]`)

func (a *Auditor) auditJSFile(path string) ([]EnvUsage, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	content := string(data)
	lines := strings.Split(content, "\n")

	var vars []EnvUsage

	for i, line := range lines {
		matches := jsProcessEnvPattern.FindAllStringSubmatch(line, -1)
		for _, match := range matches {
			if len(match) > 1 && match[1] != "" {
				vars = append(vars, EnvUsage{
					Key:      match[1],
					File:     path,
					Lines:    []int{i + 1},
					Language: LangJS,
				})
			}
		}

		matches = jsDotEnvPattern.FindAllStringSubmatch(line, -1)
		for _, match := range matches {
			if len(match) > 1 && match[1] != "" {
				vars = append(vars, EnvUsage{
					Key:      match[1],
					File:     path,
					Lines:    []int{i + 1},
					Language: LangJS,
				})
			}
		}

		matches = jsProcessEnvBracketPattern.FindAllStringSubmatch(line, -1)
		for _, match := range matches {
			if len(match) > 1 && match[1] != "" {
				vars = append(vars, EnvUsage{
					Key:      match[1],
					File:     path,
					Lines:    []int{i + 1},
					Language: LangJS,
				})
			}
		}
	}

	return vars, nil
}

var pythonOsEnvironPattern = regexp.MustCompile(`os\.environ(?:ment)?\[["']([A-Z_][A-Z0-9_]*)["']\]`)
var pythonOsGetenvPattern = regexp.MustCompile(`os\.getenv\(["']([A-Z_][A-Z0-9_]*)["']`)
var pythonOsEnvironGetPattern = regexp.MustCompile(`os\.environ(?:ment)?\.get\(["']([A-Z_][A-Z0-9_]*)["']`)

func (a *Auditor) auditPythonFile(path string) ([]EnvUsage, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	content := string(data)
	lines := strings.Split(content, "\n")

	var vars []EnvUsage

	for i, line := range lines {
		matches := pythonOsEnvironPattern.FindAllStringSubmatch(line, -1)
		for _, match := range matches {
			if len(match) > 1 {
				vars = append(vars, EnvUsage{
					Key:      match[1],
					File:     path,
					Lines:    []int{i + 1},
					Language: LangPython,
				})
			}
		}

		matches = pythonOsGetenvPattern.FindAllStringSubmatch(line, -1)
		for _, match := range matches {
			if len(match) > 1 {
				vars = append(vars, EnvUsage{
					Key:      match[1],
					File:     path,
					Lines:    []int{i + 1},
					Language: LangPython,
				})
			}
		}

		matches = pythonOsEnvironGetPattern.FindAllStringSubmatch(line, -1)
		for _, match := range matches {
			if len(match) > 1 {
				vars = append(vars, EnvUsage{
					Key:      match[1],
					File:     path,
					Lines:    []int{i + 1},
					Language: LangPython,
				})
			}
		}
	}

	return vars, nil
}

var shellVarPattern = regexp.MustCompile(`\$([A-Z_][A-Z0-9_]*)|\$\{([A-Z_][A-Z0-9_]*)\}`)

func (a *Auditor) auditShellFile(path string) ([]EnvUsage, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	content := string(data)
	lines := strings.Split(content, "\n")

	var vars []EnvUsage

	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "#") {
			continue
		}

		matches := shellVarPattern.FindAllStringSubmatch(line, -1)
		for _, match := range matches {
			key := match[1]
			if key == "" {
				key = match[2]
			}
			if key != "" {
				vars = append(vars, EnvUsage{
					Key:      key,
					File:     path,
					Lines:    []int{i + 1},
					Language: LangShell,
				})
			}
		}
	}

	return vars, nil
}

func (a *Auditor) buildResult() (*AuditResult, error) {
	result := &AuditResult{
		UsedNotDeclared: []EnvUsage{},
		DeclaredNotUsed: []string{},
		DeclaredAndUsed: []string{},
	}

	used := make(map[string]bool)
	for _, r := range a.results {
		used[r.Key] = true
	}

	for key := range a.declared {
		if !used[key] {
			result.DeclaredNotUsed = append(result.DeclaredNotUsed, key)
		} else {
			result.DeclaredAndUsed = append(result.DeclaredAndUsed, key)
		}
	}

	keyToUsage := make(map[string]EnvUsage)
	for _, r := range a.results {
		if !a.declared[r.Key] {
			mapKey := r.Key + "|" + r.File
			if existing, ok := keyToUsage[mapKey]; ok {
				seen := make(map[int]bool)
				for _, line := range existing.Lines {
					seen[line] = true
				}
				for _, line := range r.Lines {
					if !seen[line] {
						existing.Lines = append(existing.Lines, line)
						seen[line] = true
					}
				}
				keyToUsage[mapKey] = existing
			} else {
				keyToUsage[mapKey] = r
			}
		}
	}

	for _, r := range keyToUsage {
		result.UsedNotDeclared = append(result.UsedNotDeclared, r)
	}

	return result, nil
}

func AuditDir(rootDir string, envFile string, languages []Language, exclude []string, verbose bool) (*AuditResult, error) {
	auditor := New(rootDir)
	if envFile != "" {
		auditor.SetEnvFile(envFile)
	}
	if len(languages) > 0 {
		auditor.SetLanguages(languages)
	}
	if len(exclude) > 0 {
		auditor.SetExclude(exclude)
	}

	return auditor.Run()
}
