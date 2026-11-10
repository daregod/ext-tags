# Changelog

Все заметные изменения проекта. Формат основан на
[Keep a Changelog](https://keepachangelog.com/ru/1.1.0/), версии — по
[семантическому версионированию](https://semver.org/lang/ru/).

## [0.16.4] — 2026-11-10

### Добавлено

- `artwork list --all` и `artwork extract --ignore-limit` для обложек больше `cover.hard_max_bytes`.
- Предупреждение `artwork list` о кандидате, отброшенном по `hard_max_bytes`.

## [0.16.3] — 2026-11-10

### Исправлено

- `artwork`/`render` используют `Document.Cover` как fallback для слота `front`, если читатель не заполнил `Artwork`.

## [0.16.2] — 2026-11-10

### Исправлено

- `album.yaml` применяется к аудиофайлам без встроенных тегов: такой файл больше не пропускается как `no-tags`, если album-поля дают непустой `.tag`.

## [0.16.1] — 2026-11-10

### Добавлено

- `render --raw` отдаёт `.artwork`, `.containerartwork`, `.tagartwork`, их `...count` и `...<slot>`/`...<slot>bytes`.

## [0.16.0] — 2026-11-10

### Добавлено

- `render` отдаёт в шаблон данные об обложках: `.artwork`, `.artworkcount` и `.artwork<slot>`/`.artwork<slot>bytes`.
- `render --raw` отдаёт сырые поля контейнера и `.tag` через `.fields`, `.raw` (alias), `.container`, `.tag` и `.vars` без album/rules.

### Изменено

- `render` всегда отдаёт `.fields`: финальные поля в обычном режиме и сырые merged-поля в `--raw`.

## [0.15.1] — 2026-11-09

### Добавлено

- `album_init.multiline`: `block` (по умолчанию) или `quoted` для редакторов, которые не понимают YAML block scalar.
- `album_init.quote`: `minimal` (по умолчанию) или `all` — двойные кавычки для ключей и значений, чтобы редакторы не путались на пробелах/апострофах.

## [0.15.0] — 2026-11-09

### Добавлено

- Секция `album_init`: `ignore`, `album_order`, `track_order` для генерации `album.yaml`; числовая сортировка ключей треков.

### Изменено

- `album init` записывает многострочные поля (`Www`, `track_url`, …) YAML-блоками `|`, нормализует переводы строк и обрезает хвостовые пробелы.

## [0.14.1] — 2026-11-09

### Исправлено

- Многострочные поля (включая `track_url`) нормализуются в `\r\n`; хвостовые пробелы и табы в строках обрезаются.

## [0.14.0] — 2026-11-09

### Добавлено

- Команда `ext-tags artwork list` и `ext-tags artwork extract` для просмотра и сохранения выбранной обложки.

## [0.13.0] — 2026-11-09

### Добавлено

- Команда `ext-tags render` для вывода на экран финальных полей по Go-шаблону.
- Шаблонная функция `sanitize` для безопасных имён файлов.

## [0.12.5] — 2026-11-09

### Исправлено

- Ключи `tracks` в `album.yaml` находятся по числовому префиксу имени файла (`01. Artist - Track`); `01` и `1` считаются одним треком.
- `album init` форматирует `Track` с ведущими нулями по ширине общего числа (`01/12`).

## [0.12.4] — 2026-11-08

### Исправлено

- Поля из `album.yaml` перекрывают встроенные алиасы и существующий `.tag` (например, `Album artist`).

## [0.12.3] — 2026-11-08

### Исправлено

- `album init` подмешивает поля из существующего `.tag`, как `extract`; `--no-merge` отключает это.
- `album init --merge-album` дополняет существующий `album.yaml`, сохраняя ручные значения.

## [0.12.2] — 2026-11-08

### Изменено

- `album.yaml`: явное пустое значение (`Comment: ""`) теперь удаляет поле; отсутствующий ключ по-прежнему ничего не меняет.

## [0.12.1] — 2026-11-06

### Исправлено

- Правила видят поля из существующего `.tag`: `Track`/`Disc` восстанавливают `.tracknumber`/`.totaltracks`/`.discnumber`/`.disctotal`, а `Releasetime` из сайдкара больше не перезаписывается годом.

### Добавлено

- Шаблонный helper `atoi` для числовых условий, например `gt (atoi .disctotal) 1`.

## [0.12.0] — 2026-11-06

### Добавлено

- `extract --output PATH` и `extract --stdout`; существующий `.tag` читается рядом с исходным аудиофайлом.
- `fields.include`/`exclude` применяются к итоговому набору после merge и правил; обложки получили имена для фильтров.
- Профили ищутся в обычных слоях конфига без обязательного `--config`.
- `import --verify=error|warn|off`; в `warn` импорт продолжается с предупреждением `SKIP field`.
- `TAGS.md` и справочник полей в комментариях `config init --full`; описан механизм пользовательских полей.
- Уточнена разница между встроенным шаблоном `Track` и `formatTrack`.
- Добавлены `cover.max_bytes` и `cover.hard_max_bytes` с fallback по размеру; уточнено, что бюджет не меняет порядок `cover.names`.

### Изменено

- Импорт сообщает контейнер в ошибке проверки.

## [0.11.3] — 2026-11-05

### Исправлено

- `self-update` дополнительно определяет opkg по базе `/opt/lib/opkg/info/ext-tags.list` и по путям Entware любого варианта.

## [0.11.2] — 2026-11-05

### Исправлено

- `self-update` распознаёт Entware/opkg, даже если `/opt/bin` разрешается в реальный путь вида `/share/.../Entware/bin`.

## [0.11.1] — 2026-11-05

### Исправлено

- Release workflow обновляет Scoop/Homebrew по хешам release-ассетов, а не `dist/`.

## [0.11.0] — 2026-11-05

### Добавлено

- `ext-tags self-update` для standalone-бинарников: `--check`, `--dry-run`, `--force`; проверка SHA256 и атомарная замена.
- Для установок через opkg/apt/Homebrew/Scoop self-update подсказывает правильную команду пакетного менеджера.

### Изменено

- Release workflow обновляет Scoop bucket и Homebrew tap вместе с opkg-feed.

## [0.10.3] — 2026-11-05

### Добавлено

- Entware/opkg feed для QNAP: `opkg install ext-tags` и `opkg upgrade ext-tags`.
- Release workflow собирает `.ipk` и автоматически обновляет `opkg-feed`.

## [0.10.2] — 2026-11-05

### Добавлено

- Linux-пакеты `.deb`, `.rpm` и `.apk` в релизах.
- Scoop bucket, Homebrew tap и AUR `PKGBUILD`.
- `INSTALLING.md` с вариантами установки.

## [0.10.1] — 2026-11-05

### Исправлено

- `album init` заполняет общее количество треков в `Track` (`1/1`, `1/8`).
- `Www` в `.tag` записывается с CRLF, чтобы многострочные ссылки одинаково
  отображались в foobar2000.

## [0.10.0] — 2026-11-05

### Добавлено

- `ext-tags album init`: генерация `album.yaml` из тегов треков.

## [0.9.3] — 2026-11-05

### Исправлено

- YAML-конфиги (`ext-tags.yaml`, `album.yaml`) принимают UTF-8 BOM и CRLF.

## [0.9.2] — 2026-11-05

### Исправлено

- Linux-бинарники релиза собираются со `CGO_ENABLED=0`; они больше не
  требуют GLIBC 2.32/2.34 и запускаются на старых NAS (QNAP и т.п.).

## [0.9.1] — 2026-11-05

### Добавлено

- Per-album `album.yaml`: переопределение полей альбома, треков и artwork.
- `cover.prefer`: `embedded`, `sidecar` или `existing`.
- Шаблонные функции `year` и `formatTrack`.

### Исправлено

- Примеры README приведены к реальным тегам альбома: без вымышленного
  `Remixer`, с правильным глобальным разделителем артистов `&` и
  per-track жанром для трека 8.

## [0.9.0] — 2026-11-04

### Добавлено

- Обратный импорт для WAV/RF64/Wave64/AIFF через ID3v2-чанк; старые
  LIST/INFO и текстовые IFF-чанки заменяются.
- `import` пишет атомарно: временный файл → проверка чтением → замена
  оригинала; `--backup` создаётся до записи.

### Изменено

- Форматные реализации вынесены в `internal/formats`, контракты чтения и
  записи — в `internal/tagformat`; compatibility-facade `internal/extractor`
  удалён. Пользовательские команды и формат `.tag` не изменились.

## [0.8.0] — 2026-11-04

### Добавлено

- Обратный импорт ASF/WMA/WMV: Content Description, Extended Content
  Description и `WM/Picture` (пять artwork-типов, лимит 64 КиБ на descriptor).
- Обратный импорт Matroska/WebM: `Tags`/`SimpleTag`, `Attachments` с пятью
  artwork-слотами, обновление `Cues`, без перепаковки кластеров; необязательный
  `SeekHead` удаляется, чтобы не оставалось устаревших позиций.
- Обратный импорт MP4/M4A: iTunes-атомы (`©nam`, `©ART`, `aART`, `trkn`,
  `covr`, …), пересчёт `stco`/`co64` при сдвиге `moov`; front cover.
- Обратный импорт APEv2: APE, WavPack, Musepack, TAK, AC3/DTS — в конец
  файла пишется тег с пятью artwork-ключами, аудио не меняется.
- `fixext`: переименование Ogg-файлов с неверным расширением (например,
  Opus в `.flac`) в `.opus`/`.ogg`/`.oga`/`.spx`; dry-run по умолчанию,
  сопутствующий `.tag` переименовывается вместе с файлом.
- Обратный импорт для Ogg FLAC: FLAC-блоки внутри Ogg, счётчик header-пакетов
  и перепаковка Ogg-страниц.
- Обратный импорт для Ogg Vorbis/Opus/Speex: Vorbis comments и
  `METADATA_BLOCK_PICTURE`, с перепаковкой Ogg-страниц и CRC; аудио не
  перекодируется.
- Обратный импорт для FLAC: Vorbis comments + PICTURE-блоки (пять artwork-
  слотов), аудиокадры не меняются.

### Исправлено

- ASF: reserved-байты заголовка `01 02`, правильные типы Extended Content
  (строки — type 0), `date` и `WM/Year` пишутся оба; WMF/foobar2000 снова открывает
  metadata editor.
- Ogg: непрерывные sequence-номера страниц и корректные continuation-флаги
  при больших comment-пакетах.
- Ogg/Opus с расширением `.flac`: контейнер определяется по магии; при импорте
  выводится предупреждение, что foobar2000 показывает artwork только для `.opus`.
- MP4: пересчёт `stco`/`co64` при сдвиге `moov`, включая faststart-файлы.

## [0.7.0] — 2026-11-04

### Добавлено

- Обратный импорт переносит все пять artwork-слотов в ID3 APIC (front/back/
  artist/disc/icon) — проверено на MP3.
- Обложки: пять artwork-слотов foobar2000 — `front`/`back`/`artist`/`disc`/`icon`.
  Ключи APEv2: `Cover Art (Front/Back/Disc/Icon)` и `Artist Picture`; имена
  sidecar-изображений настраиваются отдельно (`cover.names`, `cover.back`, …).
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
