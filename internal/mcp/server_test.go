package mcp_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	mcp_sdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ioplane/scrapedoctl/internal/mcp"
	"github.com/ioplane/scrapedoctl/pkg/scrapedo"
	"github.com/ioplane/scrapedoctl/pkg/search"
)

func TestMCPServer_Resource(t *testing.T) {
	client, _ := scrapedo.NewClient("test-token")
	server, _ := mcp.NewServerWithClient(client)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientTransport, serverTransport := mcp_sdk.NewInMemoryTransports()
	go func() {
		_ = server.Run(ctx, serverTransport)
	}()

	mcpClient := mcp_sdk.NewClient(&mcp_sdk.Implementation{Name: "test", Version: "1.0.0"}, nil)
	session, err := mcpClient.Connect(ctx, clientTransport, nil)
	require.NoError(t, err)
	defer session.Close()

	// Read Resource
	res, err := session.ReadResource(ctx, &mcp_sdk.ReadResourceParams{
		URI: "resource://cli/help",
	})
	require.NoError(t, err)

	require.NotEmpty(t, res.Contents)
	assert.Equal(t, "resource://cli/help", res.Contents[0].URI)
}

func TestMCPServer_Tool(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("mocked markdown content"))
	}))
	defer ts.Close()

	client, _ := scrapedo.NewClient("test-token")
	client.SetBaseURL(ts.URL)
	server, _ := mcp.NewServerWithClient(client)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientTransport, serverTransport := mcp_sdk.NewInMemoryTransports()
	go func() {
		_ = server.Run(ctx, serverTransport)
	}()

	mcpClient := mcp_sdk.NewClient(&mcp_sdk.Implementation{Name: "test", Version: "1.0.0"}, nil)
	session, err := mcpClient.Connect(ctx, clientTransport, nil)
	require.NoError(t, err)
	defer session.Close()

	// Call Tool
	args := map[string]any{"url": "http://example.com"}
	res, err := session.CallTool(ctx, &mcp_sdk.CallToolParams{
		Name:      "scrape_url",
		Arguments: args,
	})

	require.NoError(t, err)
	assert.False(t, res.IsError)
	require.NotEmpty(t, res.Content)

	textContent, ok := res.Content[0].(*mcp_sdk.TextContent)
	require.True(t, ok)
	assert.Equal(t, "mocked markdown content", textContent.Text)
}

func TestMCPServer_ToolError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("api error"))
	}))
	defer ts.Close()

	client, _ := scrapedo.NewClient("test-token")
	client.SetBaseURL(ts.URL)
	server, _ := mcp.NewServerWithClient(client)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientTransport, serverTransport := mcp_sdk.NewInMemoryTransports()
	go func() {
		_ = server.Run(ctx, serverTransport)
	}()

	mcpClient := mcp_sdk.NewClient(&mcp_sdk.Implementation{Name: "test", Version: "1.0.0"}, nil)
	session, err := mcpClient.Connect(ctx, clientTransport, nil)
	require.NoError(t, err)
	defer session.Close()

	args := map[string]any{"url": "http://example.com"}
	res, _ := session.CallTool(ctx, &mcp_sdk.CallToolParams{
		Name:      "scrape_url",
		Arguments: args,
	})

	assert.True(t, res.IsError)
}

func TestNewServer_Error(t *testing.T) {
	_, err := mcp.NewServer("")
	assert.Error(t, err)
}

func TestMCPServer_SearchTool(t *testing.T) {
	// Mock a search API server that returns a valid JSON response.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		resp := map[string]any{
			"search_information": map[string]any{"total_results": 1},
			"organic_results": []map[string]any{
				{
					"position":       1,
					"title":          "Test Result",
					"link":           "https://example.com",
					"snippet":        "A test snippet",
					"displayed_link": "example.com",
				},
			},
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	// Create a scrapedo provider pointing at the mock server.
	provider := search.NewScrapedoProvider("test-token")
	provider.SetBaseURL(ts.URL)

	router := search.NewRouter()
	router.Register(provider)

	client, _ := scrapedo.NewClient("test-token")
	server, err := mcp.NewServerWithClientAndRouter(client, router)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientTransport, serverTransport := mcp_sdk.NewInMemoryTransports()
	go func() {
		_ = server.Run(ctx, serverTransport)
	}()

	mcpClient := mcp_sdk.NewClient(&mcp_sdk.Implementation{Name: "test", Version: "1.0.0"}, nil)
	session, err := mcpClient.Connect(ctx, clientTransport, nil)
	require.NoError(t, err)
	defer session.Close()

	args := map[string]any{"query": "test query", "engine": "google"}
	res, err := session.CallTool(ctx, &mcp_sdk.CallToolParams{
		Name:      "web_search",
		Arguments: args,
	})

	require.NoError(t, err)
	assert.False(t, res.IsError)
	require.NotEmpty(t, res.Content)

	textContent, ok := res.Content[0].(*mcp_sdk.TextContent)
	require.True(t, ok)
	assert.Contains(t, textContent.Text, "Test Result")
	assert.Contains(t, textContent.Text, "example.com")
}

func TestMCPServer_SearchTool_EmptyQuery(t *testing.T) {
	provider := search.NewScrapedoProvider("test-token")
	router := search.NewRouter()
	router.Register(provider)

	client, _ := scrapedo.NewClient("test-token")
	server, err := mcp.NewServerWithClientAndRouter(client, router)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientTransport, serverTransport := mcp_sdk.NewInMemoryTransports()
	go func() {
		_ = server.Run(ctx, serverTransport)
	}()

	mcpClient := mcp_sdk.NewClient(&mcp_sdk.Implementation{Name: "test", Version: "1.0.0"}, nil)
	session, err := mcpClient.Connect(ctx, clientTransport, nil)
	require.NoError(t, err)
	defer session.Close()

	args := map[string]any{"query": ""}
	res, err := session.CallTool(ctx, &mcp_sdk.CallToolParams{
		Name:      "web_search",
		Arguments: args,
	})

	require.NoError(t, err)
	assert.True(t, res.IsError)
}

func TestMCPServer_NoSearchToolWithoutRouter(t *testing.T) {
	client, _ := scrapedo.NewClient("test-token")
	server, err := mcp.NewServerWithClient(client)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientTransport, serverTransport := mcp_sdk.NewInMemoryTransports()
	go func() {
		_ = server.Run(ctx, serverTransport)
	}()

	mcpClient := mcp_sdk.NewClient(&mcp_sdk.Implementation{Name: "test", Version: "1.0.0"}, nil)
	session, err := mcpClient.Connect(ctx, clientTransport, nil)
	require.NoError(t, err)
	defer session.Close()

	// web_search tool should not exist when no router is provided.
	tools, err := session.ListTools(ctx, nil)
	require.NoError(t, err)
	for _, tool := range tools.Tools {
		assert.NotEqual(t, "web_search", tool.Name, "web_search tool should not exist without router")
	}
}
