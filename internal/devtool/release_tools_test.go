package devtool

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
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
