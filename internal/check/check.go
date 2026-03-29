package check

import (
	"fmt"
	"os"
	"strings"

	"github.com/Santiago1809/envoy/internal/parser"
)

type CheckResult struct {
	Valid         bool
	MissingKeys   []string
	EmptyKeys     []string
	MissingCount  int
	EmptyCount    int
	TotalRequired int
}

type Options struct {
	Required   []string
	FromFile   string
	AllowEmpty bool
	Prefix     string
}

func Check(opts *Options) (*CheckResult, error) {
	result := &CheckResult{
		Valid:         true,
		MissingKeys:   []string{},
		EmptyKeys:     []string{},
		MissingCount:  0,
		EmptyCount:    0,
		TotalRequired: len(opts.Required),
	}

	required := opts.Required

	var envFile *parser.EnvFile
	if opts.FromFile != "" {
		var err error
		envFile, err = parser.Load(opts.FromFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load env file: %w", err)
		}
		if len(required) == 0 {
			required = envFile.Keys()
		} else {
			seen := make(map[string]bool)
			for _, k := range required {
				seen[k] = true
			}
			for _, k := range envFile.Keys() {
				if !seen[k] {
					required = append(required, k)
				}
			}
		}
		result.TotalRequired = len(required)
	}

	for _, key := range required {
		if opts.Prefix != "" && !strings.HasPrefix(key, opts.Prefix) {
			continue
		}

		var value string
		var exists bool

		if envFile != nil {
			value, exists = envFile.Get(key)
		} else {
			value, exists = os.LookupEnv(key)
		}

		if !exists {
			result.MissingKeys = append(result.MissingKeys, key)
			result.MissingCount++
			result.Valid = false
			continue
		}

		if !opts.AllowEmpty && value == "" {
			result.EmptyKeys = append(result.EmptyKeys, key)
			result.EmptyCount++
			result.Valid = false
		}
	}

	return result, nil
}

func CheckRequiredKeys(keys []string, allowEmpty bool) (*CheckResult, error) {
	opts := &Options{
		Required:   keys,
		AllowEmpty: allowEmpty,
	}
	return Check(opts)
}

func CheckFromFile(envFile string, allowEmpty bool, prefix string) (*CheckResult, error) {
	opts := &Options{
		FromFile:   envFile,
		AllowEmpty: allowEmpty,
		Prefix:     prefix,
	}
	return Check(opts)
}

func CheckKeys(keys []string) error {
	opts := &Options{
		Required: keys,
	}

	result, err := Check(opts)
	if err != nil {
		return err
	}

	if !result.Valid {
		if len(result.MissingKeys) > 0 {
			fmt.Printf("Error: Missing required environment variables:\n")
			for _, key := range result.MissingKeys {
				fmt.Printf("  - %s\n", key)
			}
		}
		if len(result.EmptyKeys) > 0 {
			fmt.Printf("Error: Required environment variables are empty:\n")
			for _, key := range result.EmptyKeys {
				fmt.Printf("  - %s\n", key)
			}
		}
	}

	return nil
}

func GetEnv(key string) (string, bool) {
	return os.LookupEnv(key)
}

func GetEnvOrDefault(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func GetAllWithPrefix(prefix string) map[string]string {
	result := make(map[string]string)
	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := parts[0]
		value := parts[1]
		if strings.HasPrefix(key, prefix) {
			result[key] = value
		}
	}
	return result
}
