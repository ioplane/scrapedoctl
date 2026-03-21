package mcp

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunServer_ContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := RunServer(ctx, "test-token")
	// If initialization succeeds, it should return context.Canceled when trying to run the transport
	assert.Error(t, err)
}

func TestRunServerWithClient_ContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := RunServerWithClient(ctx, nil) // Should return error because client is nil
	assert.Error(t, err)
}

func TestNewServer_Success(t *testing.T) {
	server, err := NewServer("token")
	assert.NoError(t, err)
	assert.NotNil(t, server)
}
