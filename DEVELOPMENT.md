# Разработка ext-tags

Документ для тех, кто меняет код. Пользовательская документация — в
[README.md](README.md).

## Требования

- Go 1.26+;
- GNU Make 4.x (необязательно, но команды ниже рассчитаны на него);
- `golangci-lint` v2 — `make setup` поставит при необходимости;
- для перегенерации фикстур: `ffmpeg` с libmp3lame/libvorbis/libopus/libspeex/
  wavpack, по возможности `flac` и `mp2enc` (mjpegtools).

## Быстрый старт

```bash
make setup      # проверка тулчейна, go mod download, линтер
make check      # формат, tidy, vet, lint, тесты
make build      # bin/ext-tags или bin/ext-tags.exe
make run ARGS="--dry-run testdata/formats/audio.mp3"
make ci         # tidy-check + check
```

Полезные цели: `test`, `test-race`, `cover`, `cover-html`, `bench`, `vet`,
`lint`, `lint-fix`, `fmt`, `fmt-check`, `tidy`, `tidy-check`, `dist`, `clean`,
`version`, `tag`, `fixtures`, `compat`, `compat-fixtures`.

`make dist` собирает `dist/ext-tags_{windows,linux,darwin}_{amd64,arm64}`.
Версия, коммит и дата подставляются через `-ldflags`; VCS-стампинг отключён
(`-buildvcs=false`) — проект может лежать на сетевой шаре.

## Структура

```
cmd/ext-tags        CLI (тонкая обёртка над internal/cli)
cmd/compat          сверка .tag с эталонами foobar2000
cmd/builddate       дата последнего коммита для -ldflags
cmd/devtool         mkdir/rmdir/rm/rmglob для Makefile
cmd/genfixtures     генератор testdata/formats
internal/cli        команды cobra, вывод text/json, импорт
internal/config     YAML-схема, слои, профили, валидация
internal/metadata   Document/Cover/ArtworkKind — общие типы
internal/tagformat  Reader/Writer + Registry, явный tagformat.New
internal/registry   NewDefault: явный список reader/writer по форматам
internal/read       общие read-хелперы + dhowden/tag adapter
internal/rules      aliases/ignore/set/copy/rename/default/remove
internal/tagfile    запись .tag: merge, backup, dry-run
internal/importer   .tag → формат (ID3, APEv2, FLAC, Ogg, MP4…)
internal/formats/id3        свой writer ID3v1/2.2/2.3/2.4
internal/formats/ape        APEv2 (чтение и запись)
internal/formats/{asf,matroska,ogg,riff,flac,mp4,vorbis}  форматы/кодеки
internal/artwork    поиск обложки рядом с аудио
internal/version    версия, коммит, дата
```

Поток данных:

```
cli → registry.NewDefault → tagformat.Reader → формат-пакет/read.Taglib
                                             → metadata.Document
                                             → rules → tagfile

cli → registry.NewDefault → tagformat.Writer → importer → формат
```

`extract` ничего не пишет в аудио. `import` — единственная команда, которая
меняет аудиофайл. Реестр создаётся явно в `cmd/ext-tags`/`cli.Execute` и
передаётся в `app`; глобального `Default`/`init()` нет.

## Вспомогательные утилиты на Go

Служебные операции вынесены в отдельные Go-команды; Makefile запускает их
через `go run`. Так сборка, тесты и очистка одинаково работают в `cmd.exe` и
`sh`, без внешних скриптов и платформенных утилит:

| Команда | Makefile |
|---|---|
| `go run ./cmd/builddate` | `BUILD_DATE` для `-ldflags` |
| `go run ./cmd/devtool mkdir/rm/rmglob` | `mkdir`, `clean` |
| `go run ./cmd/genfixtures` | `make fixtures` |

Новую служебную логику лучше добавлять сюда же, а не в рецепты Makefile.

## Тесты

```bash
make test          # go test ./...
make test-race     # -race (где поддерживается)
make check         # fmt-check + tidy-check + vet + lint + test
```

Юнит-тесты лежат рядом с кодом (`*_test.go`). Внешние сервисы не нужны.
Для парсеров используются синтетические данные и/или файлы из
`testdata/formats`.

## Фикстуры

`testdata/formats/` — маленькие файлы всех поддерживаемых форматов, они
закоммичены. Генератор:

```bash
make fixtures          # только отсутствующие файлы
make fixtures FORCE=1  # перегенерировать всё
```

Внутри `cmd/genfixtures`:

- ffmpeg создаёт аудио/видео и картинки;
- для MP2/AAC/MP1 сверху приклеивается ID3v2 от `audio.mp3`;
- для `audio.mp1` используется `mp2enc -l 1` (если найден);
- для `audio.w64` — `flac --force-wave64-format`, затем ID3-чанк вставляется
  **до `data`**: именно так его читает foobar2000;
- `.tag`/`.tag.bak` в `testdata/formats` не коммитятся и удаляются
  `make clean`.

Сгенерированные файлы контейнеров могут быть недетерминированными
(Matroska SegmentUID, Ogg serial), поэтому инкрементальный режим важен.
`FORCE=1` используй осознанно — будет шум в diff.

## Сверка с foobar2000

