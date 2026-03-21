# Spec: Scrapedoctl Advanced Logging System

**Date:** 2026-03-21  
**Status:** Approved  
**Author:** Gemini CLI Agent  
**Version:** 0.1.0  

---

## 1. Goal
Implement a production-grade logging system for `scrapedoctl` that supports structured JSON output, log rotation, retention policies, and detailed request tracing.

## 2. Architecture & Components

### 2.1 Configuration (`internal/config`)
Add a `LoggingConfig` struct to manage the following parameters in `conf.toml`:
- `level`: `debug`, `info`, `warn`, `error` (Default: `info`).
- `format`: `json`, `text` (Default: `json`).
- `path`: Absolute path to the log file (Default: `/var/log/scrapedoctl/scrapedoctl.log`).
- `max_size`: Size in megabytes before the log file is rotated (Default: `10`).
- `max_age`: Maximum number of days to retain old log files (Default: `7`).
- `max_backups`: Maximum number of old log files to retain (Default: `5`).
- `compress`: Boolean flag to determine if rotated logs should be gzipped (Default: `true`).

### 2.2 Logger Implementation (`internal/logger`)
- **Backend**: Use `log/slog` for structured logging.
- **Rotation**: Integrate `github.com/natefinch/lumberjack` as the `io.Writer` for `slog`.
- **Initialization**: Provide an `Init(cfg LoggingConfig)` function that sets the global logger using `slog.SetDefault()`.
- **Resilience**: If the target log file/directory is not writable, fall back to `os.Stderr` and print a clear warning to the console.

### 2.3 Scrape Client Integration (`pkg/scrapedo`)
- **Debug Mode**: Log full request details (Method, URL, Headers) before execution.
- **Info Mode**: Log Scrape.do response metadata (remaining credits, status, cost).
- **Error Mode**: Log full error context for failed requests.

## 3. Detailed Requirements

### 3.1 Permissions
- The `install` command should attempt to create `/var/log/scrapedoctl/`. 
- If it fails, it must provide instructions to the user on how to fix permissions (e.g., `sudo chown`).

### 3.2 Machine Readability
- Default output format MUST be JSON to allow seamless integration with log management systems.

### 3.3 Security
- Ensure that the API Token is never logged. Specifically, if logging headers in `debug` mode, the `token` query parameter and any sensitive headers must be masked.

## 4. Testing Strategy
1.  **Unit Tests**: Verify that `internal/logger` correctly initializes different handlers (JSON/Text) based on config.
2.  **Rotation Tests**: Verify that `lumberjack` parameters are correctly passed.
3.  **Integration Tests**: Verify that `pkg/scrapedo` uses the global logger correctly and honors the log level.
4.  **100% Coverage**: All new logging logic must be fully tested.
