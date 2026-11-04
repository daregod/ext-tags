# Changelog

Все заметные изменения проекта. Формат основан на
[Keep a Changelog](https://keepachangelog.com/ru/1.1.0/), версии — по
[семантическому версионированию](https://semver.org/lang/ru/).

## [0.7.0] — 2026-11-04

### Добавлено

- Обратный импорт переносит все пять artwork-слотов в ID3 APIC (front/back/
  artist/disc/icon) — проверено на MP3.
- Обложки: пять artwork-слотов foobar2000 — `front`/`back`/`artist`/`disc`/`icon`.
  Ключи APEv2: `Cover Art (Front/Back/Disc/Icon)` и `Artist Picture`; имена
  sidecar-файлов настраиваются отдельно (`cover.names`, `cover.back`, …).
- AC3 и DTS: чтение APEv2-тега из конца файла (как это делает foobar2000);
  для `.dts` расширение приоритетнее магии, потому что файл может быть Matroska.

### Изменено

- MPEG-TS окончательно помечен как неподдерживаемый: в контейнере нет
  тег-блока, и foobar2000 не читает даже заполненный внешний `.tag`.

## [0.6.0] — 2026-11-04

### Добавлено

- Лицензия MIT (`LICENSE`).

### Изменено

- README переписан под пользователей; архитектура, тесты, фикстуры и релизы
  вынесены в DEVELOPMENT.md.
- Модуль переименован в `github.com/daregod/ext-tags`, добавлен GitHub Actions
  workflow релиза: по тегу `v*` собираются zip/tar.gz для Windows/Linux/macOS
  и `SHA256SUMS`.

### Исправлено

- `make tag` создаёт тег через `cmd/devtool` (Go), без shell-кавычек:
  одинаково работает в `cmd.exe`, PowerShell и `sh`.
- Вспомогательные утилиты переписаны на Go: `cmd/builddate` (BUILD_DATE),
  `cmd/devtool` (mkdir/rm/rmglob для Makefile) и `cmd/genfixtures` (тестовые
  фикстуры). PowerShell-скрипты удалены: `go run` одинаково работает на
  Windows, Linux и macOS, Makefile закреплён как LF через `.gitattributes`.

## [0.5.0] — 2026-11-04

### Добавлено

- Обратный импорт тегов для MP3: `ext-tags import` читает `<file>.tag` и
  пишет ID3v2.2/2.3/2.4 (`--id3v2-version`, по умолчанию 3) и при желании
  ID3v1 (`--id3v1`): поля, обложка (PIC/APIC), неизвестные поля — TXX/TXXX;
  флаги `--replace`, `--backup`, `--dry-run`.
- Собственный writer ID3 (`internal/id3`) без внешних зависимостей; при
  импорте без `--replace` сохраняются существующие кадры той же версии.

### Изменено

- Обновлены зависимости и `go`-директива до 1.26; вендоренная `bogem/id3v2`
  удалена.
- Расширены дефолтные имена обложек: `cover`, `folder`, `front`, `album`,
  `poster`, `box`, `disc`, `media` (было пять имён).

## [0.4.0] — 2026-11-04

### Добавлено

- MP1: фикстура (получена `mp2enc -l 1`, теги — ID3v2) и тест.
- Wave64: ридер GUID-чанков (64-битные размеры) и фикстура с ID3-чанком
  перед `data` (именно так его читает foobar2000; собран через
  `flac --force-wave64-format`).
- CLI на `cobra`: команды `extract`, `config`, `version`; глобальные флаги
  `--config`, `--profile`, `--dry-run`, `--json`, `--quiet`, `-v`,
  `-r/--recursive`, `--on-existing`, `--cover`, `--merge`/`--no-merge`.
  Вызов `ext-tags track.flac` без подкоманды работает как раньше.
- YAML-конфигурация (`internal/config`): `files`, `tagfile`, `cover`,
  `fields` и именованные профили; команды `config init|validate|show|path`.
- Слои конфига: встроенные значения → пользовательский
  (`%APPDATA%\ext-tags\config.yaml`) → проектный (`./ext-tags.yaml`);
  явный `--config`/`EXT_TAGS_CONFIG` заменяет найденные файлы.
- `config init` пишет минимальный файл (частые ключи, остальное — дефолты),
  `--full` — полный дамп; `config path` показывает слои.
- Правила полей (`internal/rules`): `aliases`/`ignore`, `set` с Go-шаблоном,
  `copy`, `rename`, `default`, `remove`, условие `when`; функции `lower`,
  `upper`, `title`, `trim`, `default`, `replace`, `regexReplace`;
  `rules_mode: append` (по умолчанию) и `replace`.
- Режим слияния `files.merge_existing` (включён по умолчанию): при перезаписи
  `.tag` сохраняются поля, которых нет в аудио (`Rating`, обложка из
  медиатеки); предупреждение для `skip`/`error`.
- Политика `on_existing` (`overwrite`, `skip`, `backup`, `error`);
  по умолчанию старый `.tag` сохраняется как `.tag.bak`.
- Машинный вывод `--json` (NDJSON, по объекту на файл) и режим `--dry-run`.
- Форматы: WebM/MKV (Matroska), APE, WavPack, ASF/WMV/WMA, ALAC, raw AAC,
  MP2, MP1, WAV (LIST/INFO и ID3v2, в том числе после RIFF), AIFF (собственный
  IFF-ридер), RF64, Musepack, TAK, Ogg FLAC, Speex, Wave64.
- Чтение APEv2 (`ape.Parse`/`ReadFile`, в том числе тег в конце файла).
- Обложки: имена `cover.names` (`cover`, `folder`, `box`, `disc`, `media`),
  поддержка webp/gif, импорт из папки аудиофайла (`cover.directory`);
  встроенная обложка приоритетнее.
- Тестовые фикстуры `testdata/formats/`, эталоны foobar2000
  `testdata/reference/`, цели `make fixtures`, `make compat`,
  `make compat-fixtures`.
- Юнит-тесты конфигурации, движка правил, записи `.tag`, CLI и парсеров
  (Matroska, ASF, RIFF/WAV, AIFF, Ogg, APEv2), ID3v1 и ID3v2.4.

### Изменено

- Хардкод маппинга ключей, обрезки технических полей и сборки `Track`/`Disc`
  перенесён в конфиг (встроенные значения повторяют прежнее поведение).
- `extractor` больше не знает о правилах: отдаёт сырые поля контейнера,
  обложку и нормализованные номера трека/диска.
- Matroska: учитываются `TargetTypeValue`/`TargetTrackUID` — альбомные теги
  мапятся как в foobar2000 (`ARTIST` → `Album artist`, `PART_NUMBER` →
  `Disc`), кодечные `ENCODER` и `DURATION` отбрасываются.
- `make compat` собирает `bin/compat.exe`; `make clean` удаляет артефакты,
  включая сгенерированные `.tag`/`.tag.bak` фикстур.
- Добавлены зависимости `github.com/spf13/cobra` и `gopkg.in/yaml.v3`.

### Исправлено

- ID3v1: нулевой `Track` больше не попадает в `.tag`.
- Свои `.tag`/`.tag.bak` не попадают в обработку, если подхватились маской
  (`ext-tags audio.*`).

## [0.3.0] — 2026-11-04

### Добавлено

- Поля iTunes `Itunescompilation` (`cpil`) и `Tempo` (`tmpo`) в `.tag`.
- Юнит-тесты на нормализацию ключей (`mapKey`) и объединение номеров
  трека/диска.

### Исправлено

- M4A/MP4: убраны дубликаты полей и «мусорные» атомы (`©alb`, `©ART`,
  `©cmt`) с невалидным UTF-8 в именах; служебные атомы (`iTunNORM`,
  `iTunSMPB`, `Encoding Params`, `©too`) больше не попадают в `.tag`.
- OGG/Opus: служебное поле `vendor` не переносится в теги.
- FLAC/OGG/Opus: убран ложный `Composer`, который библиотека подставляла
  из `ARTIST`, если тега COMPOSER в файле нет.
- WAV и другие файлы без тегов завершаются как `SKIP`, а не как ошибка.

## [0.2.0] — 2026-11-03

### Добавлено

- `Makefile`: сборка, тесты, покрытие, `vet`, `lint`, форматирование,
  кросс-компиляция в `dist/`, очистка артефактов.
- Конфигурация `golangci-lint` (`.golangci.yml`) и `.editorconfig`.
- CI (`.github/workflows/ci.yml`): `make tidy-check`, `check`, `build`, `dist`.
- Флаг `--version`; версия, коммит и дата передаются через `-ldflags`.

### Изменено

- Сборка и проверки выполняются через `make`, README дополнен разделом
  про автоматизацию.
- Импорты сгруппированы по `goimports` (локальные пакеты отдельно).


## [0.1.0] — 2026-11-03

### Добавлено

- CLI `ext-tags`: извлечение тегов из MP3 (ID3v1, ID3v2.2–2.4), FLAC, Opus,
  OGG Vorbis, M4A/MP4, WAV, AIFF.
- Запись `.tag` в APEv2 с заголовком `TAGFILE`, совместимая с плагином
  `foo_external_tags` для foobar2000.
- Извлечение и упаковка обложки `Cover Art (Front)`.
- Нормализация ключей (`ORGANIZATION` → `Publisher`, `DATE` → `Year` и т. п.).
- Объединение `TRACKNUMBER`/`TRACKTOTAL` и `DISCNUMBER`/`DISCTOTAL`
  в значения вида `02/16`.
- Структура проекта: `cmd/ext-tags` + `internal/{ape,extractor}`, README.
