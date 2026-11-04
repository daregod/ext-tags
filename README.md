# ext-tags

[![CI](https://github.com/daregod/ext-tags/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/daregod/ext-tags/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/daregod/ext-tags?label=release&sort=semver)](https://github.com/daregod/ext-tags/releases/latest)
[![Go version](https://img.shields.io/github/go-mod/go-version/daregod/ext-tags)](https://github.com/daregod/ext-tags/blob/main/go.mod)
[![License](https://img.shields.io/github/license/daregod/ext-tags)](LICENSE)
[![Platforms](https://img.shields.io/badge/platform-Windows%20%7C%20Linux%20%7C%20macOS-lightgrey)](#)

`ext-tags` — консольная утилита для Windows, Linux и macOS, которая складывает
теги аудиофайлов в отдельные `.tag`-файлы, совместимые с компонентом
**foo_external_tags** для foobar2000.

Главное: **аудиофайлы не изменяются**. Рядом с треком появляется
`track.flac.tag`, а сам `track.flac` остаётся байт-в-байт прежним.

## Зачем это нужно

- **Иногда файлы менять нельзя.** Тогда теги можно держать рядом.
- **Теги неудобные или неполные.** Можно привести их к единому виду, убрать
  мусор, подставить `Album artist`, собрать `Track`/`Disc` и добавить свои
  поля — всё это в `.tag`, без правки контейнера.
- **Обложки.** foo_external_tags сам не создаёт обложки; `ext-tags` достаёт
  встроенную картинку или находит `cover.jpg`/`folder.jpg`/`front.jpg` и
  кладёт её в `.tag`.
- **Массовая обработка.** Один запуск на папку или всю библиотеку, с
  рекурсией, `--dry-run` и машиночитаемым `--json`.
- **Обратный импорт.** Когда теги всё-таки нужно записать в файлы, для MP3
  есть `ext-tags import` (ID3v1/ID3v2.2–2.4).

## Сценарии

### Скачал альбом, файлы портить нельзя

```bash
cd ~/Music/Artist/Album
ext-tags extract .
```

Рядом с каждым треком появятся `.tag`-файлы. В foobar2000 добавь папку —
компонент foo_external_tags подхватит теги, а файлы останутся нетронутыми.

### Обложки для foo_external_tags

Положи `cover.jpg` (или `folder.jpg`, `front.jpg`, `album.jpg`, `poster.jpg`)
в папку с альбомом и запусти:

```bash
ext-tags extract .
```

Встроенная обложка имеет приоритет; если её нет, будет взята картинка из
папки. Найденная обложка попадёт в `.tag` как `Cover Art (Front)`.

### Массовая нормализация тегов

`ext-tags.yaml` в корне библиотеки:

```yaml
fields:
  rules:
    - default: {field: "Album artist", from: Artist}
    - set: Track
      template: "{{ .tracknumber }}{{ if .totaltracks }}/{{ .totaltracks }}{{ end }}"
    - remove: [Www, Encoder]
```

```bash
cd ~/Music
ext-tags -r .
```

### Вернуть теги в файлы (MP3)

Если внешние теги больше не нужны или файл будет проигрываться в плеере без поддержки внешних тегов:

```bash
ext-tags import --backup --id3v2-version 3 track.mp3
```

Будет создан `track.mp3.bak`, а теги из `track.mp3.tag` запишутся в ID3v2.3.

## Установка

### Готовые сборки

Скачай бинарник для своей ОС из [последнего релиза](https://github.com/daregod/ext-tags/releases/latest)
или собери сам — см. «Из исходников». В релизе лежат архивы для Windows,
Linux и macOS (amd64/arm64), плюс `SHA256SUMS`. Внутри — один файл:
`ext-tags.exe` для Windows или `ext-tags` для Linux/macOS.

### Windows

1. Распакуй `ext-tags.exe` в любую папку, например
   `%LOCALAPPDATA%\ext-tags`.
2. Добавь эту папку в `PATH` (PowerShell):

```powershell
$dir = "$env:LOCALAPPDATA\ext-tags"
[Environment]::SetEnvironmentVariable(
  "Path", "$env:Path;$dir", "User")
```

3. Перезапусти терминал и проверь:

```powershell
ext-tags version
```

### Linux

```bash
chmod +x ext-tags_linux_amd64
sudo install -m 755 ext-tags_linux_amd64 /usr/local/bin/ext-tags
ext-tags version
```

Без root — в `~/.local/bin` (убедись, что он есть в `PATH`).

### macOS

```bash
chmod +x ext-tags_darwin_arm64          # или darwin_amd64 для Intel
sudo install -m 755 ext-tags_darwin_arm64 /usr/local/bin/ext-tags
ext-tags version
```

Если macOS ругается на неизвестного разработчика:

```bash
xattr -d com.apple.quarantine /usr/local/bin/ext-tags
```

### Из исходников

Нужен Go 1.26+ (и GNU Make 4.x — по желанию):

```bash
git clone https://github.com/daregod/ext-tags.git
cd external-tags

make build          # bin/ext-tags.exe или bin/ext-tags
# или без make:
go build -trimpath -o bin/ext-tags ./cmd/ext-tags
```

Для установки в `GOBIN`:

```bash
go install github.com/daregod/ext-tags/cmd/ext-tags@latest
```

## Быстрый старт

```bash
# один файл
ext-tags extract track.flac

# папка и рекурсия
ext-tags extract album/
ext-tags -r ~/Music

# без команды — то же, что extract (совместимость)
ext-tags track.flac

# посмотреть, что будет сделано
ext-tags --dry-run -r ~/Music

# машинный вывод (NDJSON, по объекту на файл)
ext-tags --json extract track.flac
```

После запуска рядом с файлом появится `track.flac.tag`. Если тегов нет и
обложки тоже нет, утилита напишет `SKIP` и ничего не создаст.

## Команды

| Команда | Назначение |
|---|---|
| `ext-tags extract <пути...>` | создать `.tag` (по умолчанию) |
| `ext-tags import <mp3...>` | записать `.tag` обратно в MP3 |
| `ext-tags config init`, `validate`, `show`, `path` | работа с конфигом |
| `ext-tags version` | версия и дата сборки |

Глобальные флаги:

| Флаг | Что делает |
|---|---|
| `--config PATH` | явный файл конфигурации |
| `--profile NAME` | профиль из конфига |
| `--dry-run` | показать, что было бы записано |
| `--json` | NDJSON-вывод для скриптов |
| `--quiet` | только ошибки |
| `-v`, `--verbose` | подробности и итоговый конфиг |
| `-r`, `--recursive` | обходить подпапки |
| `--on-existing MODE` | `overwrite`, `skip`, `backup`, `error` |
| `--cover=false` | не встраивать обложку |
| `--no-merge` | не подхватывать поля из старого `.tag` |

При `backup` (по умолчанию) старый `track.flac.tag` сохраняется как
`track.flac.tag.bak`.

## Конфигурация

Без конфига работают встроенные значения. Конфиг ищется слоями:

| Слой | Где |
|---|---|
| пользовательский | `%APPDATA%\ext-tags\config.yaml` (Windows), `$XDG_CONFIG_HOME/ext-tags/config.yaml` (Linux/macOS) |
| проект | `./ext-tags.yaml`, `./ext-tags.yml`, `./.ext-tags.yaml` |
| явный | `--config PATH` или `EXT_TAGS_CONFIG` |

```bash
ext-tags config init          # минимальный ext-tags.yaml
ext-tags config init --full   # полный дамп
ext-tags config validate      # проверить конфиг
ext-tags config show          # итоговый конфиг
ext-tags config path          # какие файлы применяются
```

Минимальный пример:

```yaml
files:
  recursive: true
  on_existing: backup

cover:
  enabled: true
  directory: true
  names: [cover, folder, front, album, poster, box, disc, media]

fields:
  rules:
    - default: {field: "Album artist", from: Artist}
    - remove: [Www]
```

Правила (`set`, `copy`, `rename`, `default`, `remove`) и шаблоны —
продвинутая настройка; полный список ключей и примеры даёт
`ext-tags config init --full`.

## Обложки

1. Сначала берётся встроенная картинка из файла.
2. Если её нет — ищется файл в папке: `cover`, `folder`, `front`, `album`,
   `poster`, `box`, `disc`, `media` с расширениями `.jpg`, `.jpeg`, `.png`,
   `.webp`, `.gif` (регистр не важен).
3. Найденная картинка попадает в `.tag` как `Cover Art (Front)`.

Отключить: `--cover=false` или `cover.enabled: false`. Поиск в папке
отключается через `cover.directory: false`.

## Поддерживаемые форматы

MP3 (ID3v1/v2.2–2.4), MP2, MP1, M4A/MP4 (AAC, ALAC), FLAC, Ogg Vorbis, Opus,
Ogg FLAC, Speex, APE, WavPack, Musepack, TAK, WAV, RF64, Wave64, AIFF,
WMA/WMV/ASF, Matroska/WebM, AC3 и DTS (APEv2 в конце файла — так их тегирует
foobar2000).

Форматы без тегов (AU, TS, Audio CD) не поддерживаются — см.
[TODO.md](TODO.md). В MPEG-TS нет контейнера тегов, и fb2k его тоже не тегирует
(проверено на 1-секундном audio-only фрагменте).

## Решение проблем

**`.tag` не появился.** Смотри строку `SKIP`: у файла нет тегов и обложки,
формат не поддерживается или сработала политика `on_existing: skip`.

**foobar2000 не видит теги.** Нужен установленный компонент
`foo_external_tags`; после добавления папки сделай Reload. `.tag` должен
лежать рядом с аудиофайлом и называться `<имя>.<расширение>.tag`.

**Не подхватывается обложка.** Проверь имена файлов (`cover.jpg`,
`folder.jpg`, `front.jpg`, `album.jpg`, `poster.jpg`) и что
`cover.enabled`/`cover.directory` не выключены.

**Нужно откатить изменения.** При `on_existing: backup` рядом остаётся
`.tag.bak`. Для `import` используй `--backup` — будет `<файл>.bak`.

**Другой компьютер/ОС.** Конфиг и `.tag` переносимы; бинарники — нет,
скачай сборку под нужную ОС.

## Документация

- [DEVELOPMENT.md](DEVELOPMENT.md) — архитектура, тесты, фикстуры, релизы.
- [TODO.md](TODO.md) — форматы и планы.
- [CHANGELOG.md](CHANGELOG.md) — история версий.
