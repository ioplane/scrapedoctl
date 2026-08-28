package devtool

import (
	"slices"
	"testing"
)

func TestAuditCommandsAreFixedAndShellFree(t *testing.T) {
	specs := auditSpecs("localhost/scrapedoctl:p0")
	if len(specs) != 4 {
		t.Fatalf("audit command count = %d, want 4", len(specs))
	}
	wantTools := []string{"semgrep", "snyk", "snyk", "snyk"}
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
