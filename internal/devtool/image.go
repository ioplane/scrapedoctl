package devtool

import (
	"context"
	"fmt"
	"io"
	"path/filepath"

	"github.com/testcontainers/testcontainers-go"
)

const (
	imageRepository = "localhost/scrapedoctl"
	imageTag        = "p0"
	// DefaultAuditImage is the retained local image scanned by the default audit command.
	DefaultAuditImage = imageRepository + ":" + imageTag
)

// BuildAuditImage builds and retains the local production image through Testcontainers-Go and Podman.
func BuildAuditImage(ctx context.Context, repository string, output io.Writer) error {
	if err := configurePodman(); err != nil {
		return fmt.Errorf("configure Podman: %w", err)
	}

	_, _ = fmt.Fprintf(output, "building local audit image %s\n", DefaultAuditImage)
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ProviderType:     testcontainers.ProviderPodman,
		ContainerRequest: imageRequest(repository),
	})
	if err != nil {
		return fmt.Errorf("build audit image: %w", err)
	}
	if err := testcontainers.TerminateContainer(container); err != nil {
		return fmt.Errorf("remove audit build container: %w", err)
	}

	return nil
}

func imageRequest(repository string) testcontainers.ContainerRequest {
	return testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:    filepath.Clean(repository),
			Dockerfile: "Containerfile",
			Repo:       imageRepository,
			Tag:        imageTag,
			KeepImage:  true,
		},
	}
}
