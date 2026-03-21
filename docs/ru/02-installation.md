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
Для активации автодополнения в PowerShell импортируйте сгенерированный модуль:
```powershell
scrapedoctl completion powershell > scrapedoctl.psm1
Import-Module ./scrapedoctl.psm1
```
Чтобы настройка сохранялась, добавьте команду импорта в ваш `$PROFILE`.
