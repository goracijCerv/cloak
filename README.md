# 🧥 Cloak

> **A safe haven for your work-in-progress code.**
> Cloak safely backs up untracked or modified files in your Git repository before you perform dangerous operations like `git reset --hard`, `git clean`, or messy branch switches.

---

## 📌 Table of Contents

* [💡 Why Cloak?](#-why-cloak)
* [✨ Features](#-features)
* [📖 Usage Examples](#-usage-examples)
* [⚙️ Global Configuration](#️-global-configuration)
* [🤖 Automation and CI/CD](#-automation-and-cicd)
* [📝 Logging](#-logging)
* [📄 License](#-license)

---

## 🏷️ Badges

[![Go Version](https://img.shields.io/badge/Go-1.25.0+-00ADD8?logo=go)](https://go.dev/)
[![Zero Dependencies](https://img.shields.io/badge/Dependencies-0-brightgreen)](#-features)
[![CI/CD Ready](https://img.shields.io/badge/CI%2FCD-JSON_Ready-blue)](#-automation-and-cicd)

---

## 💡 Why Cloak?

Have you ever lost hours of work because you forgot to commit a new file before a `git reset --hard`?

Cloak solves this by:

* Detecting **untracked, staged, and modified files**
* Creating a **secure, timestamped backup**
* Allowing **instant restoration**

---

## ✨ Features

* 🛡️ **Fail-Safe Restoration**
  Uses a strict `manifest.json` system to restore files to their exact original paths.

* ⚡ **Zero Bloat**
  Written in pure Go using only the standard library (plus `cobra` for CLI).

* 🤖 **CI/CD Ready**
  Universal `--json` flag for machine-readable output.

* ⚙️ **Global Configuration**
  Auto-generates a `config.json` for custom defaults.

* 🧹 **Smart Retention**
  Powerful cleanup filters:

  * `--before`
  * `--after`
  * `--all`

---



---

## 📖 Usage Examples

Cloak automatically detects the Git repository in your current directory.

---

### 📦 Backup files

```bash
# Basic backup
cloak backup

# Backup with label
cloak backup -m "before_rebase"

# Dry run
cloak backup --dry-run
```

---

### 📂 List backups

```bash
cloak list
```

---

### 🔍 Inspect backup

```bash
cloak info /path/to/backup
```

---

### ♻️ Restore backup

```bash
cloak restore --back /path/to/backup
```

⚠️ You will be prompted before overwriting files.

---

### 🧹 Cleanup backups

```bash
# Delete all
cloak delete --all

# Delete by date
cloak delete --before 2026-01-01

# Force delete
cloak delete --all --yes
```

---

## ⚙️ Global Configuration

On first run, Cloak creates:

* **Linux/macOS:** `~/.config/cloak/config.json`
* **Windows:** `%AppData%\\cloak\\config.json`

Example:

```json
{
  "default_output_dir": "/Users/dev/External_Drive/Cloak_Backups",
  "always_skip_confirm": false,
  "default_json_output": false
}
```

---

## 🤖 Automation and CI/CD

All commands support `--json`:

```bash
cloak backup --json
```

Example output:

```json
{
  "status": "success",
  "message": "Backup completed successfully",
  "data": {
    "FilesBackedUp": [
      "/home/user/dev/my_api/main.go",
      "/home/user/dev/my_api/.env.local"
    ],
    "OutputDirectory": "/home/user/dev/backup/[my_api]-2026-04-01_16-00-00",
    "TotalFiles": 2
  }
}
```

⚠️ Use `--yes` with destructive commands to avoid blocking:

```bash
cloak delete --all --yes --json
```

---

## 📝 Logging

```bash
# Log path
cloak logs --path

# Last lines
cloak logs --tail 20

# Clear logs
cloak logs --clear
```

* Auto-rotated logs (max 5MB)
* Tracks internal errors and operations

---

## 📄 License

MIT License
