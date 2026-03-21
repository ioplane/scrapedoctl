# 03 - Usage Guide

## Core Scrape Command

The `scrape` command is used for one-off scraping tasks. It outputs the result directly to your terminal.

```bash
scrapedoctl scrape https://google.com --render --super
```

### Options:
- `--render`: Enable JavaScript rendering (uses a real browser on Scrape.do servers).
- `--super`: Use residential/mobile proxies to bypass advanced bot protection.
- `--no-cache`: Bypass the local SQLite cache and force a new API call without saving.
- `--refresh`: Force a new API call and store the result as a new version in history.

## Interactive REPL

For sessions involving multiple URLs, use the built-in shell:

```bash
scrapedoctl repl
scrapedoctl> scrape https://example.com render=true
```

## Persistent Cache & History

`scrapedoctl` automatically saves successful scrapes to an internal SQLite database (`~/.scrapedoctl/cache.db`).

### View History:
```bash
scrapedoctl history https://example.com
```

### Cache Maintenance:
```bash
scrapedoctl cache stats   # See DB size and savings
scrapedoctl cache clear   # Wipe all stored results
```

## Configuration Management

You can manage your settings via the CLI or by editing `~/.scrapedoctl/conf.toml`.

```bash
scrapedoctl config list
scrapedoctl config set global.timeout=30000
```
