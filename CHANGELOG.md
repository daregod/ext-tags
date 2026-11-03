# Changelog

Все заметные изменения проекта. Формат основан на
[Keep a Changelog](https://keepachangelog.com/ru/1.1.0/), версии — по
[семантическому версионированию](https://semver.org/lang/ru/).

## [Unreleased]

### Добавлено

- CLI на `cobra`: команды `extract`, `config`, `version`; глобальные флаги
  `--config`, `--profile`, `--dry-run`, `--json`, `--quiet`, `-v`.
  Вызов `ext-tags track.flac` без подкоманды работает как раньше.
- YAML-конфигурация (`internal/config`): `files`, `tagfile`, `cover`,
  `fields` и именованные профили; команды `config init|validate|show|path`.
- Конфиг применяется слоями: встроенные значения → пользовательский
  (`%APPDATA%\ext-tags\config.yaml`) → проектный (`./ext-tags.yaml`);
  явный `--config`/`EXT_TAGS_CONFIG` заменяет найденные файлы.
- Режим слияния `files.merge_existing` (по умолчанию включён): при перезаписи
  `.tag` сохраняются поля, которых нет в аудио (`Rating`, обложка из
  медиатеки); флаги `--merge`/`--no-merge`, предупреждение для `skip`/`error`.
- `config init` пишет минимальный файл (частые ключи, остальное — дефолты),
  `config init --full` — полный дамп; `config path` показывает слои.
- Правила из конфига накапливаются поверх встроенных (`rules_mode: append`
  по умолчанию, `replace` — полная замена), поэтому `Track`/`Disc`
  продолжают собираться при добавлении своих правил.
- Правила полей (`internal/rules`): `set` с Go-шаблоном, `copy`, `rename`,
  `default`, `remove`, условие `when`; функции `lower`, `upper`, `title`,
  `trim`, `default`, `replace`, `regexReplace`.
- Машинный вывод `--json` (NDJSON, по объекту на файл) и режим `--dry-run`.
- Политика `on_existing` (`overwrite`, `skip`, `backup`, `error`);
  по умолчанию старый `.tag` сохраняется как `.tag.bak`.
- Юнит-тесты конфигурации, движка правил, записи `.tag` и CLI.
- Чтение APEv2 (`ape.Parse`/`ReadFile`) и команда `make compat`: сравнение
  наших `.tag` с эталонами `foo_external_tags` (с учётом допустимых
  косметических отличий).

### Изменено

- Хардкод маппинга ключей, обрезки технических полей и сборки `Track`/`Disc`
  перенесён в конфиг (встроенные значения повторяют прежнее поведение).
- `extractor` больше не знает о правилах: он отдаёт сырые поля контейнера
  и нормализованные номера трека/диска.
- Добавлены зависимости `github.com/spf13/cobra` и `gopkg.in/yaml.v3`.

## [0.3.0] — 2026-11-04

### Добавлено

- Поддержка Monkey Audio (`.ape`): чтение APEv2-тега из конца файла,
  обложка `Cover Art (Front)`, номера трека/диска.
- Поддержка ASF (`.wmv`, `.wma`, `.asf`): Content Description
  (Title/Author/Comment/Rating) и Extended Content Description
  (`WM/AlbumTitle`, `WM/Genre`, `WM/TrackNumber`, `WM/Picture` как обложка);
  технические дескрипторы (`WMFSDK*`, `IsVBR`, …) игнорируются.
- Поддержка Matroska/WebM (`.webm`, `.mkv`): `Tags/SimpleTag`, `Info/Title`,
  вложения-картинки как обложка, `PART_NUMBER`/`TOTAL_PARTS` → `Track`;
  пакет `internal/matroska`, `Cluster`/`Cues` пропускаются без чтения.
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
