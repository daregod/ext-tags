# ext-tags

[![Go Reference](https://pkg.go.dev/badge/github.com/yourname/external-tags.svg)](https://pkg.go.dev/github.com/yourname/external-tags)

`ext-tags` — это утилита командной строки на Go, которая извлекает метаданные (теги и обложки) из аудиофайлов и сохраняет их в отдельные файлы `.tag` (в бинарном формате APEv2). 

Файлы `.tag` полностью совместимы с плагином [foo_external_tags](https://foobar.hyv.fi/?view=foo_external_tags) для плеера **foobar2000**. Утилита позволяет хранить метаданные отдельно от аудиофайлов, не изменяя сами медиафайлы.

## 📦 Поддерживаемые форматы

На данный момент утилита умеет извлекать теги из следующих контейнеров:
*   **MP3** (`.mp3`) — ID3v1, ID3v2.2, ID3v2.3, ID3v2.4
*   **FLAC** (`.flac`) — Vorbis Comments
*   **Opus** (`.opus`) — Vorbis Comments (в контейнере Ogg)
*   **Ogg Vorbis** (`.ogg`) — Vorbis Comments
*   **M4A / MP4** (`.m4a`, `.mp4`) — iTunes-атомы
*   **WAV** (`.wav`) — ID3v2 чанки
*   **AIFF** (`.aif`, `.aiff`) — ID3v2 чанки

## ✨ Особенности

*   **Чистый Go**: Без CGO и внешних зависимостей вроде `ffmpeg`.
*   **Обложки**: Извлекает встроенные изображения (Front Cover) и корректно упаковывает их в APEv2 (с префиксом имени файла, как того требует спецификация).
*   **Нормализация ключей**: Приводит теги к единому виду (например, `ORGANIZATION` -> `Publisher`, `DATE` -> `Year`), предотвращая дублирование полей в foobar2000.
*   **Безопасность**: Не изменяет оригинальные аудиофайлы, а лишь создаёт рядом файл `<имя_файла>.<расширение>.tag`.

## 🚀 Сборка и установка

Для сборки нужны Go 1.24+ и GNU Make 4.x. В Windows всё ставится через scoop:

```powershell
scoop install make golangci-lint
```

```bash
git clone https://github.com/yourname/external-tags.git
cd external-tags

make setup    # проверка тулчейна, загрузка модулей, установка линтера
make build    # bin/ext-tags.exe (Windows) или bin/ext-tags
make run ARGS="track.flac"
```

Если make использовать не хочется, сборка вручную:

```bash
go build -trimpath -o bin/ext-tags ./cmd/ext-tags
```

## 🛠 Использование

```bash
./bin/ext-tags track.flac track2.mp3
./bin/ext-tags --version
```

Рядом с аудиофайлом появится `track.flac.tag`:

```text
OK    track.flac → track.flac.tag (18 items, 145032 bytes)
```

Если тегов нет, файл `.tag` не создаётся.

## 🏷 Версии и changelog

Проект следует [SemVer](https://semver.org/lang/ru/): ветка `0.x` — MVP,
ломающие изменения повышают minor. История изменений ведётся в
[CHANGELOG.md](CHANGELOG.md).

Версия не хранится в коде: она берётся из git-тега через `git describe`,
поэтому любой бинарник знает, из какого коммита он собран.

```bash
make version          # что попадёт в -ldflags
make check            # тесты и линтер перед релизом
make tag TAG=v0.4.0   # аннотированный тег на текущий коммит
```

Порядок релиза: перенести пункты из `[Unreleased]` в новый раздел
CHANGELOG, закоммитить, прогнать `make check`, поставить тег `make tag`,
при необходимости отправить его в удалённый репозиторий (`git push --tags`).

## ⚙️ Автоматизация

Вся рутина собрана в `Makefile`. `make` без аргументов печатает справку по целям.

| Цель | Действие |
|------|----------|
| `make setup` | проверка окружения, `go mod download`, установка `golangci-lint` (scoop в Windows, `go install` в остальных ОС) |
| `make build` | сборка `bin/ext-tags[.exe]`; версия, коммит и дата передаются через `-ldflags` |
| `make run ARGS="..."` | запуск CLI без сборки |
| `make install` | установка в `GOBIN` |
| `make dist` | кросс-компиляция под windows/linux/darwin (amd64 + arm64) в `dist/` |
| `make test` | юнит-тесты |
| `make cover`, `make cover-html` | покрытие и HTML-отчёт |
| `make bench` | бенчмарки |
| `make vet`, `make lint`, `make lint-fix` | `go vet` и `golangci-lint` (конфиг `.golangci.yml`) |
| `make fmt`, `make fmt-check` | форматирование и его проверка |
| `make tidy`, `make tidy-check` | `go mod tidy` и проверка, что `go.mod`/`go.sum` не разъехались |
| `make check` | `fmt-check` + `vet` + `lint` + `test` |
| `make ci` | `tidy-check` + `check` — то же, что гоняет CI |
| `make clean` | удаляет артефакты сборки (`bin/ext-tags*`, `dist/`, coverage) |

GitHub Actions (`.github/workflows/ci.yml`) на каждый push и pull request запускает `make tidy-check`, `make check`, `make build` и `make dist`.

Версии, теги и порядок релиза описаны в разделе «Версии и changelog» выше.
