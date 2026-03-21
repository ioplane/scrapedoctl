# 03 - Руководство пользователя

## Основная команда Scrape

Команда `scrape` используется для разовых задач по скрапингу. Результат выводится прямо в терминал.

```bash
scrapedoctl scrape https://google.com --render --super
```

### Опции:
- `--render`: Включить рендеринг JavaScript (использует реальный браузер на серверах Scrape.do).
- `--super`: Использовать резидентные/мобильные прокси для обхода продвинутых защит.
- `--no-cache`: Игнорировать локальный кеш SQLite и выполнить новый запрос к API без сохранения.
- `--refresh`: Выполнить новый запрос к API и сохранить результат как новую версию в истории.

## Интерактивный REPL

Для сессий, включающих работу с несколькими URL, используйте встроенную оболочку:

```bash
scrapedoctl repl
scrapedoctl> scrape https://example.com render=true
```

## Персистентный кеш и история

`scrapedoctl` автоматически сохраняет успешные запросы во внутреннюю базу данных SQLite (`~/.scrapedoctl/cache.db`).

### Просмотр истории:
```bash
scrapedoctl history https://example.com
```

### Обслуживание кеша:
```bash
scrapedoctl cache stats   # Показать размер БД и экономию
scrapedoctl cache clear   # Полностью очистить сохраненные результаты
```

## Управление конфигурацией

Вы можете управлять настройками через CLI или редактируя `~/.scrapedoctl/conf.toml`.

```bash
scrapedoctl config list
scrapedoctl config set global.timeout=30000
```
