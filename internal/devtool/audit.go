package devtool

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const outputPlaceholder = "{output}"

type auditSpec struct {
	name       string
	executable string
	arguments  []string
	outputName string
}

type auditFailure struct {
	name     string
	exitCode int
}

// AuditError reports every local scanner that found issues, was blocked, or failed.
type AuditError struct {
	failures []auditFailure
}

func (e *AuditError) Error() string {
	return fmt.Sprintf("local security audit completed with %d non-passing scan(s)", len(e.failures))
}

// AuditLocal runs the approved host-installed scanners and stores ignored raw evidence.
func AuditLocal(
	ctx context.Context,
	repository string,
	containerImage string,
	stdout io.Writer,
	stderr io.Writer,
) error {
	repo, err := filepath.Abs(repository)
	if err != nil {
		return fmt.Errorf("resolve repository: %w", err)
	}
	outputDir := filepath.Join(repo, ".ai", "audits", "p0")
	if err := os.MkdirAll(outputDir, 0o700); err != nil {
		return fmt.Errorf("create audit output directory: %w", err)
	}

	var failures []auditFailure
	for _, spec := range auditSpecs(containerImage) {
		exitCode, runErr := runAuditSpec(ctx, repo, outputDir, spec)
		status := auditStatus(exitCode)
		_, _ = fmt.Fprintf(stdout, "%s: %s (exit %d) -> %s\n",
			spec.name, status, exitCode, auditEvidencePath(outputDir, spec))
		if runErr != nil {
			_, _ = fmt.Fprintf(stderr, "%s: %v\n", spec.name, runErr)
		}
		if exitCode != 0 {
			failures = append(failures, auditFailure{name: spec.name, exitCode: exitCode})
		}
	}
	if len(failures) > 0 {
		return &AuditError{failures: failures}
	}
	return nil
}

func auditSpecs(containerImage string) []auditSpec {
	return []auditSpec{
		{
			name:       "semgrep",
			executable: "semgrep",
			arguments: []string{
				"scan", "--error", "--config", "auto", "--json", "--json-output", outputPlaceholder,
				"--exclude", ".git", "--exclude", ".ai", "--exclude", ".beads", ".",
			},
			outputName: "semgrep.json",
		},
		{
			name:       "snyk-open-source",
			executable: "snyk",
			arguments: []string{
				"test", "--all-projects", "--severity-threshold=high",
				"--json-file-output=" + outputPlaceholder,
			},
			outputName: "snyk-open-source.json",
		},
		{
			name:       "snyk-code",
			executable: "snyk",
			arguments: []string{
				"code", "test", "--severity-threshold=high",
				"--json-file-output=" + outputPlaceholder,
			},
			outputName: "snyk-code.json",
		},
		{
			name:       "snyk-container",
			executable: "snyk",
			arguments: []string{
				"container", "test", containerImage, "--severity-threshold=high", "--app-vulns",
				"--json-file-output=" + outputPlaceholder,
			},
			outputName: "snyk-container.json",
		},
	}
}

func runAuditSpec(
	ctx context.Context,
	repository string,
	outputDir string,
	spec auditSpec,
) (int, error) {
	outputPath := filepath.Join(outputDir, spec.outputName)
	if err := os.Remove(outputPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return -1, fmt.Errorf("remove stale scanner output: %w", err)
	}
	arguments := make([]string, len(spec.arguments))
	for index, argument := range spec.arguments {
		arguments[index] = strings.ReplaceAll(argument, outputPlaceholder, outputPath)
	}

	logPath := filepath.Join(outputDir, spec.name+".log")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
	if err != nil {
		return -1, fmt.Errorf("open scanner log: %w", err)
	}
	defer logFile.Close()

	//nolint:gosec // Executable and arguments come from a fixed internal scanner table; no shell is used.
	// nosemgrep: go.lang.security.audit.dangerous-exec-command.dangerous-exec-command
	command := exec.CommandContext(ctx, spec.executable, arguments...)
	command.Dir = repository
	command.Env = append(os.Environ(), "SNYK_CI=1")
	command.Stdout = logFile
	command.Stderr = logFile
	err = command.Run()
	if chmodErr := os.Chmod(outputPath, 0o600); chmodErr != nil && !errors.Is(chmodErr, os.ErrNotExist) {
		return -1, fmt.Errorf("secure scanner output: %w", chmodErr)
	}
	if err == nil {
		return 0, nil
	}
	if exitError, ok := errors.AsType[*exec.ExitError](err); ok {
		return exitError.ExitCode(), fmt.Errorf("run %s scanner: %w", spec.name, err)
	}
	return -1, fmt.Errorf("run %s scanner: %w", spec.name, err)
}

func auditEvidencePath(outputDir string, spec auditSpec) string {
	outputPath := filepath.Join(outputDir, spec.outputName)
	if _, err := os.Stat(outputPath); err == nil {
		return outputPath
	}
	return filepath.Join(outputDir, spec.name+".log")
}

func auditStatus(exitCode int) string {
	switch exitCode {
	case 0:
		return "passed"
	case 1:
		return "findings"
	case 77:
		return "blocked-auth"
	default:
		return "failed"
	}
}
