package devtool

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	outputPlaceholder = "{output}"
	semgrepName       = "semgrep"
	snykName          = "snyk"
	semgrepVersion    = "1.175.0"
	snykVersion       = "1.1307.0"
)

var errScannerVersionMismatch = errors.New("scanner version mismatch")

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

type scannerVersion struct {
	Name       string `json:"name"`
	Expected   string `json:"expected"`
	Executable string `json:"executable"`
	Actual     string `json:"actual"`
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
	if mkdirErr := os.MkdirAll(outputDir, 0o700); mkdirErr != nil {
		return fmt.Errorf("create audit output directory: %w", mkdirErr)
	}
	if versionErr := verifyScannerVersions(ctx, repo, outputDir); versionErr != nil {
		return versionErr
	}
	productionGoFiles, err := countProductionGoFiles(repo)
	if err != nil {
		return fmt.Errorf("measure audit source corpus: %w", err)
	}

	var failures []auditFailure
	for _, spec := range auditSpecs(containerImage) {
		exitCode, runErr := runAuditSpec(ctx, repo, outputDir, spec)
		status := auditStatus(exitCode)
		corpus := ""
		if exitCode == 0 {
			corpus, err = validateAuditEvidence(repo, outputDir, containerImage, spec, productionGoFiles)
			if err != nil {
				status = "invalid-evidence"
				exitCode = -1
				runErr = errors.Join(runErr, err)
			}
		}
		_, _ = fmt.Fprintf(stdout, "%s: %s (exit %d, corpus %s) -> %s\n",
			spec.name, status, exitCode, corpus, auditEvidencePath(outputDir, spec))
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
			name:       semgrepName,
			executable: semgrepName,
			arguments: []string{
				"scan", "--error", "--config", "auto", "--json", "--json-output", outputPlaceholder,
				"--exclude", ".git", "--exclude", ".ai", "--exclude", ".beads",
				"--exclude", "vendor", ".",
			},
			outputName: "semgrep.json",
		},
		{
			name:       "snyk-open-source",
			executable: snykName,
			arguments: []string{
				"test", "--all-projects", "--severity-threshold=high",
				"--json-file-output=" + outputPlaceholder,
			},
			outputName: "snyk-open-source.json",
		},
		{
			name:       "snyk-code",
			executable: snykName,
			arguments: []string{
				"code", "test", "--severity-threshold=high",
				"--json-file-output=" + outputPlaceholder,
			},
			outputName: "snyk-code.json",
		},
		{
			name:       "snyk-container",
			executable: snykName,
			arguments: []string{
				"container", "test", containerImage, "--severity-threshold=high", "--app-vulns",
				"--json-file-output=" + outputPlaceholder,
			},
			outputName: "snyk-container.json",
		},
	}
}

func verifyScannerVersions(ctx context.Context, repository, outputDir string) error {
	versions := []scannerVersion{
		{Name: semgrepName, Expected: semgrepVersion, Executable: semgrepName},
		{Name: snykName, Expected: snykVersion, Executable: snykName},
	}
	var failures []error
	for index := range versions {
		//nolint:gosec // Executable and arguments come from the fixed scanner policy.
		// nosemgrep: go.lang.security.audit.dangerous-exec-command.dangerous-exec-command
		command := exec.CommandContext(ctx, versions[index].Executable, "--version")
		command.Dir = repository
		var output bytes.Buffer
		command.Stdout = &output
		command.Stderr = &output
		if err := command.Run(); err != nil {
			failures = append(failures, fmt.Errorf("read %s version: %w", versions[index].Name, err))
		}
		versions[index].Actual = strings.TrimSpace(output.String())
		if versions[index].Actual != versions[index].Expected {
			failures = append(failures, fmt.Errorf(
				"%w: %s expected %s, got %s",
				errScannerVersionMismatch,
				versions[index].Name,
				versions[index].Expected,
				versions[index].Actual,
			))
		}
	}
	evidence, err := json.MarshalIndent(versions, "", "  ")
	if err != nil {
		return fmt.Errorf("encode scanner versions: %w", err)
	}
	if err := os.WriteFile(filepath.Join(outputDir, "versions.json"), evidence, 0o600); err != nil {
		return fmt.Errorf("write scanner versions: %w", err)
	}
	return errors.Join(failures...)
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
