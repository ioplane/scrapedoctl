package mcp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/ioplane/scrapedoctl/pkg/scrapedo"
)

func TestMCPServer_Resource(t *testing.T) {
	client, _ := scrapedo.NewClient("test-token")
	server, _ := NewServerWithClient(client)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientTransport, serverTransport := mcp.NewInMemoryTransports()
	go func() {
		_ = server.Run(ctx, serverTransport)
	}()

	mcpClient := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "1.0.0"}, nil)
	session, err := mcpClient.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer session.Close()

	// Read Resource
	res, err := session.ReadResource(ctx, &mcp.ReadResourceParams{
		URI: "resource://cli/help",
	})
	if err != nil {
		t.Fatalf("ReadResource failed: %v", err)
	}

	if len(res.Contents) == 0 {
		t.Fatal("Expected content in resource result")
	}
	if res.Contents[0].URI != "resource://cli/help" {
		t.Errorf("Unexpected URI: %s", res.Contents[0].URI)
	}
}

func TestMCPServer_Tool(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("mocked markdown content"))
	}))
	defer ts.Close()

	client, _ := scrapedo.NewClient("test-token")
	client.SetBaseURL(ts.URL)
	server, _ := NewServerWithClient(client)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientTransport, serverTransport := mcp.NewInMemoryTransports()
	go func() {
		_ = server.Run(ctx, serverTransport)
	}()

	mcpClient := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "1.0.0"}, nil)
	session, err := mcpClient.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer session.Close()

	// Call Tool
	args := map[string]any{"url": "http://example.com"}
	res, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      "scrape_url",
		Arguments: args,
	})

	if err != nil {
		t.Fatalf("CallTool failed: %v", err)
	}

	if res.IsError {
		t.Fatalf("Expected no error, got: %v", res.Content)
	}

	if len(res.Content) == 0 {
		t.Fatal("Expected content in tool result")
	}

	textContent, ok := res.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("Expected TextContent, got %T", res.Content[0])
	}

	if textContent.Text != "mocked markdown content" {
		t.Errorf("Expected 'mocked markdown content', got '%s'", textContent.Text)
	}
}

func TestMCPServer_ToolError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("api error"))
	}))
	defer ts.Close()

	client, _ := scrapedo.NewClient("test-token")
	client.SetBaseURL(ts.URL)
	server, _ := NewServerWithClient(client)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientTransport, serverTransport := mcp.NewInMemoryTransports()
	go func() {
		_ = server.Run(ctx, serverTransport)
	}()

	mcpClient := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "1.0.0"}, nil)
	session, err := mcpClient.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer session.Close()

	args := map[string]any{"url": "http://example.com"}
	res, _ := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      "scrape_url",
		Arguments: args,
	})

	if !res.IsError {
		t.Error("Expected IsError=true")
	}
}

func TestNewServer_Error(t *testing.T) {
	_, err := NewServer("")
	if err == nil {
		t.Error("Expected error with empty token")
	}
}
