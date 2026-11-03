# ext-tags

[![Go Reference](https://pkg.go.dev/badge/github.com/yourname/external-tags.svg)](https://pkg.go.dev/github.com/yourname/external-tags)

`ext-tags` — это утилита командной строки на Go, которая извлекает метаданные (теги и обложки) из аудиофайлов и сохраняет их в отдельные файлы `.tag` (в бинарном формате APEv2). 

Файлы `.tag` полностью совместимы с плагином [foo_external_tags](https://foobar.hyv.fi/?view=foo_external_tags) для плеера **foobar2000**. Утилита позволяет хранить метаданные отдельно от аудиофайлов, не изменяя сами медиафайлы.

## 📦 Поддерживаемые форматы

На данный момент утилита умеет извлекать теги из следующих контейнеров:
*   **MP3 / MP2 / MP1** (`.mp3`, `.mp2`, `.mp1`) — ID3v1, ID3v2.2, ID3v2.3, ID3v2.4
*   **FLAC** (`.flac`) — Vorbis Comments
*   **Opus** (`.opus`) — Vorbis Comments (в контейнере Ogg)
*   **Ogg Vorbis** (`.ogg`) — Vorbis Comments
*   **M4A / MP4** (`.m4a`, `.mp4`, `.alac`) — iTunes-атомы (AAC и ALAC)
*   **AAC (raw)** (`.aac`) — ID3v2 перед ADTS-потоком
*   **WebM / MKV** (`.webm`, `.mkv`) — Matroska: `Tags/SimpleTag`, `Info/Title`, вложения-обложки, `PART_NUMBER`/`TOTAL_PARTS`
*   **APE** (`.ape`) — APEv2-тег в конце файла (Monkey Audio)
*   **WavPack** (`.wv`) — APEv2-тег в конце файла
*   **WMV / WMA** (`.wmv`, `.wma`, `.asf`) — ASF: Content Description и Extended Content Description, `WM/Picture` как обложка
*   **WAV** (`.wav`) — LIST/INFO и ID3v2, в том числе дописанный после RIFF-данных
*   **AIFF** (`.aif`, `.aiff`) — NAME/AUTH/ANNO и ID3v2-чанк

## ✨ Особенности

*   **Чистый Go**: Без CGO и внешних зависимостей вроде `ffmpeg`.
*   **Обложки**: Извлекает встроенные изображения (Front Cover) и корректно упаковывает их в APEv2 (с префиксом имени файла, как того требует спецификация).
*   **Нормализация ключей**: Приводит теги к единому виду (например, `ORGANIZATION` -> `Publisher`, `DATE` -> `Year`), предотвращая дублирование полей в foobar2000.
*   **Matroska**: читает `Tags/SimpleTag` (TITLE, ARTIST, GENRE, …), `Info/Title` и вложения-картинки как обложку; учитывает область тегов как foobar2000 (альбомные `ARTIST`/`PART_NUMBER` → `Album artist`/`Disc`). Track-level теги (`Artist` у дорожки) foobar2000 сам не показывает, но мы их переносим — в нашем `.tag` они появятся; `Cluster`/`Cues` пропускаются, большие видео не загружаются в память.
*   **APE и ASF**: APEv2 из хвоста `.ape` и Content Description из `.wmv`/`.wma`; технические поля (`WMFSDK*`, `IsVBR`, …) отбрасываются.
*   **Безопасность**: Не изменяет оригинальные аудиофайлы, а лишь создаёт рядом файл `<имя_файла>.<расширение>.tag`.

Полный список форматов и того, что ещё в планах, — в [TODO.md](TODO.md).

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
# обработка файлов и каталогов
ext-tags extract track.flac album/*.mp3
ext-tags extract -r D:\Music\Album        # рекурсивно
ext-tags --dry-run extract track.flac     # показать, что будет записано
ext-tags --json extract track.flac        # машинный вывод (по объекту на строку)
ext-tags track.flac                       # то же, что extract (обратная совместимость)

# конфигурация
ext-tags config init                      # создать ext-tags.yaml
ext-tags config validate                  # проверить конфиг и шаблоны
ext-tags config show                      # напечатать итоговый конфиг
ext-tags config path                      # какой файл используется

# версия
ext-tags version
ext-tags --version
```

Глобальные флаги: `--config PATH`, `--profile NAME`, `--dry-run`, `--json`,
`--quiet`, `-v/--verbose`, `-r/--recursive`,
`--on-existing=overwrite|skip|backup|error`, `--cover=false` (не встраивать
обложку), `--no-merge` (не подхватывать поля из старого `.tag`).

Рядом с аудиофайлом появляется `track.flac.tag`:

```text
OK      track.flac → track.flac.tag (21 items, 197434 bytes)
```

Если файл без тегов, формат не поддержан или `.tag` уже существует при
`on_existing: skip`, выводится `SKIP` с причиной. Код возврата `1` — только
при реальных ошибках чтения/записи.

`--json` печатает по объекту на строку (NDJSON):

```json
{"file":"track.flac","status":"ok","tagfile":"track.flac.tag","items":21,"bytes":197434,"fields":{"Title":"…","Artist":"…"},"cover":{"filename":"cover.jpg","bytes":196440}}
```

## ⚙️ Конфигурация

Поведение задаётся YAML-файлом; без него действуют встроенные значения
(они повторяют прежнее поведение утилиты).

### Где лежит конфиг

Файлы применяются слоями: встроенные значения → пользовательский конфиг →
конфиг проекта. Явный файл (`--config` или `EXT_TAGS_CONFIG`) заменяет оба
найденных и применяется поверх встроенных значений.

| Слой | Путь |
|---|---|
| пользовательский | `%APPDATA%\ext-tags\config.yaml`, `$XDG_CONFIG_HOME/ext-tags/config.yaml` |
| проект | `./ext-tags.yaml`, `./ext-tags.yml`, `./.ext-tags.yaml` |
| явный | `--config PATH` или `EXT_TAGS_CONFIG` |

Слои сливаются: скалярные поля переопределяются, `aliases` дополняются,
списки `extensions`, `ignore`, `include`, `exclude` заменяются целиком, а
`rules` накапливаются (встроенные → пользовательские → проектные). Правила
`Track`/`Disc` продолжают работать, даже если слой добавляет своё правило;
чтобы слой заменил все предыдущие правила, укажите `fields.rules_mode:
replace`. Личные правила удобно держать в пользовательском конфиге, а
настройки библиотеки — в `ext-tags.yaml` в её корне.

Конфиг библиотеки лучше держать в её корне и запускать утилиту оттуда:
`cd D:\Music && ext-tags -r .`. Профили читаются из файла с наибольшим
приоритетом.

```bash
ext-tags config init          # минимальный ext-tags.yaml (только частые ключи)
ext-tags config init --full   # полный дамп со всеми значениями
ext-tags config validate      # проверить структуру и шаблоны правил
ext-tags config show          # напечатать итоговый конфиг (с учётом --profile)
ext-tags config path          # показать, какие файлы применяются
```

`config init` не перезаписывает существующий файл без `--force`. В минимальном
файле перечислены только часто меняемые ключи, остальное берётся из дефолтов —
так обновления встроенных значений продолжают действовать.

### Пример

```yaml
version: 1

files:
  extensions: [.mp3, .flac, .opus, .ogg, .m4a, .mp4, .webm, .mkv, .ape, .wmv, .wma, .asf, .wav, .aif, .aiff]
  recursive: false
  on_existing: backup          # overwrite | skip | backup | error
  merge_existing: true         # сохранять поля из старого .tag (Rating и т. п.)

tagfile:
  suffix: .tag                 # foo_external_tags: <name>.<ext>.tag

cover:
  enabled: true
  field: Cover Art (Front)
  filename: cover              # базовое имя в .tag (расширение — по картинке)
  names: [cover, folder, box, disc, media]  # какие имена вложений считать обложкой

fields:
  aliases:                     # ключ контейнера -> поле foobar2000
    ORGANIZATION: Publisher
    DATE: Year
    "©nam": Title
  ignore: [vendor, iTunNORM, iTunSMPB, "Encoding Params"]
  rules:                       # применяются по порядку, видят результат друг друга
    - set: Track
      template: "{{ .tracknumber }}{{ if .totaltracks }}/{{ .totaltracks }}{{ end }}"
    - copy: {from: Label, to: Publisher}
    - rename: {from: "Comment catalog number", to: "Catalog#"}
    - default: {field: "Album artist", from: Artist}
    - remove: [Www]
  include: []                  # пусто — оставить все поля
  exclude: []
```

### Правила

| Правило | Что делает |
|---|---|
| `set: Field` + `template` | вычисляет значение и записывает поле |
| `copy: {from, to}` | дублирует поле |
| `rename: {from, to}` | переносит поле под новым именем |
| `default: {field, from}` | заполняет поле из другого, если оно пустое |
| `remove: [Field, …]` | удаляет поля |

У любого правила может быть `when` — шаблон-условие: правило выполняется,
если результат не пустой, не `false` и не `0`.

Правила из файлов применяются поверх встроенных, поэтому `Track`/`Disc`
продолжают собираться, если вы добавили только своё правило. Полностью
заменить встроенные правила можно через `fields.rules_mode: replace`.

### Шаблоны

Шаблоны — стандартный Go `text/template`. Данные: все поля в виде
идентификатора (`Album artist` → `.albumartist`, `Title` → `.title`) и
нормализованные значения `.tracknumber`, `.totaltracks`, `.discnumber`,
`.disctotal`, а также `.filename`, `.dirname`, `.ext`, `.path`, `.now`.

Функции: `lower`, `upper`, `title`, `trim`, `default "…" .field`,
`replace "a" "b" .field`, `regexReplace "^pattern" "replacement" .field`.

### Профили

Именованные наборы переопределений поверх базового конфига:

```yaml
profiles:
  music:
    files:
      extensions: [.flac, .mp3]
  video:
    files:
      recursive: true
    cover:
      enabled: false
```

```bash
ext-tags --profile music extract -r D:\Music
```

Профиль накладывается на базовые значения: то, что в нём не указано,
остаётся из базовой части файла (списки заменяются целиком, `rules`
накапливаются, `aliases` дополняются).

### Существующие `.tag` и слияние

`files.on_existing` — что делать с уже существующим `<file>.tag`:

* `backup` (по умолчанию) — старый файл переименовывается в `<file>.tag.bak`,
  затем пишется новый;
* `overwrite` — перезаписать;
* `skip` — оставить как есть и вывести `SKIP ... (already exists)`;
* `error` — считать это ошибкой (код возврата `1`).

`files.merge_existing` (по умолчанию `true`) — перед записью прочитать старый
`.tag` и сохранить поля, которых нет в аудиофайле: `Rating`, обложку из
медиатеки и т. п. Поля, извлечённые из аудио, всегда перекрывают старые.

| `on_existing` | merge | Результат |
|---|---|---|
| `backup` | да | `.bak` = состояние до слияния, новый `.tag` = аудио + сохранённые поля |
| `overwrite` | да | старый перезаписан; уникальные поля сохранены, но исходник нигде не остаётся |
| `skip` | да | записи нет — merge не применяется (утилита предупреждает об этом) |
| `error` | да | падение до записи — merge не применяется |
| `backup` | нет | `.bak` = старый файл, новый `.tag` только из аудио |

Отключить слияние: `--no-merge` или `files.merge_existing: false`. Комбинация
merge со `skip`/`error` не является ошибкой (merge может быть включён
глобально, а `skip` — локально): при `-v` и в `config validate` печатается
предупреждение, что слияние не сработает.

Перегенерировать `.tag` строго из аудио — `--no-merge`; бэкап при этом
по-прежнему создаётся.

## 🔍 Проверка совместимости

### Автоматическая: `make compat`

Команда сравнивает то, что генерирует утилита, с эталонными `.tag`, которые
создал сам плагин `foo_external_tags` (файлы лежат рядом с аудио в
`bin/examples`, в git они не входят):

```bash
make compat
```

Команда собирает `bin/compat.exe` и запускает его; вручную её можно вызвать
как `bin/compat.exe -examples bin/examples`.

Разбор вывода:

* `OK` — поля и обложки совпадают;
* `changed=[…]` — реальное расхождение (команда вернёт код `1`);
* `cosmetic=[…]` — различия, не влияющие на чтение: имя файла внутри обложки
  (`cover.jpg` vs `hello.jpg`), суффикс ` dB` у `replaygain_*`, лишние пробелы;
* `ref-only=[Rating]` — поля, которые foobar2000 хранит только в своей
  медиатеке (`Rating`, иногда обложка), в аудиофайле их нет. При обычной
  записи их сохраняет `merge_existing`, но `make compat` сравнивает именно
  извлечение из файла, поэтому показывает их как `ref-only`;
* `gen-only=[Cover Art (Front)]` — наша обложка: плагин её в `.tag` не кладёт,
  а утилита кладёт (ради этого всё и затевалось).

### Тестовые фикстуры

В `testdata/formats/` лежат крошечные файлы почти всех форматов (генерируются
через ffmpeg, в среднем 2–18 КБ):

```bash
make fixtures            # добавить отсутствующие файлы (нужен ffmpeg)
make fixtures FORCE=1    # перегенерировать все (файлы контейнеров недетерминированы)
```

Там же `cover.jpg`, `folder.jpg`, `poster.jpg` и `audio-attachments.mkv` с
двумя вложениями — на нём проверяется приоритет имён из `cover.names`
(`folder` должен победить `poster`).

Покрыты: MP3, MP2, AAC (raw), M4A/AAC, M4A/ALAC, FLAC, Ogg Vorbis, Opus,
WavPack, WMA, WMV, MKV, WebM, WAV. Отдельные случаи:

* `audio-track.mkv` и `audio-both.mkv` — track-level теги Matroska
  (`Artist` с уровня трека и `Album artist` с уровня альбома);
* `audio-cover.png.mkv` — вложение `box.png` (проверка имён и png);
* `audio-unicode.flac` — UTF-8 кириллица в тегах;
* `audio-empty.wav` — файл без тегов (ветка `no-tags`).

Для форматов из TODO (AC3, DTS, AU, Wave64, RF64, TS, Ogg FLAC, Speex)
файлы тоже есть — пригодятся, когда появится ридер.

Тест `internal/extractor/fixtures_test.go` прогоняет фикстуры через
`extractor + rules` и проверяет `Title`, `Artist`, `Track` и наличие обложки.

Сгенерированные рядом с фикстурами `.tag`/`.tag.bak` — это выход утилиты, они
не коммитятся (в `.gitignore`) и удаляются `make clean`.

Эталоны от foobar2000 лежат отдельно, в `testdata/reference/` (в git), чтобы
их не перетирали запуски `ext-tags`. Как их получить: открыть файлы из
`testdata/formats` в foobar2000 (плагин `foo_external_tags` в режиме
«tag files»), выполнить Tagging → Rewrite file tags и перенести появившиеся
`.tag` в `testdata/reference/` с теми же именами. Проверка:

```bash
make compat-fixtures
```

Сравнение с эталонами идёт по тем же правилам, что и `make compat`:
`ref-only` (Rating из медиатеки) и `cosmetic` (пробелы, ` dB`, имя обложки)
ошибкой не считаются.
**Ручная проверка в foobar2000:** откройте `testdata/formats/*` в плеере,
сгенерируйте `.tag` (`ext-tags testdata/formats/audio.mp3`) и сравните
Properties: должны совпадать Title/Artist/Album/Genre/Year/Track = 3, а для
MP3, MP2, FLAC, M4A, MKV — и обложка.
### Ручная: 5 минут в foobar2000

1. Preferences → Components → `foo_external_tags`: включить чтение внешних
   тегов и запись «in .tag files».
2. Сгенерировать `.tag` для тестового трека (`ext-tags track.flac`).
3. Добавить трек в foobar2000 и открыть Properties — поля и обложка должны
   прийти из `.tag`.
4. Изменить любой тег в foobar2000 — плагин перезапишет наш `.tag`
   (проверка, что формат ему «родной»).
5. Удалить `.tag` — foobar2000 покажет теги из самого файла.

Полностью автоматизировать GUI плеера можно, но хрупко: foobar2000 v2
поддерживает `/context_command:"Tagging/Rewrite file tags"` и портативный
профиль, так что при желании несложно собрать отдельный smoke-скрипт.
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
| `make compat` | сверка `.tag` с эталонами `foo_external_tags` из `bin/examples` |
| `make clean` | удаляет артефакты сборки (`bin/ext-tags*`, `dist/`, coverage) |

GitHub Actions (`.github/workflows/ci.yml`) на каждый push и pull request запускает `make tidy-check`, `make check`, `make build` и `make dist`.

Версии, теги и порядок релиза описаны в разделе «Версии и changelog» выше.
