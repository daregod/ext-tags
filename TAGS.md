# Справочник полей ext-tags

Имена полей сравниваются без учёта регистра. В `.tag` и в `include`/`exclude`
лучше использовать каноническое написание из таблицы — так результат
предсказуемее.

Незнакомые поля не теряются: если поле не входит в `fields.ignore`, оно
попадает в `.tag` под исходным именем. `fields.aliases` может добавить свои
соответствия.

## Свои поля

Список полей не является закрытым. Любое незнакомое поле из контейнера
попадает в `.tag` под своим именем, если его не отсекает `fields.ignore`.
Это же имя можно использовать в `include`, `exclude`, правилах и `album.yaml`.

Пример: в файлах есть тег `CUSTOM_TAG`, а в библиотеке нужно поле
`My field`:

```yaml
fields:
  aliases:
    CUSTOM_TAG: My field
  include: [Artist, Title, "My field"]
```

Если алиас не нужен, поле можно оставить как есть:

```yaml
fields:
  include: [Artist, Title, CUSTOM_TAG]
```

Или добавить значение только для одного альбома через `album.yaml`:

```yaml
album:
  My field: значение из буклета
```

В `include`/`exclude` имена сравниваются без учёта регистра; каноническое
написание из таблицы ниже — только рекомендация для предсказуемого вывода.

## Канонические поля

| Поле | Исходные имена (примеры) | Примечание |
|---|---|---|
| `Album` | `TALB`, `ALBUM`, `©alb` | |
| `Album artist` | `TPE2`, `ALBUMARTIST`, `ALBUM ARTIST`, `aART` | именно с пробелом |
| `Artist` | `TPE1`, `ARTIST`, `©ART` | |
| `BPM` | `TBPM` | |
| `Catalognumber` | `CATALOGNUMBER`, `CATALOG NUMBER` | без пробела в каноническом имени |
| `Comment` | `COMM`, `COMMENT`, `©cmt` | |
| `Comment catalog number` | `COMMENT CATALOG NUMBER` | |
| `Composer` | `TCOM`, `COMPOSER`, `©wrt` | |
| `Copyright` | `TCOP`, `COPYRIGHT` | |
| `Disc` | `DISC` | обычно генерируется правилом `Disc` |
| `Encoded by` | `TENC` | |
| `Encoder` | `TSSE` | |
| `Genre` | `TCON`, `GENRE`, `©gen` | |
| `Isrc` | `ISRC` | |
| `Itunescompilation` | `cpil` | |
| `Key` | `TKEY` | |
| `Label` | `LABEL` | |
| `Language` | `TLAN` | |
| `Lyricist` | `TEXT` | |
| `Media` | `TMED` | |
| `Musicbrainz_albumid` | `MUSICBRAINZ_ALBUMID` | |
| `Musicbrainz_trackid` | `MUSICBRAINZ_TRACKID` | |
| `Performer` | `PERFORMER` | |
| `Publisher` | `TPUB`, `PUBLISHER`, `ORGANIZATION` | |
| `Rating` | `RATING` | |
| `Releasetime` | `RELEASETIME` | |
| `replaygain_track_gain` | `REPLAYGAIN_TRACK_GAIN` | нижний регистр |
| `replaygain_track_peak` | `REPLAYGAIN_TRACK_PEAK` | нижний регистр |
| `Tempo` | `tmpo` | |
| `Title` | `TIT2`, `TITLE`, `©nam` | |
| `Track` | `TRACK` | обычно генерируется правилом `Track` |
| `Upc` | `UPC` | |
| `Www` | `WWW` | многострочные значения нормализуются в `\r\n` |
| `Year` | `TDRC`, `TYER`, `DATE`, `YEAR`, `DATE_RELEASED`, `©day` | |

## Обложки

Имена для `fields.include`/`fields.exclude`:

| Слот | Имя |
|---|---|
| front | `cover.field` из конфига, по умолчанию `Cover Art (Front)` |
| back | `Cover Art (Back)` |
| artist | `Artist Picture` |
| disc | `Cover Art (Disc)` |
| icon | `Cover Art (Icon)` |

Если `fields.include` непустой, обложку нужно явно добавить в список, иначе
она будет отфильтрована. `exclude` убирает только перечисленные слоты.

## Генерируемые и служебные поля

- `Track` — встроенный шаблон `{{ .tracknumber }}{{ if .totaltracks }}/{{ .totaltracks }}{{ end }}`; ведущие нули не добавляет.
- `Disc` — аналогично для `discnumber`/`disctotal`.
- `formatTrack` — отдельная функция для **готового** `X/Y`: `1/12` → `01/12`, `1/8` → `1/8`; не читает `.tracknumber`/`.totaltracks`. Подробнее в [README.md](README.md#нумерация-треков).
- Служебные ключи APEv2 для обложек: `Cover Art (Front)`, `Cover Art (Back)`,
  `Artist Picture`, `Cover Art (Disc)`, `Cover Art (Icon)`.

## Игнорируемые по умолчанию

`vendor`, `iTunNORM`, `iTunSMPB`, `Encoding Params`, `covr`, `©too`,
`trkn`, `trkn_count`, `DURATION`, номера треков/дисков
(`TRACKNUMBER`, `TRACKTOTAL`, `TOTALTRACKS`, `TRACKS`, `TOTALTRACK`, `TRCK`,
`DISCNUMBER`, `DISCTOTAL`, `TOTALDISCS`, `DISCS`, `TOTALDISC`, `TPOS`,
`PART_NUMBER`, `TOTAL_PARTS`).

Этот список можно расширить через `fields.ignore`.

## Где смотреть

```bash
ext-tags config init --full ext-tags.yaml
```

В начале полного конфига будет такой же справочник в комментариях. Он
генерируется из встроенной таблицы алиасов, поэтому не должен расходиться с
кодом.
