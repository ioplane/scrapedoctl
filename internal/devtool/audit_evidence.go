package devtool

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

var (
	errEmptyAuditCorpus   = errors.New("scanner evidence contains no input corpus")
	errInvalidAuditOutput = errors.New("scanner evidence is invalid")
)

type semgrepEvidence struct {
	Version string `json:"version"`
	Paths   struct {
		Scanned []string `json:"scanned"`
	} `json:"paths"`
	Errors []json.RawMessage `json:"errors"`
}

type snykEvidence struct {
	ProjectName     string `json:"projectName"`
	DependencyCount int    `json:"dependencyCount"`
	PackageManager  string `json:"packageManager"`
	Path            string `json:"path"`
}

func validateAuditEvidence(
	repository string,
	outputDir string,
	containerImage string,
	spec auditSpec,
	productionGoFiles int,
) (string, error) {
	switch spec.name {
	case semgrepName:
		return validateSemgrepEvidence(filepath.Join(outputDir, spec.outputName))
	case "snyk-open-source":
		return validateSnykOpenSourceEvidence(filepath.Join(outputDir, spec.outputName))
	case "snyk-code":
		return validateSnykCodeEvidence(
			repository,
			filepath.Join(outputDir, spec.name+".log"),
			productionGoFiles,
		)
	case "snyk-container":
		return validateSnykContainerEvidence(filepath.Join(outputDir, spec.outputName), containerImage)
	default:
		return "", fmt.Errorf("%w: unknown scanner %q", errInvalidAuditOutput, spec.name)
	}
}

func validateSemgrepEvidence(path string) (string, error) {
	var evidence semgrepEvidence
	if err := readJSON(path, &evidence); err != nil {
		return "", err
	}
	if evidence.Version != semgrepVersion || len(evidence.Errors) != 0 {
		return "", fmt.Errorf("%w: Semgrep version or scan errors", errInvalidAuditOutput)
	}
	if len(evidence.Paths.Scanned) == 0 {
		return "", errEmptyAuditCorpus
	}
	return fmt.Sprintf("%d paths", len(evidence.Paths.Scanned)), nil
}

func validateSnykOpenSourceEvidence(path string) (string, error) {
	projects, err := readSnykProjects(path)
	if err != nil {
		return "", err
	}
	dependencies := 0
	for _, project := range projects {
		if project.ProjectName == "" {
			return "", fmt.Errorf("%w: Snyk project name is empty", errInvalidAuditOutput)
		}
		dependencies += project.DependencyCount
	}
	if dependencies == 0 {
		return "", errEmptyAuditCorpus
	}
	return fmt.Sprintf("%d dependencies", dependencies), nil
}

func validateSnykCodeEvidence(repository, path string, productionGoFiles int) (string, error) {
	log, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read Snyk Code evidence: %w", err)
	}
	if productionGoFiles == 0 {
		return "", errEmptyAuditCorpus
	}
	if !bytes.Contains(log, []byte("Testing "+repository+" ...")) ||
		!bytes.Contains(log, []byte("Test type:         Static code analysis")) {
		return "", fmt.Errorf("%w: Snyk Code target is not the repository", errInvalidAuditOutput)
	}
	return fmt.Sprintf("%d production Go files", productionGoFiles), nil
}

func validateSnykContainerEvidence(path, containerImage string) (string, error) {
	var evidence snykEvidence
	if err := readJSON(path, &evidence); err != nil {
		return "", err
	}
	if !strings.HasPrefix(evidence.ProjectName, "docker-image|") ||
		evidence.PackageManager == "" ||
		!strings.HasPrefix(evidence.Path, containerImage+"/") {
		return "", fmt.Errorf("%w: Snyk Container target mismatch", errInvalidAuditOutput)
	}
	return containerImage + " (" + evidence.PackageManager + ")", nil
}

func readSnykProjects(path string) ([]snykEvidence, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read Snyk evidence: %w", err)
	}
	data = bytes.TrimSpace(data)
	if len(data) == 0 {
		return nil, errEmptyAuditCorpus
	}
	if data[0] == '[' {
		var projects []snykEvidence
		if err := json.Unmarshal(data, &projects); err != nil {
			return nil, fmt.Errorf("decode Snyk projects: %w", err)
		}
		return projects, nil
	}
	var project snykEvidence
	if err := json.Unmarshal(data, &project); err != nil {
		return nil, fmt.Errorf("decode Snyk project: %w", err)
	}
	return []snykEvidence{project}, nil
}

func readJSON(path string, destination any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read scanner evidence: %w", err)
	}
	if len(bytes.TrimSpace(data)) == 0 {
		return errEmptyAuditCorpus
	}
	if err := json.Unmarshal(data, destination); err != nil {
		return fmt.Errorf("decode scanner evidence: %w", err)
	}
	return nil
}

func countProductionGoFiles(repository string) (int, error) {
	count := 0
	err := filepath.WalkDir(repository, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() && path != repository {
			switch entry.Name() {
			case ".git", ".ai", ".beads", "vendor":
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) == ".go" && !strings.HasSuffix(path, "_test.go") {
			count++
		}
		return nil
	})
	if err != nil {
		return 0, fmt.Errorf("walk production Go corpus: %w", err)
	}
	return count, nil
}
