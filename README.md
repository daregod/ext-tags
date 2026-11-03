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

Для сборки необходим установленный Go 1.21 или выше.

```bash
git clone https://github.com/yourname/external-tags.git
cd external-tags
go mod tidy

# Сборка бинарника (Linux/macOS)
go build -o bin/ext-tags ./cmd/ext-tags

# Сборка бинарника (Windows)
go build -o bin/ext-tags.exe ./cmd/ext-tags