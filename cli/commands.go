package cli

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"envoy/internal/auditor"
	"envoy/internal/check"
	"envoy/internal/crypto"
	"envoy/internal/differ"
	"envoy/internal/parser"
	"envoy/internal/watcher"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func joinInts(nums []int) string {
	strs := make([]string, len(nums))
	for i, n := range nums {
		strs[i] = strconv.Itoa(n)
	}
	return strings.Join(strs, ",")
}

var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
	noColor bool
)

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "disable colored output")
	rootCmd.PersistentFlags().StringP("config", "c", "", "config file (default: ~/.config/envoy/config.yaml)")

	rootCmd.AddCommand(diffCmd)
	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(auditCmd)
	rootCmd.AddCommand(checkCmd)
	rootCmd.AddCommand(encryptCmd)
	rootCmd.AddCommand(decryptCmd)
	rootCmd.AddCommand(keygenCmd)
	rootCmd.AddCommand(verifyCmd)
	rootCmd.AddCommand(watchCmd)
	rootCmd.AddCommand(infoCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(completionCmd)

	viper.BindPFlag("no-color", rootCmd.PersistentFlags().Lookup("no-color"))
}

func initConfig() {
	if noColor {
		os.Setenv("NO_COLOR", "1")
	}

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("$HOME/.config/envoy")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	_ = viper.ReadInConfig()
}

var rootCmd = &cobra.Command{
	Use:   "envoy",
	Short: "Smart Environment Variable Manager",
	Long: `envoy is a developer CLI tool for managing .env files.
It helps you compare, sync, audit, encrypt, and watch your environment variables.`,
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("envoy version %s\n", version)
		fmt.Printf("  commit: %s\n", commit)
		fmt.Printf("  date: %s\n", date)
	},
}

var diffCmd = &cobra.Command{
	Use:   "diff [file1] [file2]",
	Short: "Compare two .env files",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		file1 := args[0]
		file2 := ".env"
		if len(args) > 1 {
			file2 = args[1]
		}

		format, _ := cmd.Flags().GetString("format")
		showValues, _ := cmd.Flags().GetBool("show-values")
		verbose, _ := cmd.Flags().GetBool("verbose")

		hasDiffs, err := differ.DiffFiles(file1, file2, differ.OutputFormat(format), showValues, verbose)
		if err != nil {
			return err
		}

		if hasDiffs {
			os.Exit(1)
		}
		return nil
	},
}

func init() {
	diffCmd.Flags().String("format", "table", "output format: table, json, github")
	diffCmd.Flags().Bool("show-values", false, "show values in diff output (use with caution)")
	diffCmd.Flags().BoolP("verbose", "v", false, "show matching keys as well")
}

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync keys from .env to .env.example",
	RunE: func(cmd *cobra.Command, args []string) error {
		source := ".env"
		if len(args) > 0 {
			source = args[0]
		}
		yes, _ := cmd.Flags().GetBool("yes")
		return differ.SyncToExample(source, yes)
	},
}

func init() {
	syncCmd.Flags().BoolP("yes", "y", false, "skip confirmation prompt")
}

var auditCmd = &cobra.Command{
	Use:   "audit [dir]",
	Short: "Scan source code for environment variable usage",
	Args:  cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := "."
		if len(args) > 0 {
			dir = args[0]
		}

		envFile, _ := cmd.Flags().GetString("env-file")
		langs, _ := cmd.Flags().GetStringSlice("lang")
		exclude, _ := cmd.Flags().GetStringSlice("exclude")
		verbose, _ := cmd.Flags().GetBool("verbose")

		var languages []auditor.Language
		for _, l := range langs {
			languages = append(languages, auditor.Language(l))
		}
		if len(languages) == 0 {
			languages = []auditor.Language{auditor.LangAll}
		}

		result, err := auditor.AuditDir(dir, envFile, languages, exclude, verbose)
		if err != nil {
			return err
		}

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

		if verbose && len(result.DeclaredAndUsed) > 0 {
			fmt.Printf("DECLARED and USED (%d):\n", len(result.DeclaredAndUsed))
			for _, k := range result.DeclaredAndUsed {
				fmt.Printf("  = %s\n", k)
			}
		}

		return nil
	},
}

func init() {
	auditCmd.Flags().String("env-file", "", "path to .env.example file")
	auditCmd.Flags().StringSlice("lang", []string{}, "languages to scan: go, js, py, sh (comma-separated)")
	auditCmd.Flags().StringSlice("exclude", []string{}, "additional directories to exclude (appends to default: testdata, vendor, node_modules, .git, dist, build, bin)")
	auditCmd.Flags().BoolP("verbose", "v", false, "show declared and used variables")
}

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Validate required environment variables",
	RunE: func(cmd *cobra.Command, args []string) error {
		required, _ := cmd.Flags().GetStringSlice("required")
		fromFile, _ := cmd.Flags().GetString("from")
		allowEmpty := cmd.Flags().Changed("allow-empty")
		prefix, _ := cmd.Flags().GetString("prefix")

		opts := &check.Options{
			Required:   required,
			FromFile:   fromFile,
			AllowEmpty: allowEmpty,
			Prefix:     prefix,
		}

		result, err := check.Check(opts)
		if err != nil {
			return err
		}

		if !result.Valid {
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
		}

		fmt.Println("All required environment variables are set")
		return nil
	},
}

func init() {
	checkCmd.Flags().StringSlice("required", []string{}, "comma-separated list of required keys")
	checkCmd.Flags().String("from", "", "use keys from .env.example file")
	checkCmd.Flags().Bool("allow-empty", false, "allow empty values")
	checkCmd.Flags().String("prefix", "", "filter by key prefix (e.g. AWS_)")
}

