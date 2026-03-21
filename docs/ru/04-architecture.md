# 04 - Архитектура и дизайн

## Обзор системы

`scrapedoctl` построен как модульная система, разделяющая API-клиент, слой хранения и интерфейсы взаимодействия (CLI/MCP).

### Схема работы

```mermaid
graph TD
    User[Пользователь / AI-агент] --> CLI[Точка входа CLI]
    CLI --> Config[Загрузчик конфига]
    Config --> Cache{Проверка кеша}
    Cache -- Попадание --> Result[Возврат результата]
    Cache -- Промах --> API[Scrape.do API]
    API --> Save[Сохранение в SQLite]
    Save --> Result
```

## Model Context Protocol (MCP)

Реализация MCP позволяет любому совместимому клиенту (например, Claude Desktop или VS Code) использовать `scrapedoctl` как удаленный инструмент.

### Последовательность взаимодействия

```mermaid
sequenceDiagram
    participant Agent as AI-агент
    participant Server as MCP-сервер
    participant DB as Кеш SQLite
    participant API as Scrape.do API

    Agent->>Server: listTools()
    Server-->>Agent: [scrape_url]
    Agent->>Server: callTool(url, render=true)
    Server->>DB: GetLatest(url_hash)
    alt Кеш найден
        DB-->>Server: контент
    else Кеш не найден
        Server->>API: GET /scrape?url=...
        API-->>Server: 200 OK (Markdown)
        Server->>DB: InsertScrape(url, контент)
    end
    Server-->>Agent: ToolResult(контент)
```

## Слой хранения (SQLite)

Персистентный слой использует pure-Go реализацию SQLite (`modernc.org/sqlite`) в сочетании с `sqlc` для типобезопасного доступа к данным и `goose` для управления версиями миграций.

- **Нормализация запросов**: Все запросы нормализуются (сортировка параметров/заголовков) перед хешированием для обеспечения точности попадания в кеш.
- **Авто-очистка**: База данных самостоятельно управляет дисковым пространством на основе параметров конфигурации `keep_versions` и `max_size_mb`.
