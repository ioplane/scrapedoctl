package devtool

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestAuditCommandsAreFixedAndShellFree(t *testing.T) {
	specs := auditSpecs("localhost/scrapedoctl:p0")
	if len(specs) != 4 {
		t.Fatalf("audit command count = %d, want 4", len(specs))
	}
	wantTools := []string{semgrepName, snykName, snykName, snykName}
	for index, spec := range specs {
		if spec.executable != wantTools[index] {
			t.Fatalf("command %d executable = %q, want %q", index, spec.executable, wantTools[index])
		}
		if slices.Contains([]string{"sh", "bash", "zsh", "python", "python3"}, spec.executable) {
			t.Fatalf("command %d invokes forbidden wrapper %q", index, spec.executable)
		}
	}
	if !slices.Contains(specs[0].arguments, "--error") {
		t.Fatal("semgrep command does not fail the audit when findings exist")
	}
}

func TestAuditEvidencePolicy(t *testing.T) {
	const image = "localhost/scrapedoctl:p0"
	repository := t.TempDir()
	outputDir := t.TempDir()
	fixtures := map[string]string{
		semgrepName:        `{"version":"1.175.0","paths":{"scanned":["main.go"]},"errors":[]}`,
		"snyk-open-source": `{"projectName":"github.com/ioplane/scrapedoctl","dependencyCount":1}`,
		"snyk-code":        fmt.Sprintf("Testing %s ...\nTest type:         Static code analysis\n", repository),
		"snyk-container": `{"projectName":"docker-image|localhost/scrapedoctl",` +
			`"packageManager":"linux","path":"localhost/scrapedoctl:p0/scrapedoctl"}`,
	}

	for _, spec := range auditSpecs(image) {
		path := filepath.Join(outputDir, spec.outputName)
		if spec.name == "snyk-code" {
			path = filepath.Join(outputDir, spec.name+".log")
		}
		if err := os.WriteFile(path, []byte(fixtures[spec.name]), 0o600); err != nil {
			t.Fatal(err)
		}
		corpus, err := validateAuditEvidence(repository, outputDir, image, spec, 1)
		if err != nil {
			t.Fatalf("%s valid evidence: %v", spec.name, err)
		}
		if corpus == "" {
			t.Fatalf("%s returned empty corpus evidence", spec.name)
		}
	}

	semgrep := auditSpecs(image)[0]
	if err := os.WriteFile(
		filepath.Join(outputDir, semgrep.outputName),
		[]byte(`{"version":"1.175.0","paths":{"scanned":[]},"errors":[]}`),
		0o600,
	); err != nil {
		t.Fatal(err)
	}
	if _, err := validateAuditEvidence(repository, outputDir, image, semgrep, 1); err == nil {
		t.Fatal("empty Semgrep corpus passed the release gate")
	}
}
