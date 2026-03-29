# envforge

[![Build Status](https://img.shields.io/github/actions/workflow/status/Santiago1809/envforge/release.yml)](https://github.com/Santiago1809/envforge/actions)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://golang.org/dl/)
[![Platform Support](https://img.shields.io/badge/platform-linux%20%7C%20macOS%20%7C%20Windows-blue)](https://github.com/Santiago1809/envforge/releases)

A smart environment variable manager for developers. Envforge helps you compare, sync, audit, encrypt, and watch your `.env` files with zero configuration.

## The Problem

Managing environment variables across projects is chaotic:
- `.env` files drift from `.env.example`
- Hard to know which env vars are actually used in code
- Sharing secrets with teammates is risky
- No way to validate env vars before running your app

Envforge solves all of this with a single CLI.

## Quick Demo

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

## Installation

### Option 1 — Go Install (all platforms)

If you have Go installed:
```bash
go install github.com/Santiago1809/envforge/cmd/envforge@latest
```

The binary will be placed in `$GOPATH/bin` (usually `~/go/bin`).
Make sure that directory is in your PATH (see platform instructions below).

---

### Option 2 — Download Binary

Download the latest release for your platform from
[GitHub Releases](https://github.com/Santiago1809/envforge/releases):

| Platform | File |
|---|---|
| Windows (64-bit) | `envforge_windows_amd64.zip` |
| macOS (Apple Silicon) | `envforge_darwin_arm64.tar.gz` |
| macOS (Intel) | `envforge_darwin_amd64.tar.gz` |
| Linux (64-bit) | `envforge_linux_amd64.tar.gz` |
| Linux (ARM) | `envforge_linux_arm64.tar.gz` |

---

### Windows Setup

1. Download and extract `envforge_windows_amd64.zip`
2. Move `envforge.exe` to a permanent folder, for example:
   `C:\Users\YOUR_USER\tools\envforge\`
3. Add that folder to your PATH:
   - Press `Win + S` and search for **"environment variables"**
   - Click **"Edit the system environment variables"**
   - Click **"Environment Variables..."**
   - Under **"User variables"**, select **Path** and click **Edit**
   - Click **New** and add: `C:\Users\YOUR_USER\tools\envforge`
   - Click OK on all dialogs
4. Open a **new** terminal and verify:
```bash
   envforge --version
```

---

### macOS Setup

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
   - Run `envforge --version` again and click **Open** on the dialog
4. Verify:
```bash
   envforge --version
```

---

### Linux Setup
```bash
tar -xzf envforge_linux_amd64.tar.gz
mv envforge /usr/local/bin/envforge
chmod +x /usr/local/bin/envforge
envforge --version
```

---

## Commands

### info

Print information about a `.env` file.

```bash
$ envforge info .env
File: .env
Keys: 10
Size: 287 bytes
Modified: 2026-03-28 14:30:00

Keys:
  DATABASE_URL = postgres://localhost:5432/mydb
  DB_HOST = localhost
  DB_PORT = 5432
  DB_NAME = myapp
  DB_PASSWORD = ********
  APP_PORT = 8080
  APP_ENV = development
  JWT_SECRET = ********
  DEBUG_MODE = true
  MY_LOCAL_VAR = local_value
```

### diff

Compare two `.env` files.

```bash
$ envforge diff .env .env.example
MISSING in .env (1):
  + API_KEY

EXTRA in .env (1):
  - MY_LOCAL_VAR
```

### sync

Sync keys from `.env` to `.env.example` (strips values).

```bash
$ envforge sync
Sync .env -> .env.example
This will add missing keys from .env to .env.example (values will be stripped).
Continue? [y/N]: y
Successfully synced to .env.example
```

### check

Validate required environment variables are set.

```bash
$ envforge check --required DATABASE_URL,DB_HOST,API_KEY
Missing required environment variables:
  - API_KEY
exit status 1

$ envforge check --from .env.example
All required environment variables are set
```

### audit

Scan source code for environment variable usage.

```bash
$ envforge audit ./src --env-file .env.example

USED but NOT DECLARED (2):
  + DATABASE_URL (src/db/connection.go:15)
  + JWT_SECRET (src/auth/middleware.go:8)

DECLARED but NOT USED (1):
  - DEBUG_MODE
```

Supported languages: Go, JavaScript/TypeScript, Python, Shell.

### encrypt

Encrypt a `.env` file for safe sharing.

```bash
$ envforge encrypt .env --key "your-secure-passphrase"
Encrypted: .env -> .env.enc.b64
```

### decrypt

Decrypt an encrypted file.

```bash
$ envforge decrypt .env.enc.b64 --key "your-secure-passphrase"
# Outputs decrypted content to stdout

$ envforge decrypt .env.enc.b64 --key "your-secure-passphrase" --out .env.decrypted
Decrypted: .env.enc.b64 -> .env.decrypted
```

### verify

Verify integrity of an encrypted file without decrypting.

```bash
$ envforge verify .env.enc.b64 --key "your-secure-passphrase"
Integrity OK
```

### watch

Watch a `.env` file for changes and re-check automatically.

```bash
$ envforge watch .env --exec "make restart"
Watching .env for changes... (Ctrl+C to stop)
```

### keygen

Generate a random 32-byte encryption key.

```bash
$ envforge keygen
a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2

Store this key in a password manager!
```

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
# Sync new vars to .env.example
$ envforge sync
Sync .env -> .env.example
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
          go-version: '1.22'

      - name: Install envforge
        run: go install github.com/Santiago1809/envforge/cmd/envforge@latest

      - name: Check environment variables
        run: envforge check --from .env.example
```

### Docker Entrypoint

```dockerfile
FROM golang:1.22-alpine AS builder
RUN go install github.com/Santiago1809/envforge/cmd/envforge@latest

FROM alpine:latest
COPY --from=builder /go/bin/envforge /usr/local/bin/envforge
COPY .env.example /app/.env.example

# Validate env vars before starting app
ENTRYPOINT ["envforge", "check", "--from", ".env.example", "&&", "myapp"]
```

## Shell Completions

Generate shell completion scripts for envforge:

```bash
envforge completion bash
envforge completion zsh
envforge completion fish
envforge completion powershell
```

### Bash

```bash
envforge completion bash > /etc/bash_completion.d/envforge
```

### Zsh

```bash
envforge completion zsh > "${fpath[1]}/_envforge"
```

### Fish

```bash
envforge completion fish > ~/.config/fish/completions/envforge.fish
```

### PowerShell

```powershell
envforge completion powershell >> $PROFILE
```

## Contributing

Contributions are welcome! Please open an issue or submit a pull request.

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests: `go test ./...`
5. Submit a pull request

## License

MIT License - see [LICENSE](LICENSE) for details.
