# Changelog

Все заметные изменения проекта. Формат основан на
[Keep a Changelog](https://keepachangelog.com/ru/1.1.0/), версии — по
[семантическому версионированию](https://semver.org/lang/ru/).

## [Unreleased]

### Добавлено

- Пакет `internal/version`: версия, коммит и дата сборки передаются через
  `-ldflags`, вывод `ext-tags --version` стал единым для всех сборок.
- `CHANGELOG.md` и цель `make tag TAG=vX.Y.Z` для выпуска версий.

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
