package devtool

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ulikunitz/xz"
)

const (
	upxVersion         = "5.2.1"
	powershellVersion  = "7.6.5"
	upxDigest          = "402162aad30af47e60dbd767fb2e64ca394ace9727ba1f40283641f1d1b91657"
	powershellDigest   = "b34ab3b19acac1d3d4d0d3cfdb02acf62f457b0b6a962ff008132033f7566844"
	maxReleaseToolSize = 128 << 20
	maxExtractedSize   = 1 << 30
)

var (
	errReleaseAssetStatus = errors.New("release asset returned a non-success status")
	errReleaseAssetSize   = errors.New("release asset exceeds size limit")
	errReleaseAssetDigest = errors.New("release asset digest mismatch")
	errReleaseArchivePath = errors.New("release archive contains an unsafe path")
	errReleaseArchiveType = errors.New("release archive contains an unsupported entry")
)

// InstallReleaseTools downloads, verifies, and installs the exact release tools.
func InstallReleaseTools(ctx context.Context, repository string, output io.Writer) error {
	root := filepath.Join(repository, ".release-tools")
	if err := os.RemoveAll(root); err != nil {
		return fmt.Errorf("reset release tools: %w", err)
	}
	binDir := filepath.Join(root, "bin")
	if err := os.MkdirAll(binDir, 0o700); err != nil {
		return fmt.Errorf("create release tools directory: %w", err)
	}

	client := &http.Client{Timeout: 5 * time.Minute}
	upxArchive, powershellArchive, err := downloadReleaseToolArchives(ctx, client, root)
	if err != nil {
		return err
	}

	upxDir := filepath.Join(root, "upx")
	powershellDir := filepath.Join(root, "powershell")
	if err := extractReleaseArchive(upxArchive, upxDir, "xz"); err != nil {
		return fmt.Errorf("extract UPX: %w", err)
	}
	if err := extractReleaseArchive(powershellArchive, powershellDir, "gzip"); err != nil {
		return fmt.Errorf("extract PowerShell: %w", err)
	}
	if err := copyExecutable(
		filepath.Join(upxDir, "upx-"+upxVersion+"-amd64_linux", "upx"),
		filepath.Join(binDir, "upx"),
	); err != nil {
		return fmt.Errorf("install UPX executable: %w", err)
	}
	//nolint:gosec // The verified release executable must be owner-executable.
	if err := os.Chmod(filepath.Join(powershellDir, "pwsh"), 0o700); err != nil {
		return fmt.Errorf("make PowerShell executable: %w", err)
	}
	if err := os.Symlink(filepath.Join("..", "powershell", "pwsh"), filepath.Join(binDir, "pwsh")); err != nil {
		return fmt.Errorf("link PowerShell executable: %w", err)
	}
	if err := appendGitHubPath(binDir); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(
		output,
		"installed verified UPX %s and PowerShell %s in %s\n",
		upxVersion,
		powershellVersion,
		binDir,
	)
	return nil
}

func downloadReleaseToolArchives(
	ctx context.Context,
	client *http.Client,
	root string,
) (string, string, error) {
	upxArchive := filepath.Join(root, "upx.tar.xz")
	powershellArchive := filepath.Join(root, "powershell.tar.gz")
	if err := downloadVerified(
		ctx,
		client,
		"https://github.com/upx/upx/releases/download/v"+upxVersion+
			"/upx-"+upxVersion+"-amd64_linux.tar.xz",
		upxDigest,
		upxArchive,
	); err != nil {
		return "", "", fmt.Errorf("install UPX: %w", err)
	}
	if err := downloadVerified(
		ctx,
		client,
		"https://github.com/PowerShell/PowerShell/releases/download/v"+powershellVersion+
			"/powershell-"+powershellVersion+"-linux-x64.tar.gz",
		powershellDigest,
		powershellArchive,
	); err != nil {
		return "", "", fmt.Errorf("install PowerShell: %w", err)
	}
	return upxArchive, powershellArchive, nil
}

func downloadVerified(
	ctx context.Context,
	client *http.Client,
	rawURL string,
	expectedDigest string,
	destination string,
) (returnErr error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, http.NoBody)
	if err != nil {
		return fmt.Errorf("create release asset request: %w", err)
	}
	request.Header.Set("User-Agent", "scrapedoctl-release-tools")
	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("download release asset: %w", err)
	}
	defer func() {
		returnErr = errors.Join(returnErr, response.Body.Close())
	}()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("%w: %s", errReleaseAssetStatus, response.Status)
	}

	temporary, err := os.CreateTemp(filepath.Dir(destination), ".release-asset-*")
	if err != nil {
		return fmt.Errorf("create release asset: %w", err)
	}
	temporaryPath := temporary.Name()
	defer func() {
		if removeErr := os.Remove(temporaryPath); removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
			returnErr = errors.Join(returnErr, fmt.Errorf("remove temporary release asset: %w", removeErr))
		}
	}()

	hash := sha256.New()
	written, copyErr := io.Copy(
		io.MultiWriter(temporary, hash),
		io.LimitReader(response.Body, maxReleaseToolSize+1),
	)
	closeErr := temporary.Close()
	if copyErr != nil {
		return fmt.Errorf("write release asset: %w", copyErr)
	}
	if closeErr != nil {
		return fmt.Errorf("close release asset: %w", closeErr)
	}
	if written > maxReleaseToolSize {
		return errReleaseAssetSize
	}
	actualDigest := hex.EncodeToString(hash.Sum(nil))
	if actualDigest != expectedDigest {
		return fmt.Errorf("%w: expected %s, got %s", errReleaseAssetDigest, expectedDigest, actualDigest)
	}
	if err := os.Rename(temporaryPath, destination); err != nil {
		return fmt.Errorf("commit release asset: %w", err)
	}
	return nil
}

