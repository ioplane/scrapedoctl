# 01 - Introduction

`scrapedoctl` is a professional-grade CLI tool and Model Context Protocol (MCP) server designed to bridge the gap between AI agents and the web. By utilizing the [Scrape.do](https://scrape.do/) API, it provides a robust, anti-bot-bypassing, and JS-rendering capable scraping engine that is both human and machine friendly.

## Project Goal

The primary goal of `scrapedoctl` is to provide AI agents (like Claude Code, Gemini CLI, etc.) with high-fidelity, LLM-optimized Markdown representations of web pages while minimizing costs through persistent caching and efficient request management.

## Key Features

- **Interactive REPL**: A rich shell environment for manual exploration.
- **Persistent Caching**: Built-in SQLite storage to save tokens and maintain history.
- **Machine Interface**: Full MCP support and JSON metadata for dynamic tool discovery.
- **Anti-Bot Bypassing**: Leveraging Scrape.do's proxy rotation and browser rendering.
- **Modern Architecture**: Written in Go 1.26, zero-dependency core, and strict linting.