var encryptCmd = &cobra.Command{
	Use:   "encrypt [file]",
	Short: "Encrypt a .env file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		inputFile := args[0]
		key, _ := cmd.Flags().GetString("key")

		if key == "" {
			return fmt.Errorf("please provide encryption key (--key)")
		}

		outputFile := inputFile + ".enc.b64"
		err := crypto.EncryptFile(inputFile, outputFile, key)
		if err != nil {
			return err
		}

		fmt.Printf("Encrypted: %s -> %s\n", inputFile, outputFile)
		return nil
	},
}

func init() {
	encryptCmd.Flags().StringP("key", "k", "", "encryption passphrase or key file")
}

var decryptCmd = &cobra.Command{
	Use:   "decrypt [file]",
	Short: "Decrypt an encrypted .env file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		inputFile := args[0]
		key, _ := cmd.Flags().GetString("key")
		outputFile, _ := cmd.Flags().GetString("out")

		if key == "" {
			return fmt.Errorf("please provide decryption key (--key)")
		}

		if outputFile == "" {
			data, err := os.ReadFile(inputFile)
			if err != nil {
				return err
			}
			decrypted, err := crypto.Decrypt(data, key)
			if err != nil {
				return err
			}
			fmt.Print(string(decrypted))
		} else {
			err := crypto.DecryptFile(inputFile, outputFile, key)
			if err != nil {
				return err
			}
			fmt.Printf("Decrypted: %s -> %s\n", inputFile, outputFile)
		}
		return nil
	},
}

func init() {
	decryptCmd.Flags().StringP("key", "k", "", "decryption passphrase or key file")
	decryptCmd.Flags().StringP("out", "o", "", "output file (default: stdout)")
}

var keygenCmd = &cobra.Command{
	Use:   "keygen",
	Short: "Generate a random 32-byte encryption key",
	RunE: func(cmd *cobra.Command, args []string) error {
		key, err := crypto.GenerateKey()
		if err != nil {
			return err
		}
		fmt.Println(key)
		fmt.Println("\nStore this key in a password manager!")
		return nil
	},
}

var verifyCmd = &cobra.Command{
	Use:   "verify [file]",
	Short: "Verify integrity of an encrypted .env file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		inputFile := args[0]
		key, _ := cmd.Flags().GetString("key")

		if key == "" {
			return fmt.Errorf("please provide decryption key (--key)")
		}

		data, err := os.ReadFile(inputFile)
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}

		err = crypto.Verify(data, key)
		if err != nil {
			fmt.Printf("Integrity check failed: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Integrity OK")
		return nil
	},
}

func init() {
	verifyCmd.Flags().StringP("key", "k", "", "decryption passphrase or key file")
}

var watchCmd = &cobra.Command{
	Use:   "watch [file]",
	Short: "Watch a .env file for changes",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		file := args[0]
		execCmd, _ := cmd.Flags().GetString("exec")
		debounce, _ := cmd.Flags().GetInt("debounce")

		w, err := watcher.New()
		if err != nil {
			return err
		}
		defer w.Stop()

		w.SetDebounce(time.Duration(debounce) * time.Millisecond)

		w.OnChange(func(e watcher.Event) {
			fmt.Printf("\n[Change detected] %s\n", e.Path)

			env, err := parser.Load(file)
			if err != nil {
				fmt.Printf("Error loading file: %v\n", err)
				return
			}

			fmt.Println("Current keys:")
			for _, k := range env.Keys() {
				v, _ := env.Get(k)
				fmt.Printf("  %s=%s\n", k, v)
			}

			if execCmd != "" {
				fmt.Printf("\nExecuting: %s\n", execCmd)
			}
		})

		err = w.Add(file)
		if err != nil {
			return err
		}

		fmt.Printf("Watching %s for changes... (Ctrl+C to stop)\n", file)

		sigChan := make(chan os.Signal, 1)
		<-sigChan

		return nil
	},
}

func init() {
	watchCmd.Flags().String("exec", "", "command to execute on change")
	watchCmd.Flags().Int("debounce", 50, "debounce time in milliseconds")
}

var infoCmd = &cobra.Command{
	Use:   "info [file]",
	Short: "Print information about a .env file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		file := args[0]

		env, err := parser.Load(file)
		if err != nil {
			return err
		}

		info, _ := os.Stat(file)

		fmt.Printf("File: %s\n", file)
		fmt.Printf("Keys: %d\n", len(env.Keys()))
		fmt.Printf("Size: %d bytes\n", info.Size())
		fmt.Printf("Modified: %s\n", info.ModTime().Format("2006-01-02 15:04:05"))
		fmt.Println("\nKeys:")
		for _, k := range env.Keys() {
			v, _ := env.Get(k)
			if v == "" {
				fmt.Printf("  %s (empty)\n", k)
			} else if len(v) > 50 {
				fmt.Printf("  %s = %s...\n", k, v[:50])
			} else {
				fmt.Printf("  %s = %s\n", k, v)
			}
		}

		return nil
	},
}

var completionCmd = &cobra.Command{
	Use:   "completion [shell]",
	Short: "Generate shell completion scripts",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		shell := args[0]
		switch shell {
		case "bash":
			rootCmd.GenBashCompletion(os.Stdout)
		case "zsh":
			rootCmd.GenZshCompletion(os.Stdout)
		case "fish":
			rootCmd.GenFishCompletion(os.Stdout, true)
		default:
			return fmt.Errorf("unsupported shell: %s (bash, zsh, fish)", shell)
		}
		return nil
	},
}
