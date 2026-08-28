package devtool

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestDownloadVerifiedRejectsDigestMismatch(t *testing.T) {
	t.Parallel()

	payload := []byte("verified release tool")
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		_, _ = response.Write(payload)
	}))
	t.Cleanup(server.Close)

	destination := filepath.Join(t.TempDir(), "tool.tar.gz")
	wantDigest := sha256.Sum256(payload)
	if err := downloadVerified(
		context.Background(),
		server.Client(),
		server.URL,
		hex.EncodeToString(wantDigest[:]),
		destination,
	); err != nil {
		t.Fatalf("download verified asset: %v", err)
	}
	if _, err := os.Stat(destination); err != nil {
		t.Fatalf("stat verified asset: %v", err)
	}

	if err := downloadVerified(
		context.Background(),
		server.Client(),
		server.URL,
		"0000000000000000000000000000000000000000000000000000000000000000",
		destination+".tampered",
	); err == nil {
		t.Fatal("digest mismatch was accepted")
	}
	if _, err := os.Stat(destination + ".tampered"); !os.IsNotExist(err) {
		t.Fatalf("tampered asset remains on disk: %v", err)
	}
}

func TestValidateReleaseManifestRequiresValidSBOMForEveryArtifact(t *testing.T) {
	t.Parallel()

	directory := t.TempDir()
	artifactName := "scrapedoctl_Linux_x86_64.tar.gz"
	sbomName := artifactName + ".sbom.json"
	artifact := []byte("release archive")
	sbom := []byte(`{
  "spdxVersion": "SPDX-2.3",
  "dataLicense": "CC0-1.0",
  "SPDXID": "SPDXRef-DOCUMENT",
  "name": "scrapedoctl",
  "documentNamespace": "https://example.test/scrapedoctl",
  "creationInfo": {"creators": ["Tool: syft-1.51.1"]},
  "packages": [{"name": "scrapedoctl"}]
}`)
	if err := os.WriteFile(filepath.Join(directory, artifactName), artifact, 0o600); err != nil {
		t.Fatalf("write artifact: %v", err)
	}
	if err := os.WriteFile(filepath.Join(directory, sbomName), sbom, 0o600); err != nil {
		t.Fatalf("write SBOM: %v", err)
	}
	manifest := filepath.Join(directory, "checksums.txt")
	artifactDigest := sha256.Sum256(artifact)
	sbomDigest := sha256.Sum256(sbom)
	valid := fmt.Sprintf("%x  %s\n%x  %s\n", artifactDigest, artifactName, sbomDigest, sbomName)
	if err := os.WriteFile(manifest, []byte(valid), 0o600); err != nil {
		t.Fatalf("write valid manifest: %v", err)
	}
	if err := ValidateReleaseManifest(manifest); err != nil {
		t.Fatalf("validate release manifest: %v", err)
	}

	missingSBOM := fmt.Sprintf("%x  %s\n", artifactDigest, artifactName)
	if err := os.WriteFile(manifest, []byte(missingSBOM), 0o600); err != nil {
		t.Fatalf("write incomplete manifest: %v", err)
	}
	if err := ValidateReleaseManifest(manifest); err == nil {
		t.Fatal("release artifact without SBOM was accepted")
	}
}
