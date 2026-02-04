# Spotter 🏋️‍♂️

**Observability & Logging for Fitia Services**

`spotter` is a lightweight, opinionated wrapper around Go's standard `log/slog`. It ensures consistent log formatting across Fitia's multi-cloud infrastructure (GCP & Azure) while providing a developer-friendly experience for local debugging.

> **Note:** This repository is public to facilitate Go module imports across our microservices. It is strictly tailored to Fitia's internal infrastructure standards. Pull requests from the public may be declined if they deviate from internal requirements.

## Why "Spotter"?

In the gym, a spotter watches your form and has your back when the weight gets too heavy. In our backend, this package watches your "form" (log structure) and helps you debug when the load (traffic) gets heavy.

## Features

* **Cloud Agnostic:** Switch between **Google Cloud Run** and **Azure Container Apps** formats with a single config flag.
* **Standardized Severity:** Automatically maps Go log levels to cloud-specific severity levels (e.g., `severity` for GCP, lowercase `level` for Azure).
* **Source Location:** (GCP) Maps source code location so logs in Cloud Logging are clickable, taking you straight to the line of code.
* **Context Injection:** Easily inject container metadata (Revision, Replica, Service Name) into every log entry.
* **Zero-Dependency Runtime:** Relies only on the standard library.

## Installation

```bash
go get github.com/fitia/spotter

```

## Quick Start

Initialize `spotter` at the start of your application (usually in `main.go`). It configures the global `slog` default, so you can use standard library calls throughout your app.

```go
package main

import (
    "log/slog"
    "os"

    "github.com/fitia/spotter"
)

func main() {
    // 1. Define Environment Metadata
    meta := map[string]string{
        "service": "meal-planner",
        "env":     os.Getenv("APP_ENV"),
        "version": os.Getenv("CONTAINER_APP_REVISION"), // specific to your runtime
    }

    // 2. Configure the Spotter
    cfg := spotter.Config{
        // Provider: spotter.ProviderGCP or spotter.ProviderAzure or spotter.ProviderLocal
        Provider:    spotter.ProviderAzure, 
        Level:       slog.LevelInfo,
        AddSource:   true, // Adds file/line number
        StaticAttrs: meta,
    }

    // 3. Initialize (Returns the logger instance if you need dependency injection)
    log := spotter.New(cfg)

    // 4. Log away!
    // You can use the returned instance...
    log.Info("Service started")
    
    // ...or the global default
    slog.Info("Processing user request", "user_id", 101, "status", "active")
}

```

## Configuration

### Providers

The `Provider` setting determines how keys and values are transformed before output.

| Provider | Enum | Key Behavior | Time Format |
| --- | --- | --- | --- |
| **Google Cloud** | `spotter.ProviderGCP` | `msg` → `message`<br>

<br>`level` → `severity` (UPPERCASE)<br>

<br>`source` → `logging.googleapis...` | RFC3339Nano (UTC) |
| **Azure** | `spotter.ProviderAzure` | `level` → `level` (lowercase)<br>

<br>Standard JSON structure | RFC3339Nano (UTC) |
| **Local** | `spotter.ProviderLocal` | Standard JSON behavior (Defaults to Azure-like structure) | RFC3339Nano (UTC) |

### Static Attributes

Use `StaticAttrs` in the config to inject constant metadata into every log line. This is cleaner than reading `os.Getenv` inside the package and keeps the library side-effect free.

## Output Examples

**Azure / Local:**

```json
{
  "time": "2023-10-27T10:00:00.000Z",
  "level": "info",
  "msg": "user created",
  "service": "meal-planner",
  "version": "v1.2.0"
}

```

**GCP:**

```json
{
  "time": "2023-10-27T10:00:00.000Z",
  "severity": "INFO",
  "message": "user created",
  "logging.googleapis.com/sourceLocation": { "file": "main.go", "line": 42 },
  "service": "meal-planner",
  "version": "v1.2.0"
}

```

## Contributing

1. Clone the repo.
2. Create a feature branch (`git checkout -b feature/better-logs`).
3. Commit your changes.
4. Open a Pull Request.

---

Built with ❤️ and 🥗 at [Fitia](https://fitia.app)