`testdata/reference/` — `.tag`, созданные самим foobar2000 для файлов из
`testdata/formats`. Они коммитятся.

```bash
make compat-fixtures   # сравнить testdata/formats с testdata/reference
make compat            # сравнить с bin/examples (локальные примеры, не в git)
```

Статусы `compat`:

- `OK` — совпало;
- `OK ... cosmetic` — допустимые косметические отличия;
- `DIFF`/`changed` — реальное расхождение, CI падает;
- `ref-only`/`gen-only` — поле только с одной стороны (разрешено, печатается).

Косметика: NUL-разделённые дубли значений (APEv2/ASF-стиль foobar2000),
суффикс ` dB` в ReplayGain, имя файла обложки, лишние пробелы.

Как обновить эталон:

1. `make fixtures` (или `FORCE=1`);
2. добавить файлы в foobar2000, `Tagging → Rewrite file tags`;
3. получившиеся `.tag` перенести в `testdata/reference/` с тем же именем;
4. `make compat-fixtures` — должно быть 0 failed.

## Добавление нового формата

1. Написать/расширить парсер в `internal/formats/<format>`.
2. Смапить поля в `internal/formats/<format>/read.go` или `read.Taglib`;
   технические поля — в `ignore` конфига, если они не нужны.
3. Зарегистрировать reader в `internal/registry` (имя + расширения).
4. Для записи: `importer.Import...` + writerSpec в `internal/registry`.
5. Добавить файл в `cmd/genfixtures` и сгенерировать его.
6. Добавить кейс в `internal/extractor/fixtures_test.go` и round-trip тест.
7. Снять эталон в foobar2000 → `testdata/reference`.
8. Обновить TODO/README (форматы) и CHANGELOG.

Если формат без тегов (AC3/DTS/AU/TS), он остаётся в TODO как «вне области».

## Правила и конфиг

- Встроенные правила и алиасы — в `internal/config` (дефолты).
- Пользовательские слои накапливаются: `rules_mode: append` (по умолчанию)
  или `replace`.
- Шаблоны — `text/template` с `missingkey=zero`; данные — поля в нижнем
  регистре без пробелов (`Album artist` → `.albumartist`) плюс
  `.tracknumber`, `.totaltracks`, `.discnumber`, `.disctotal`.
- Для правил есть тесты в `internal/rules`.

## ID3 и импорт

`internal/id3` — собственный writer, без внешних зависимостей: ID3v1 и
ID3v2.2/2.3/2.4, пять APIC-типов (front/back/artist/disc/icon), TXXX,
сохранение существующих кадров при `--replace=false`. `internal/importer` читает `.tag` (APEv2) и вызывает
writer. При добавлении полей не забудь тесты на все версии.

## Релизы

1. Перенести `[Unreleased]` в `[X.Y.Z] — YYYY-MM-DD` в `CHANGELOG.md`.
2. Убедиться, что `make ci` и `make compat-fixtures` зелёные.
3. Закоммитить.
4. Поставить аннотированный тег (требует чистого дерева):

```bash
make tag TAG=vX.Y.Z
```

5. Отправить: `git push && git push --tags`.

После пуша тега `v*` workflow `.github/workflows/release.yml` собирает
`make dist`, упаковывает бинарники в zip/tar.gz, считает `SHA256SUMS` и
создаёт GitHub Release. `make version` и `-ldflags` берут версию из
`git describe`.

## Кроссплатформенность

- Makefile: `SHELL=cmd.exe` на Windows, `sh` на Unix; файловые операции —
  `cmd/devtool`.
- Не используй `%`-форматы в shell-командах: разные make/шеллы обрабатывают
  `%%` по-разному. Дата собирается `cmd/builddate`.
- `.gitattributes` закрепляет `Makefile` как LF; при `core.autocrlf=true` это
  важно.
- На сетевых шарах git требует `safe.directory=*`; Makefile добавляет
  `-c safe.directory=*` на Windows, Go-код — там, где вызывает git.
- Go-код не должен зависеть от регистра файловой системы и `\`-путей:
  используй `filepath`.

## Стиль

- `gofmt`/`goimports`, локальные пакеты отдельной группой.
- Ошибки — с контекстом (`fmt.Errorf("...: %w", err)`), без паник в
  библиотечном коде.
- Новое поведение — с тестом; пользовательские изменения — в CHANGELOG.
## Обложки

Пять слотов foobar2000 хранятся в `.tag` под фиксированными ключами APEv2:

| Слот | Ключ | Sidecar по умолчанию |
|---|---|---|
| front | `Cover Art (Front)` | `cover`, `folder`, `front`, `album`, `poster`, `box` |
| back | `Cover Art (Back)` | `back` |
| artist | `Artist Picture` | `artist` |
| disc | `Cover Art (Disc)` | `disc`, `media` |
| icon | `Cover Art (Icon)` | `icon` |

`extractor.Result.Artwork` — map `front/back/artist/disc/icon` → `*Cover`.
Встроенные картинки сейчас дают только front; остальные слоты заполняются
sidecar-файлами (`cover.back`, `cover.artist`, …). `tagfile.Build` принимает
этот map и пишет/сохраняет соответствующие APEv2-элементы.
