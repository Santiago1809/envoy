# envforge

[![Build Status](https://img.shields.io/github/actions/workflow/status/Santiago1809/envforge/release.yml)](https://github.com/Santiago1809/envforge/actions)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.26+-00ADD8?style=flat&logo=go)](https://golang.org/dl/)
[![Platform Support](https://img.shields.io/badge/platform-linux%20%7C%20macOS%20%7C%20Windows-blue)](https://github.com/Santiago1809/envforge/releases)

A smart environment variable manager for developers. Envforge helps you compare, sync, audit, encrypt, and watch your `.env` files with zero configuration.

## The Problem

Managing environment variables across projects is chaotic:
- `.env` files drift from `.env.example`
- Hard to know which env vars are actually used in code
- Sharing secrets with teammates is risky
- No way to validate env vars before running your app

Envforge solves all of this with a single CLI.

## Installation

### Option 1 — Go Install (all platforms)

If you have Go installed:

```bash
go install github.com/Santiago1809/envforge/cmd/envforge@latest
```

The binary will be placed in `$GOPATH/bin` (usually `~/go/bin`).
Make sure that directory is in your PATH.

> **Note:** `go install` will show version "dev" — this is expected because the
> version is injected at build time by GoReleaser. To get the correct version
> string, download the binary from GitHub Releases. Also note that the
> `envforge update` command only works correctly when installed via the GitHub
> Releases binary, not via `go install`.

### Option 2 — Download Binary (Recommended)

Download the latest release for your platform from
[GitHub Releases](https://github.com/Santiago1809/envforge/releases):

| Platform | File |
|---|---|
| Windows (64-bit) | `envforge_windows_amd64.zip` |
| macOS (Apple Silicon) | `envforge_darwin_arm64.tar.gz` |
| macOS (Intel) | `envforge_darwin_amd64.tar.gz` |
| Linux (64-bit) | `envforge_linux_amd64.tar.gz` |
| Linux (ARM) | `envforge_linux_arm64.tar.gz` |

#### Platform-Specific Setup

**Windows:**

1. Download and extract `envforge_windows_amd64.zip`
2. Move `envforge.exe` to a permanent folder, e.g. `C:\Users\YOUR_USER\tools\envforge\`
3. Add that folder to your PATH:
   - Press `Win + S` and search for **"environment variables"**
   - Click **"Edit the system environment variables"**
   - Click **"Environment Variables..."**
   - Under **"User variables"**, select **Path** and click **Edit**
   - Click **New** and add: `C:\Users\YOUR_USER\tools\envforge`
   - Click OK on all dialogs
4. Open a **new** terminal and verify:

```powershell
envforge version
```

**macOS:**

1. Download and extract the `.tar.gz` for your architecture:

```bash
# Apple Silicon (M1/M2/M3)
tar -xzf envforge_darwin_arm64.tar.gz

# Intel
tar -xzf envforge_darwin_amd64.tar.gz
```

2. Move the binary to `/usr/local/bin`:

```bash
mv envforge /usr/local/bin/envforge
chmod +x /usr/local/bin/envforge
```

3. On first run, macOS may block the binary (Gatekeeper).
   If you see a security warning:
   - Open **System Settings** → **Privacy & Security**
   - Scroll down and click **"Allow Anyway"** next to envforge
   - Run `envforge version` again and click **Open** on the dialog
4. Verify:

```bash
envforge version
```

**Linux:**

```bash
tar -xzf envforge_linux_amd64.tar.gz
sudo mv envforge /usr/local/bin/envforge
chmod +x /usr/local/bin/envforge
envforge version
```

## Quick Start

```bash
# Check your env vars are set before running
$ envforge check --from .env.example
All required environment variables are set

# See what's different between .env and .env.example
$ envforge diff
MISSING in .env (1):
  + API_KEY

EXTRA in .env (1):
  - MY_LOCAL_VAR

# Audit your code to find used but undeclared vars
$ envforge audit ./src --env-file .env.example
USED but NOT DECLARED (2):
  + DATABASE_URL (src/db.go:15)
  + JWT_SECRET (src/auth.go:8)

DECLARED but NOT USED (1):
  - DEBUG_MODE
```

## Commands

### `diff`

Compare two `.env` files. By default compares `.env` vs `.env.example`.

```bash
# Compare .env and .env.example
envforge diff

# Compare specific files
envforge diff .env.staging .env.production

# Show values (use with caution - may expose secrets)
envforge diff --show-values

# JSON output
envforge diff --format json

# GitHub Actions format
envforge diff --format github
```

**Flags:**

- `--format, -f`: Output format: `table` (default), `json`, `github`
- `--show-values`: Show values in diff output
- `--verbose, -v`: Show matching keys as well

---

### `sync`

Sync keys from a source `.env` file to a target `.env.example` file, stripping values.

**Basic usage:**

```bash
# Default: sync .env → .env.example
envforge sync

# Non-interactive (auto-confirm)
envforge sync --yes
```

**Multi-stage environments:**

```bash
# Sync a specific stage
envforge sync --stage development  # → .env → .env.example
envforge sync --stage staging      # → .env.staging → .env.staging.example
envforge sync --stage production   # → .env.production → .env.production.example

# Explicit source and destination
envforge sync --from .env.production --to .env.production.example

# Combine stage with explicit paths (stage is ignored if --from/--to are set)
envforge sync --stage production --from .env --to .env.dev.example
```

**Flags:**

- `--stage, -s`: Environment stage (`development`, `staging`, `production`). Automatically resolves source/target files.
- `--from, -f`: Source `.env` file (default: `.env` or stage-based)
- `--to, -t`: Target `.env.example` file (default: derived from source by appending `.example` or replacing suffix)
- `--yes, -y`: Skip confirmation prompt

**Examples:**

```bash
# Using stages (recommended for multi-env projects)
envforge sync --stage production
# Syncing .env.production → .env.production.example
#  + NEW_KEY (added)
# Successfully synced 1 new key to .env.production.example

# Explicit files
envforge sync --from .env.staging --to .env.staging.example --yes

# With custom suffix (if you use .env.dev instead of .env.staging)
envforge sync --from .env.dev --to .env.dev.example
```


---

### `audit`

Scan source code for environment variable usage.

```bash
# Audit current directory using .env.example
envforge audit . --env-file .env.example

# Audit specific directory
envforge audit ./src --env-file .env.example

# Show all variables (including declared and used)
envforge audit . --env-file .env.example --verbose

# Scan specific languages
envforge audit . --env-file .env.example --lang go,js,py

# Exclude additional directories
envforge audit . --env-file .env.example --exclude coverage,build
```

**Flags:**

- `--env-file, -e`: Path to `.env.example` file (default: `.env.example`)
- `--lang, -l`: Languages to scan: `go`, `js`, `py`, `sh` (comma-separated)
- `--exclude, -x`: Additional directories to exclude (appends to defaults: `testdata, vendor, node_modules, .git, dist, build, bin, .agents, .claude, .skills, skills`)
- `--verbose, -v`: Show declared and used variables

**Supported languages:** Go, JavaScript/TypeScript, Python, Shell

---

### `check`

Validate required environment variables are set.

```bash
# Check against .env.example
envforge check --from .env.example

# Check specific required keys
envforge check --required DATABASE_URL,API_KEY,JWT_SECRET

# Check with prefix filter
envforge check --from .env.example --prefix AWS_

# Allow empty values
envforge check --from .env.example --allow-empty
```

**Flags:**

- `--required`: Comma-separated list of required keys
- `--from, -f`: Use keys from `.env.example` file
- `--allow-empty`: Allow empty values
- `--prefix`: Filter by key prefix (e.g. `AWS_`)

---

### `encrypt`

Encrypt a `.env` file for safe sharing.

```bash
# Encrypt with passphrase
envforge encrypt .env --key "your-secure-passphrase"

# Encrypt using a key file
envforge encrypt .env --key ~/.ssh/id_rsa

# Encrypt and specify output
envforge encrypt .env --key "pass" --out .env.enc
```

**Flags:**

- `--key, -k`: Encryption passphrase or key file (required)

---

### `decrypt`

Decrypt an encrypted `.env` file.

```bash
# Decrypt to stdout
envforge decrypt .env.enc.b64 --key "your-secure-passphrase"

# Decrypt to file
envforge decrypt .env.enc.b64 --key "your-secure-passphrase" --out .env.decrypted

# Decrypt using key file
envforge decrypt .env.enc.b64 --key ~/.ssh/id_rsa --out .env
```

**Flags:**

- `--key, -k`: Decryption passphrase or key file (required)
- `--out, -o`: Output file (default: stdout)

---

### `verify`

Verify integrity of an encrypted file without decrypting.

```bash
envforge verify .env.enc.b64 --key "your-secure-passphrase"
```

**Flags:**

- `--key, -k`: Decryption passphrase or key file (required)

---

### `watch`

Watch a `.env` file for changes and optionally execute a command.

```bash
# Watch .env file
envforge watch .env

# Watch and execute command on change
envforge watch .env --exec "make restart"

# Custom debounce time (ms)
envforge watch .env --exec "systemctl reload app" --debounce 500
```

**Flags:**

- `--exec`: Command to execute on change
- `--debounce`: Debounce time in milliseconds (default: 50)

---

### `info`

Print information about a `.env` file.

```bash
envforge info .env
envforge info .env.example
```

---

### `keygen`

Generate a random 32-byte encryption key.

```bash
envforge keygen
# Output: a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2

# Store this key in a password manager!
```

---

### `version`

Print version information.

```bash
envforge version
# Output:
# envforge version v1.0.2
#   commit: abc1234
#   date:   2026-03-29T12:00:00Z
```

---

### `update`

Update envforge to the latest release.

```bash
# Interactive update (asks for confirmation)
envforge update

# Skip confirmation
envforge update --yes
```

> **Note:** This command only works when envforge is installed from GitHub Releases
> (i.e., placed in `%LOCALAPPDATA%\envforge\` on Windows or `/usr/local/bin/envforge` on Unix).
> It does not work with `go install` because the version cannot be determined.

**Flags:**

- `--yes, -y`: Skip confirmation prompt

**Windows-specific:** The binary is installed to `%LOCALAPPDATA%\envforge\`. This directory
is always writable without admin privileges.

---

### `tui`

Open interactive Terminal User Interface.

```bash
envforge tui
```

**Features:**

- **Overview Tab:** Shows all keys from `.env` with masked values
  - Red: Missing from `.env.example`
  - Yellow: Extra keys not in `.env.example`
  - Green: Present in both

- **Audit Tab:** Shows code audit results
  - Red: Used but not declared
  - Yellow: Declared but not used

- **Health Tab:** Health check of required variables
  - Green checkmark if set, red X if missing
  - File sizes, key counts, last modified dates

**Navigation:**

- `Tab` / `Shift+Tab`: Switch tabs
- `↑` / `↓` or `j` / `k`: Scroll within tab
- `q` or `Ctrl+C`: Quit

---

### `completion`

Generate shell completion scripts.

```bash
# Bash
envforge completion bash > /etc/bash_completion.d/envforge

# Zsh
envforge completion zsh > "${fpath[1]}/_envforge"

# Fish
envforge completion fish > ~/.config/fish/completions/envforge.fish

# PowerShell
envforge completion powershell | Out-File -Encoding utf8 $env:TEMP\envforge_completion.ps1

. $env:TEMP\envforge_completion.ps1
```

---

## Configuration

Envforge can be configured via a config file or environment variables.

**Config file location:** `~/.config/envforge/config.yaml` (or `~/.config/envoy/config.yaml` for legacy)

**Global flags:**

- `--config, -c`: Config file path
- `--no-color`: Disable colored output

**Example config:**

```yaml
# ~/.config/envforge/config.yaml
audit:
  languages: [go, js, py, sh]
  exclude:
    - testdata
    - vendor
    - node_modules
    - .git
    - dist
    - build
    - bin
    - .agents
    - .claude
    - .skills
    - skills
```

---

## Multi-Stage Environments

Envforge makes it easy to manage multiple environment stages (development, staging, production) using the `--stage` flag.

**File naming convention:**

| Stage      | Source File          | Target File              |
|------------|---------------------|--------------------------|
| development| `.env`              | `.env.example`           |
| staging    | `.env.staging`      | `.env.staging.example`   |
| production | `.env.production`   | `.env.production.example`|

**Usage:**

```bash
# Sync production environment
envforge sync --stage production

# Sync staging environment
envforge sync --stage staging --yes

# Sync development (default, same as plain `envforge sync`)
envforge sync --stage development
```

You can also explicitly specify files with `--from` and `--to`:

```bash
envforge sync --from .env.production --to .env.production.example
envforge sync --from .env.staging --to .env.staging.example --yes
```

This is useful if you use custom file naming conventions.

**Typical workflow with stages:**

```bash
# Before deploying to staging
envforge sync --stage staging
envforge check --from .env.staging.example

# Before deploying to production
envforge sync --stage production
envforge check --from .env.production.example

# Local development
envforge sync --stage development
envforge check --from .env.example
```

---

## Real-World Workflow

### Morning: Start Your Project

```bash
# Check all env vars are configured
$ envforge check --from .env.example
All required environment variables are set

# Start your app
$ make dev
```

### During Development: Add a New Env Var

```bash
# Add the var to your .env
echo "NEW_FEATURE_FLAG=true" >> .env

# Run audit to see if it's used anywhere
$ envforge audit ./src --env-file .env.example
USED but NOT DECLARED (1):
  + NEW_FEATURE_FLAG (src/features/flags.go:5)
```

### Before Committing: Sync .env.example

```bash
# Sync new vars to .env.example (or use --stage for multi-env)
$ envforge sync
# or: envforge sync --stage staging
# or: envforge sync --from .env --to .env.example

Continue? [y/N]: y
Successfully synced to .env.example
```

### Code Review: Run Full Audit

```bash
$ envforge audit ./src --env-file .env.example -v
USED but NOT DECLARED (5):
  + DATABASE_URL (src/db.go:15)
  + API_KEY (src/client.go:10)
  + JWT_SECRET (src/auth.go:8)

DECLARED but NOT USED (2):
  - DEBUG_MODE
  - OLD_FEATURE

DECLARED and USED (8):
  = DB_HOST
  = DB_PORT
  = APP_PORT
```

---

## CI/CD Integration

### GitHub Actions

```yaml
name: Check Environment

on: [push, pull_request]

jobs:
  check-env:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.26'

      - name: Install envforge
        run: go install github.com/Santiago1809/envforge/cmd/envforge@latest

      - name: Check environment variables
        run: envforge check --from .env.example
```

### Docker Entrypoint

```dockerfile
FROM golang:1.26-alpine AS builder
RUN go install github.com/Santiago1809/envforge/cmd/envforge@latest

FROM alpine:latest
COPY --from=builder /go/bin/envforge /usr/local/bin/envforge
COPY .env.example /app/.env.example

# Validate env vars before starting app
ENTRYPOINT ["envforge", "check", "--from", ".env.example", "&&", "myapp"]
```

---

## Troubleshooting

### "dev" version shown when using `go install`

This is expected. The version is injected at build time by GoReleaser via ldflags.
`go install` does not pass these flags. To see the correct version, download from
GitHub Releases.

### `envforge update` doesn't work with `go install`

The update command relies on the binary being in a known location (`%LOCALAPPDATA%\envforge\` on Windows or `/usr/local/bin/envforge` on Unix) that can be overwritten. `go install` places the binary in `$GOPATH/bin` which may not be writable or the update mechanism cannot locate it correctly. Install from GitHub Releases for a proper update experience.

### Windows: "Access is denied" during update

On Windows, running processes cannot be overwritten. The update process:
1. Downloads the new binary to a temporary location
2. Creates a batch script that waits for the current process to exit
3. The batch script runs after envforge exits and replaces the binary

Ensure you're using the Windows release binary (not `go install`) for best compatibility.

### Parser errors with `.env` files

The `.env` file must follow the format `KEY=value` on each line.
Lines without `=` will cause parse errors. Comments should start with `#`.
Values can contain `:` without issues (e.g., URLs).

---

## Contributing

Contributions are welcome! Please open an issue or submit a pull request.

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests: `go test ./...`
5. Submit a pull request

## License

MIT License - see [LICENSE](LICENSE) for details.