func extractReleaseArchive(archive, destination, compression string) (returnErr error) {
	if err := os.MkdirAll(destination, 0o700); err != nil {
		return fmt.Errorf("create extraction directory: %w", err)
	}
	input, err := os.Open(archive)
	if err != nil {
		return fmt.Errorf("open release archive: %w", err)
	}
	defer func() { returnErr = errors.Join(returnErr, input.Close()) }()

	reader, err := releaseArchiveReader(input, compression)
	if err != nil {
		return err
	}
	return extractTarArchive(reader, destination)
}

func releaseArchiveReader(input io.Reader, compression string) (io.Reader, error) {
	switch compression {
	case "xz":
		reader, err := xz.NewReader(input)
		if err != nil {
			return nil, fmt.Errorf("open XZ release archive: %w", err)
		}
		return reader, nil
	case "gzip":
		reader, err := gzip.NewReader(input)
		if err != nil {
			return nil, fmt.Errorf("open gzip release archive: %w", err)
		}
		return reader, nil
	default:
		return nil, fmt.Errorf("%w: compression %q", errReleaseArchiveType, compression)
	}
}

func extractTarArchive(reader io.Reader, destination string) error {
	total := int64(0)
	archiveReader := tar.NewReader(reader)
	for {
		header, nextErr := archiveReader.Next()
		if errors.Is(nextErr, io.EOF) {
			return nil
		}
		if nextErr != nil {
			return fmt.Errorf("read release archive: %w", nextErr)
		}
		total += header.Size
		if header.Size < 0 || total > maxExtractedSize {
			return errReleaseAssetSize
		}
		if err := extractTarEntry(archiveReader, destination, header); err != nil {
			return err
		}
	}
}

func extractTarEntry(reader io.Reader, destination string, header *tar.Header) error {
	target, err := archiveTarget(destination, header.Name)
	if err != nil {
		return err
	}
	switch header.Typeflag {
	case tar.TypeDir:
		if err := os.MkdirAll(target, 0o700); err != nil {
			return fmt.Errorf("create release archive directory: %w", err)
		}
		return nil
	case tar.TypeReg:
		return extractReleaseFile(reader, target)
	default:
		return fmt.Errorf("%w: %q", errReleaseArchiveType, header.Name)
	}
}

func archiveTarget(destination, name string) (string, error) {
	root, err := filepath.Abs(destination)
	if err != nil {
		return "", fmt.Errorf("resolve extraction directory: %w", err)
	}
	target := filepath.Join(root, filepath.Clean(name))
	if target == root || !strings.HasPrefix(target, root+string(os.PathSeparator)) {
		return "", fmt.Errorf("%w: %q", errReleaseArchivePath, name)
	}
	return target, nil
}

func extractReleaseFile(reader io.Reader, target string) error {
	if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
		return fmt.Errorf("create release file directory: %w", err)
	}
	file, err := os.OpenFile(target, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("create release file: %w", err)
	}
	if _, err := io.Copy(file, reader); err != nil {
		return errors.Join(fmt.Errorf("extract release file: %w", err), file.Close())
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close release file: %w", err)
	}
	return nil
}

func copyExecutable(source, destination string) (returnErr error) {
	input, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("open executable: %w", err)
	}
	defer func() { returnErr = errors.Join(returnErr, input.Close()) }()
	//nolint:gosec // The verified release executable must be owner-executable.
	output, err := os.OpenFile(destination, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o700)
	if err != nil {
		return fmt.Errorf("create executable: %w", err)
	}
	if _, err := io.Copy(output, input); err != nil {
		return errors.Join(fmt.Errorf("copy executable: %w", err), output.Close())
	}
	if err := output.Close(); err != nil {
		return fmt.Errorf("close executable: %w", err)
	}
	return nil
}

func appendGitHubPath(path string) error {
	githubPath := os.Getenv("GITHUB_PATH")
	if githubPath == "" {
		return nil
	}
	githubPath = filepath.Clean(githubPath)
	info, err := os.Lstat(githubPath)
	if err != nil {
		return fmt.Errorf("inspect GITHUB_PATH: %w", err)
	}
	if !filepath.IsAbs(githubPath) || !info.Mode().IsRegular() {
		return fmt.Errorf("%w: GITHUB_PATH", errReleaseArchivePath)
	}
	file, err := os.OpenFile(githubPath, os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("open GITHUB_PATH: %w", err)
	}
	if _, err := fmt.Fprintln(file, path); err != nil {
		return errors.Join(fmt.Errorf("append GITHUB_PATH: %w", err), file.Close())
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close GITHUB_PATH: %w", err)
	}
	return nil
}
