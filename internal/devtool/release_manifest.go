package devtool

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const maxReleaseManifestSize = 128 << 20

var errInvalidReleaseManifest = errors.New("invalid release manifest")

type spdxDocument struct {
	SPDXVersion       string `json:"spdxVersion"`
	DataLicense       string `json:"dataLicense"`
	SPDXID            string `json:"SPDXID"`
	Name              string `json:"name"`
	DocumentNamespace string `json:"documentNamespace"`
	CreationInfo      struct {
		Creators []string `json:"creators"`
	} `json:"creationInfo"`
	Packages []struct {
		Name string `json:"name"`
	} `json:"packages"`
}

// ValidateReleaseManifest verifies every published executable bundle and its SPDX SBOM.
func ValidateReleaseManifest(manifestPath string) error {
	entries, err := readReleaseManifest(manifestPath)
	if err != nil {
		return err
	}
	directory := filepath.Dir(manifestPath)
	artifacts := 0
	for name, digest := range entries {
		if !isReleaseBundle(name) {
			continue
		}
		artifacts++
		if err := verifyManifestFile(directory, name, digest); err != nil {
			return err
		}
		sbomName := name + ".sbom.json"
		sbomDigest, ok := entries[sbomName]
		if !ok {
			return fmt.Errorf("%w: %s has no SBOM", errInvalidReleaseManifest, name)
		}
		if err := verifyManifestFile(directory, sbomName, sbomDigest); err != nil {
			return err
		}
		if err := validateSPDX(filepath.Join(directory, sbomName)); err != nil {
			return fmt.Errorf("validate %s: %w", sbomName, err)
		}
	}
	if artifacts == 0 {
		return fmt.Errorf("%w: no release bundles", errInvalidReleaseManifest)
	}
	return nil
}

func readReleaseManifest(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open release manifest: %w", err)
	}
	defer file.Close()

	entries := make(map[string]string)
	scanner := bufio.NewScanner(io.LimitReader(file, maxReleaseManifestSize+1))
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) != 2 || len(fields[0]) != sha256.Size*2 {
			return nil, fmt.Errorf("%w: malformed checksum line", errInvalidReleaseManifest)
		}
		if _, err := hex.DecodeString(fields[0]); err != nil {
			return nil, fmt.Errorf("%w: malformed digest", errInvalidReleaseManifest)
		}
		if _, exists := entries[fields[1]]; exists {
			return nil, fmt.Errorf("%w: duplicate %s", errInvalidReleaseManifest, fields[1])
		}
		entries[fields[1]] = fields[0]
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read release manifest: %w", err)
	}
	return entries, nil
}

func isReleaseBundle(name string) bool {
	return strings.HasSuffix(name, ".tar.gz") || strings.HasSuffix(name, ".zip") ||
		strings.HasSuffix(name, ".deb") || strings.HasSuffix(name, ".rpm")
}

func verifyManifestFile(directory, name, expected string) error {
	path, err := archiveTarget(directory, name)
	if err != nil {
		return err
	}
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open release artifact %s: %w", name, err)
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return fmt.Errorf("hash release artifact %s: %w", name, err)
	}
	if actual := hex.EncodeToString(hash.Sum(nil)); actual != expected {
		return fmt.Errorf("%w: checksum mismatch for %s", errInvalidReleaseManifest, name)
	}
	return nil
}

func validateSPDX(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("inspect SPDX document: %w", err)
	}
	if info.Size() > maxReleaseManifestSize {
		return fmt.Errorf("%w: SPDX document exceeds size limit", errInvalidReleaseManifest)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read SPDX document: %w", err)
	}
	var document spdxDocument
	if err := json.Unmarshal(data, &document); err != nil {
		return fmt.Errorf("decode SPDX document: %w", err)
	}
	if !strings.HasPrefix(document.SPDXVersion, "SPDX-2.") ||
		document.DataLicense != "CC0-1.0" ||
		document.SPDXID != "SPDXRef-DOCUMENT" ||
		document.Name == "" || document.DocumentNamespace == "" ||
		len(document.CreationInfo.Creators) == 0 {
		return fmt.Errorf("%w: incomplete SPDX document metadata", errInvalidReleaseManifest)
	}
	for _, pkg := range document.Packages {
		if pkg.Name != "" {
			return nil
		}
	}
	return fmt.Errorf("%w: SPDX document has no package metadata", errInvalidReleaseManifest)
}
