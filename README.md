# ext-tags

[![CI](https://github.com/daregod/ext-tags/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/daregod/ext-tags/actions/workflows/ci.yml)
[![Coverage Status](https://coveralls.io/repos/github/daregod/ext-tags/badge.svg?branch=main)](https://coveralls.io/github/daregod/ext-tags?branch=main)
[![Release](https://img.shields.io/github/v/release/daregod/ext-tags?label=release&sort=semver)](https://github.com/daregod/ext-tags/releases/latest)
[![Go version](https://img.shields.io/github/go-mod/go-version/daregod/ext-tags)](https://github.com/daregod/ext-tags/blob/main/go.mod)
[![License](https://img.shields.io/github/license/daregod/ext-tags)](LICENSE)
[![Platforms](https://img.shields.io/badge/platform-Windows%20%7C%20Linux%20%7C%20macOS-lightgrey)](#)

`ext-tags` — консольная утилита для Windows, Linux и macOS, которая складывает
теги аудиофайлов в отдельные `.tag`-файлы, совместимые с компонентом
**foo_external_tags** для foobar2000.

Главное: **аудиофайлы не изменяются**. Рядом с треком появляется
`track.flac.tag`, а сам `track.flac` остаётся байт-в-байт прежним.

## Термины

- **Встроенный тег** — метаданные внутри аудиоконтейнера (ID3, Vorbis, MP4 и
  т.п.). В коде и конфиге — `embedded`.
- **Внешний тег** или **`.tag`** — отдельный файл `track.flac.tag` рядом с
  аудиофайлом.
- **Существующий `.tag`** — внешний тег, который уже лежит рядом с исходным
  аудиофайлом; его поля подмешиваются при `merge_existing`.
- **Исходный аудиофайл** — входной `track.flac`/`track.mp3`; слово
  «исходный» относится к аудио, а не к `.tag`.
- **Выходной `.tag`** — файл, который создаёт `extract`; может лежать в другом
  месте при `--output`.
- **Sidecar** — файл изображения рядом с аудиофайлом: `cover.jpg`,
  `folder.png` и т.п. Это значение `cover.prefer: sidecar`.

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
- **Обратный импорт.** Когда теги всё-таки нужно записать в файлы,
  `ext-tags import` пишет `.tag` обратно в MP3, FLAC, Ogg, MP4, APEv2,
  Matroska и ASF/WMA.

## Сценарии

### Скачал альбом, у всех файлов есть checksumm, смена тегов его нарушит

```bash
cd ~/Music/Artist/Album
ext-tags extract .
```

Рядом с каждым треком появятся `.tag`-файлы. В foobar2000 добавь папку —
компонент foo_external_tags подхватит теги, а файлы останутся нетронутыми.

## Обложки для foo_external_tags

Положи `cover.jpg` (или `folder.jpg`, `front.jpg`, `album.jpg`, `poster.jpg`)
в папку с альбомом и запусти:

```bash
ext-tags extract .
```

Встроенная обложка имеет приоритет; если её нет, будет взята картинка из
папки. Поддерживаются все пять artwork-слотов foobar2000: front, back, artist,
disc, icon.

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

### Вернуть теги в файлы

Если внешние теги больше не нужны или файл будет проигрываться в плеере без поддержки внешних тегов. Тот же вызов работает для всех форматов, где есть writer:

```bash
ext-tags import --backup --id3v2-version 3 track.mp3
```

Будет создан `track.mp3.bak`, а теги из `track.mp3.tag` запишутся в ID3v2.3.

## Установка

Подробные варианты: [INSTALLING.md](INSTALLING.md) — Releases, `go install`,
Scoop, Homebrew, AUR, `.deb`, `.rpm`, `.apk`.

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

| Команда | Назначение                                                     |
|---|----------------------------------------------------------------|
| `ext-tags extract <пути...>` | создать `.tag` (по умолчанию)                                  |
| `ext-tags import <файлы...>` | записать `.tag` обратно в MP3, FLAC, Ogg, MP4, APEv2, Matroska |
| `ext-tags album init <папка>` | создать `album.yaml` из тегов треков                           |
| `ext-tags artwork list/extract` | посмотреть список обложек или вытащить конкретную              |
| `ext-tags config init`, `validate`, `show`, `path` | работа с конфигом                                              |
| `ext-tags fixext <пути...>` | переименовать Ogg-файлы с неверным расширением (dry-run)       |
| `ext-tags render <пути...>` | вывести поля по шаблону, ничего не записывая                   |
| `ext-tags self-update` | обновить standalone-бинарник из GitHub Release                 |
| `ext-tags version` | версия и дата сборки                                           |

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

### extract: куда писать

По умолчанию `.tag` создаётся рядом с аудиофайлом. Для пайплайнов можно
записать результат в явный путь:

```bash
ext-tags extract --output /sandbox/track.mp3.tag /library/track.mp3
ext-tags extract --output - track.mp3
ext-tags extract --stdout track.mp3
```

`--output PATH` не создаёт файл рядом с исходником. Существующий `.tag`
читается рядом с исходником, а `--on-existing` применяется к `PATH`.
`--stdout` пишет байты `.tag` в stdout; `--json` с ним несовместим, а
`--dry-run` запрещён. `--output`/`--stdout` требуют ровно один входной файл.

`--output` можно использовать с расширением, отличным от `.tag`, — это
ответственность вызывающего кода.

### import: проверка после записи

По умолчанию импорт проверяет, что ключевые поля читаются после записи, и при
расхождении завершается ошибкой. Режим можно изменить:

```bash
ext-tags import --verify=error track.mp3  # по умолчанию
ext-tags import --verify=warn  track.mp3  # предупреждение, импорт продолжается
ext-tags import --verify=off   track.mp3  # не проверять
```

В режиме `warn` в отчёте будет `SKIP field ...`, но импорт продолжится.
Запись идёт через temp-файл с заменой только после проверки, поэтому при
ошибке исходный файл остаётся целым.

### fixext

Некоторые стриминговые сервисы отдают Ogg/Opus с расширением `.flac`.
foobar2000 выбирает artwork-ридер по расширению и не показывает обложки у таких
файлов. `fixext` определяет контейнер по магии и переименовывает файл в
правильное расширение (`.opus`, `.ogg`, `.oga`, `.spx`), вместе с `.tag`:

```bash
# dry-run: только показать, что будет переименовано
ext-tags fixext -r ~/Music

# применить
ext-tags fixext -r ~/Music --apply
```

По умолчанию ничего не меняется. Если целевое имя уже существует, файл
пропускается.

### artwork

`artwork` работает с уже выбранной обложкой, но не меняет `.tag`:

```bash
ext-tags artwork list album/*.flac
ext-tags artwork list --json track.flac
ext-tags artwork extract --slot front --output cover.jpg track.flac
ext-tags artwork extract --slot back --output back.png track.flac --force
```

`list` печатает выбранные слоты с учётом `cover.prefer`,
`cover.max_bytes`/`hard_max_bytes`, `album.yaml` и существующего `.tag`.
`extract` пишет байты одного слота в файл; по умолчанию слот `front`.

### render

`render` выводит поля по Go-шаблону и **ничего не пишет на диск**. Он использует
тот же pipeline, что `extract`: встроенные теги, существующий `.tag`,
`album.yaml`, правила и `include`/`exclude`.

```bash
ext-tags render --template '{{ .artist }} - {{ .title }}' album/*.flac
ext-tags render --json --template '{{ .tracknumber }} - {{ .title }}' track.flac
ext-tags render --null --template '{{ .path }}' album/*.flac
```

`--json` печатает NDJSON:

```json
{"file":".../track.flac","value":"01 - Title"}
```

`--null` заменяет разделитель записей на NUL — удобно для `xargs -0`.
Для безопасных имён файлов есть функция `sanitize`:

```bash
ext-tags render --template '{{ sanitize .artist }} - {{ sanitize .title }}' album/*.flac
```

`sanitize` заменяет `\ / : * ? " < > |`, управляющие символы и опасные
крайние пробелы/точки, а также защищает от зарезервированных имён Windows.
Unicode-символы, включая `։`, `⁄`, `˃`, `˂`, `✶`, он не переписывает.

## Самообновление

Standalone-бинарник можно обновлять из GitHub Release:

```bash
ext-tags self-update --check     # только проверить версию
ext-tags self-update --dry-run   # показать, что будет скачано
ext-tags self-update             # скачать, проверить SHA256 и заменить
ext-tags self-update --force     # переустановить/обновить принудительно
```

Команда появилась в 0.11.0; более старую standalone-копию нужно один раз
обновить вручную. Обновление проверяет `SHA256SUMS` из релиза и заменяет файл атомарно.
Для установок через `opkg`, `apt`, Homebrew и Scoop команда откажется
обновлять файл и покажет правильную команду пакетного менеджера. `--force`
обходит эту защиту, но пакетный менеджер может перезаписать обновление.

## Конфигурация

Без конфига работают встроенные значения. Конфиг ищется слоями.
YAML-файлы могут быть в UTF-8 с BOM и с CRLF — это не ошибка.

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
  names: [cover, folder, front, album, poster, box]  # front
  back: [back]
  artist: [artist]
  disc: [disc, media]
  icon: [icon]

fields:
  rules:
    - default: {field: "Album artist", from: Artist}
    - remove: [Www]
```

Правила (`set`, `copy`, `rename`, `default`, `remove`) и шаблоны —
продвинутая настройка; полный список ключей и примеры даёт
`ext-tags config init --full`.

### Профили

Профиль — именованный набор переопределений в обычном конфиге:

```yaml
profiles:
  phone:
    fields:
      include: [Artist, Title, Album, "Album artist", Track, Genre, Year, Releasetime, Rating, "Cover Art (Front)"]
      exclude: ["Cover Art (Back)", "Artist Picture", "Cover Art (Disc)", "Cover Art (Icon)"]
    cover:
      enabled: true
```

```bash
ext-tags --profile phone extract --output /sandbox/track.mp3.tag track.mp3
```

`--profile` без `--config` ищет профиль в пользовательском и проектном слоях
конфига; отдельный файл не нужен.

### include/exclude

`fields.include` и `fields.exclude` применяются к итоговому набору — после
aliases, merge существующего `.tag` и правил.

```yaml
fields:
  include: [Artist, Title, Album, "Album artist", Track, Genre, Year, Releasetime, Rating, "Cover Art (Front)"]
  exclude: [Www, Encoder]
```

Обложки тоже имеют имена для фильтров. Можно оставить только front:

```yaml
fields:
  exclude:
    - "Cover Art (Back)"
    - "Artist Picture"
    - "Cover Art (Disc)"
    - "Cover Art (Icon)"
```

Имена слотов:

| Слот | Имя для include/exclude |
|---|---|
| front | `cover.field` из конфига, по умолчанию `Cover Art (Front)` |
| back | `Cover Art (Back)` |
| artist | `Artist Picture` |
| disc | `Cover Art (Disc)` |
| icon | `Cover Art (Icon)` |

Многострочные значения полей нормализуются в `\r\n`, хвостовые пробелы и
табы в строках обрезаются. Это касается и `Www`, и `track_url`, и других
пользовательских полей.

Если `include` непустой, обложки, которых нет в списке, **не попадут** в
результат. Если `include` не задан, обложки остаются по `cover.enabled`, а
`exclude` убирает перечисленные слоты.

Список полей не является закрытым. Незнакомое поле из контейнера сохраняется
под своим именем, если его не убирает `fields.ignore`; его можно использовать
в `include`/`exclude`, правилах и `album.yaml`. Своё имя задаётся через
`fields.aliases`:

```yaml
fields:
  aliases:
    CUSTOM_TAG: My field
  include: [Artist, Title, "My field"]
```

Без алиаса поле остаётся как есть:

```yaml
fields:
  include: [Artist, Title, CUSTOM_TAG]
```

Значение можно добавить только для альбома:

```yaml
album:
  My field: значение из буклета
```

### Нумерация треков

Есть два разных механизма, и они не взаимозаменяемы:

- встроенное правило собирает `Track` из отдельных переменных
  `.tracknumber` и `.totaltracks`;
- функция `formatTrack` форматирует **уже готовое** значение вида `X/Y`,
  которое лежит в поле `.track`.

Встроенное правило:

```yaml
- set: Track
  template: '{{ .tracknumber }}{{ if .totaltracks }}/{{ .totaltracks }}{{ end }}'
```

оставляет номер как есть: `2/10`, `1/8`, `7`. Оно не добавляет ведущие нули.

`formatTrack` работает с готовым `X/Y` и выравнивает номер по ширине общего
количества:

```yaml
- set: Track
  template: '{{ formatTrack .track }}'
```

| Значение `.track` | Результат `formatTrack` |
|---|---|
| `1/12` | `01/12` |
| `7/12` | `07/12` |
| `1/8` | `1/8` |
| `7` | `7` |

Если встроенное правило уже отработало, оно сначала записывает `Track`, поэтому
`formatTrack .track` получает готовое значение:

| `.tracknumber` | `.totaltracks` | Встроенный `Track` | `formatTrack .track` после него |
|---|---|---|---|
| `2` | `10` | `2/10` | `02/10` |
| `1` | `8` | `1/8` | `1/8` |
| `7` | — | `7` | `7` |

Поэтому `formatTrack .track` имеет смысл только если поле `Track` уже
заполнено. При `fields.rules_mode: replace` встроенное правило не выполняется,
и `.track` может быть пустым — тогда используйте `.tracknumber`/
`.totaltracks`. Если нужно одновременно собрать `X/Y` из переменных и
добавить ведущие нули:

```yaml
- set: Track
  when: '{{ .totaltracks }}'
  template: '{{ formatTrack (printf "%s/%s" .tracknumber .totaltracks) }}'
```

Для `Disc` действует та же логика: `.discnumber`, `.disctotal`, поле `Disc`.

Перед правилами поля из существующего `.tag` уже подмешаны. Если в нём есть
`Track: 1/16` или `Disc: 1/2`, недостающие `.tracknumber`/`.totaltracks` и
`.discnumber`/`.disctotal` восстанавливаются из этого значения. Поэтому
правило без `when` не затрёт `1/16` голым `1`, если в контейнере нет
`TRACKTOTAL`. Аналогично `Releasetime` из `.tag` виден условию
`{{ eq .releasetime "" }}` и не перезаписывается годом.

### Пример правил

Шаблоны умеют `year` (выделить год), `formatTrack` (отформатировать `X/Y`
по количеству треков), `atoi` (преобразовать строку в число) и `sanitize`
(сделать безопасное имя файла). Пример правил, которые можно положить в
`ext-tags.yaml`:

```yaml
cover:
  prefer: sidecar          # embedded | sidecar | existing

fields:
  rules:
    - set: Releasetime
      when: '{{ eq .releasetime "" }}'
      template: '{{ .year }}'
    - set: Year
      template: '{{ year .releasetime }}'
    - default: {field: "Comment catalog number", from: Catalognumber}
    - default: {field: "Catalog #", from: Catalognumber}
    - default: {field: Publisher, from: Label}
    - set: Artist
      template: '{{ replace ", " " & " .artist }}'
```

Шаблонные ключи — без пробелов и в нижнем регистре: `.track`, `.artist`,
`.releasetime`, `.catalognumber`. Встроенное правило `Track` уже применяется
по умолчанию; не добавляйте своё `set: Track`, если не хотите его заменить.

### Конфиг альбома

Рядом с аудиофайлами можно положить `album.yaml` (или `album.yml`,
`.ext-tags.album.yaml`). Он нужен только для данных, которых нет во
встроенных тегах или которые нужно явно переопределить. Не дублируй в нём
поля, уже правильно заполненные в аудио.

Для этого альбома embedded-теги уже содержат альбом, артиста, дату, жанр,
лейбл и каталожный номер. Во внешнем `.tag` не хватает только списка `Www`
и полной `cover.jpg`, поэтому достаточно такого оверлея:

```yaml
album:
  Www: |
    https://www.beatport.com/release/the-future-is-now/1755915
    https://www.discogs.com/master/1869963-Shiny-Radio-The-Future-Is-Now
    https://www.discogs.com/release/11490696-Shiny-Radio-The-Future-Is-Now
    https://www.discogs.com/release/16475013-Shiny-Radio-The-Future-Is-Now
    https://www.junodownload.com/products/shiny-radio-the-future-is-now/3081831-02/
    https://tidal.com/album/60225885
    https://www.deezer.com/album/13108392

artwork:
  front: cover.jpg
```

Если конкретный трек нужно поправить, добавь только его:

```yaml
tracks:
  "4":
    Title: Signature of The Soul
```

Скелет `album.yaml` можно создать командой:

```bash
ext-tags album init /path/to/album
ext-tags album init --dry-run /path/to/album
ext-tags album init --no-merge /path/to/album
ext-tags album init --merge-album /path/to/album
ext-tags album init --force /path/to/album
```

`--no-merge` не читает существующие `.tag` (по умолчанию они подмешиваются
при `files.merge_existing: true`). `--merge-album` дополняет существующий
`album.yaml`, сохраняя уже заданные значения; `--merge-album` и `--force`
несовместимы.

Она читает теги всех треков в папке, подмешивает поля из существующих
`.tag` (как `extract` при `files.merge_existing`), выносит общие поля в
`album:`, различия — в `tracks:`, а найденные sidecar-обложки — в
`artwork:`. Количество треков подставляется в `Track` с ведущими
нулями по ширине общего числа: `1/8`, `01/12`, `001/100`. Флаг
`--no-merge` отключает чтение существующих `.tag`.

Ключи `tracks` — номер трека (`"1"`, `"01"`), числовой префикс имени файла
(`01. Artist - Track`), имя файла или имя без расширения. Номер без ведущих
нулей и с ними считается одним и тем же треком. Поля `Album`, `Album artist`, `Label`, `Catalognumber` и т.п.
можно задавать в `album:` при необходимости, но не нужно повторять их,
если они уже верны во встроенных тегах.

Поля из `album:` применяются ко всем трекам и перекрывают встроенные теги и
существующий `.tag`; правила видят уже переопределённые значения. Если у
отдельного трека другой жанр — не задавай `Genre` на уровне альбома:
оставь встроенное значение или переопредели только этот трек.

Семантика значений:

```text
ключа нет       -> не трогать
ключ есть, ""   -> удалить поле
ключ есть, "x"  -> переопределить на x
```

Например, чтобы у трека 1 заменить URL в `Comment` на `File webpage URL`
и убрать исходный комментарий:

```yaml
tracks:
  "1":
    File webpage URL: https://tidal.com/browse/track/60225886
    Comment: ""
```

```yaml
tracks:
  "8":
    Genre: Breaks / Breakbeat / UK Bass
```

`album.yaml` — декларативные данные, а не команды редактирования:
`extract` по-прежнему только пишет `.tag`.
## Обложки

foobar2000 знает пять типов artwork; каждый хранится в `.tag` под своим ключом:

| Слот foobar2000 | Ключ в `.tag` | Имена файлов по умолчанию |
|---|---|---|
| Front Cover | `Cover Art (Front)` | `cover`, `folder`, `front`, `album`, `poster`, `box` |
| Back Cover | `Cover Art (Back)` | `back` |
| Artist | `Artist Picture` | `artist` |
| Disc | `Cover Art (Disc)` | `disc`, `media` |
| Icon | `Cover Art (Icon)` | `icon` |

1. `cover.prefer` задаёт порядок кандидатов: `embedded`, `sidecar` или
   `existing`.
2. Остальные слоты ищутся в папке по именам из таблицы; расширения —
   `.jpg`, `.jpeg`, `.png`, `.webp`, `.gif` (регистр не важен).
3. Найденные картинки попадают в `.tag` под соответствующими ключами.
4. Размер кандидатов ограничивается `cover.max_bytes`/`hard_max_bytes`
   (см. ниже).

Отключить всё: `--cover=false` или `cover.enabled: false`. Поиск в папке
отключается через `cover.directory: false`. Списки имён настраиваются:

```yaml
cover:
  enabled: true
  directory: true
  names: [cover, folder, front, album, poster, box]
  back: [back]
  artist: [artist]
  disc: [disc, media]
  icon: [icon]
  max_bytes: 0             # 0 = без мягкого бюджета
  hard_max_bytes: 8388608  # 8 MiB; 0 = без жёсткого потолка
```

### Бюджет обложек

Размер ограничивается без декодирования картинок — только по длине файла:

- `max_bytes` — предпочтительный бюджет. Если кандидат в него не влезает,
  выбирается следующий по `prefer`;
- `hard_max_bytes` — абсолютный потолок. Кандидаты больше него не
  рассматриваются;
- если ни один кандидат не влез в `max_bytes`, но есть кандидат под
  `hard_max_bytes`, берётся самый маленький из подходящих;
- если все кандидаты больше `hard_max_bytes` — обложка не вставляется.

Кандидаты перебираются в порядке `cover.prefer`:

```text
embedded, sidecar, existing  (prefer: embedded)
sidecar, embedded, existing  (prefer: sidecar)
existing, sidecar, embedded  (prefer: existing)
```

Явная обложка из `album.yaml` (`artwork:`) рассматривается первой.

Пример для телефона:

```yaml
cover:
  prefer: sidecar
  max_bytes: 2097152
  hard_max_bytes: 8388608
```

Так sidecar 1.5 МБ будет выбран сразу, sidecar 3 МБ уступит встроенной
1.2 МБ, а при отсутствии подходящих вариантов останется обложка до 8 МБ.
Лимиты применяются и к `extract`, и к `import`.

Важно: бюджет **не влияет на порядок выбора sidecar-файла**. `cover.names`
(а также `cover.back`, `cover.artist`, …) — это порядок поиска: берётся первый
существующий файл, и только потом проверяется его размер. Если в папке лежат
`cover.jpg` (10 МБ) и `front.jpg` (1 МБ), при списке
`names: [cover, folder, front, box]` будет найден `cover.jpg`; `front.jpg`
**не будет** рассмотрен как альтернатива. Если `cover.jpg` больше
`hard_max_bytes`, sidecar-кандидат отбрасывается целиком; если он больше
только `max_bytes`, он может остаться fallback-кандидатом, но `front.jpg`
всё равно не проверяется. В этом случае можно изменить порядок `cover.names`,
убрать большое имя из списка или подготовить подходящий файл.

## Поддерживаемые форматы

MP3 (ID3v1/v2.2–2.4), MP2, MP1, M4A/MP4 (AAC, ALAC), FLAC, Ogg Vorbis, Opus,
Ogg FLAC, Speex, APE, WavPack, Musepack, TAK, WAV, RF64, Wave64, AIFF,
WMA/WMV/ASF, Matroska/WebM, AC3 и DTS (APEv2 в конце файла — так их тегирует
foobar2000).

Форматы без тегов (AU, TS, Audio CD) не поддерживаются — см.
[TODO.md](TODO.md). В MPEG-TS нет контейнера тегов, и foobar2000 не читает даже заполненный внешний
`.tag` (видит только обложку).

## Решение проблем

**В Opus-файле с расширением `.flac` foobar2000 не видит обложки.** Это Ogg/Opus-контейнер, но foobar2000 выбирает artwork-ридер по расширению. Переименуй файл в `.opus` — встроенные обложки появятся (проверено). `ext-tags import` предупреждает о таком случае.

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

- [TAGS.md](TAGS.md) — справочник полей, алиасов и artwork-имён.
- [DEVELOPMENT.md](DEVELOPMENT.md) — архитектура, тесты, фикстуры, релизы.
- [TODO.md](TODO.md) — форматы и планы.
- [CHANGELOG.md](CHANGELOG.md) — история версий.
