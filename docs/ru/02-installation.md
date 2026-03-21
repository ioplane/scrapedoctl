# 02 - Установка и настройка

## Системные требования

- **Go 1.26+** (для сборки из исходников)
- **Podman/Docker** (опционально, для разработки в контейнерах)
- **API токен Scrape.do** (доступен на [scrape.do](https://scrape.do/))

## Сборка из исходников

Чтобы собрать бинарный файл `scrapedoctl` локально:

```bash
# Клонируйте репозиторий
git clone https://github.com/ioplane/scrapedoctl.git
cd scrapedoctl

# Соберите бинарник
go build -o bin/scrapedoctl ./cmd/scrapedoctl
```

## Интерактивная установка

`scrapedoctl` включает встроенный интерактивный установщик, который настроит файл конфигурации и автоматически интегрирует утилиту с вашими AI-агентами.

Чтобы запустить установщик, просто выполните любую команду без файла конфигурации:

```bash
./bin/scrapedoctl scrape https://example.com
```

## Автодополнение (Shell Completion)

`scrapedoctl` поддерживает автоматическое дополнение команд для Bash, Zsh, Fish и PowerShell.

### Bash
Добавьте следующую строку в ваш `~/.bashrc`:
```bash
source <(scrapedoctl completion bash)
```

### Zsh
Добавьте следующую строку в ваш `~/.zshrc`:
```zsh
source <(scrapedoctl completion zsh)
```

### Oh My Zsh
Если вы используете [Oh My Zsh](https://ohmyz.sh/), вы можете создать файл автодополнения вручную:
```bash
mkdir -p ~/.oh-my-zsh/completions
scrapedoctl completion zsh > ~/.oh-my-zsh/completions/_scrapedoctl
```
После этого перезапустите терминал или выполните `source ~/.zshrc`.

### PowerShell
`scrapedoctl` предоставляет нативный модуль PowerShell для автодополнения команд, совместимый с PowerShell 7.6+ в Windows, Linux и macOS.

#### Установка
1. Сгенерируйте модуль и манифест:
   ```powershell
   scrapedoctl completion powershell > scrapedoctl.psm1
   # Релизы также включают предварительно сгенерированный манифест scrapedoctl.psd1
   ```
2. Импортируйте модуль:
   ```powershell
   Import-Module ./scrapedoctl.psm1
   ```
3. Чтобы настройка сохранялась, добавьте команду импорта в ваш `$PROFILE`.

#### Особенности для PowerShell 7.4+
- Поддержка `NativeCommandErrorActionPreference` для улучшения обработки ошибок.
- Оптимизировано для кроссплатформенного использования в Unix-системах.
