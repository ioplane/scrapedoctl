# 04 - Architecture & Design

## High-level Overview

`scrapedoctl` is built as a modular system that decouples the API client, the storage layer, and the communication interfaces (CLI/MCP).

### System Flow Diagram

```mermaid
graph TD
    User[User / AI Agent] --> CLI[CLI Entrypoint]
    CLI --> Config[Config Loader]
    Config --> Cache{Cache Check}
    Cache -- Hit --> Result[Return Result]
    Cache -- Miss --> API[Scrape.do API]
    API --> Save[Save to SQLite]
    Save --> Result
```

## Model Context Protocol (MCP)

The MCP implementation allows any compatible client (like Claude Desktop or VS Code) to use `scrapedoctl` as a remote tool.

### Interaction Sequence

```mermaid
sequenceDiagram
    participant Agent as AI Agent
    participant Server as MCP Server
    participant DB as SQLite Cache
    participant API as Scrape.do API

    Agent->>Server: listTools()
    Server-->>Agent: [scrape_url]
    Agent->>Server: callTool(url, render=true)
    Server->>DB: GetLatest(url_hash)
    alt Cache Found
        DB-->>Server: content
    else Cache Miss
        Server->>API: GET /scrape?url=...
        API-->>Server: 200 OK (Markdown)
        Server->>DB: InsertScrape(url, content)
    end
    Server-->>Agent: ToolResult(content)
```

## Persistent Layer (SQLite)

The persistence layer uses a pure-Go SQLite implementation (`modernc.org/sqlite`) combined with `sqlc` for type-safe data access and `goose` for versioned migrations.

- **Request Normalization**: All requests are normalized (sorted params/headers) before hashing to ensure consistent cache lookup.
- **Auto-Cleanup**: The database self-manages disk space based on `keep_versions` and `max_size_mb` configuration settings.
