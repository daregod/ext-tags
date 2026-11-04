# TODO

## Форматы foobar2000

Статус: ✅ поддерживается, 🟡 частично / нужен реальный файл, ❌ пока нет.

| Формат | Расширения | Где лежат теги | Статус |
|---|---|---|---|
| MP3 | `.mp3` | ID3v1, ID3v2.2–2.4 | ✅ |
| MP2 | `.mp2` | ID3v1/ID3v2 | ✅ фикстура (ID3v2 перед потоком) |
| MP1 | `.mp1` | ID3v1/ID3v2 | ✅ фикстура от `mp2enc -l 1`, ID3v2 перед потоком |
| M4A + AAC | `.m4a`, `.mp4` | iTunes-атомы | ✅ |
| M4A + ALAC | `.m4a`, `.alac` | iTunes-атомы | ✅ фикстура |
| AAC (raw, ADTS) | `.aac` | ID3v2 перед потоком | ✅ фикстура |
| Ogg Vorbis | `.ogg` | Vorbis comments | ✅ |
| Opus | `.opus` | Vorbis comments | ✅ |
| FLAC | `.flac` | Vorbis comments | ✅ |
| Ogg FLAC | `.oga` | FLAC-комментарии в Ogg | ✅ свой Ogg-ридер, эталон fb2k совпал |
| WavPack | `.wv` | APEv2 в конце файла | ✅ |
| Monkey's Audio | `.ape` | APEv2 в конце файла | ✅ |
| Musepack | `.mpc` | APEv2 в конце файла | ✅ магия `MPCK`/`MP+` (синтетический тест) |
| TAK | `.tak` | APEv2 в конце файла | ✅ магия `tBaK` (синтетический тест) |
| WAV | `.wav` | LIST/INFO + ID3v2 (в т.ч. после RIFF) | ✅ |
| Wave64 | `.w64` | ID3v2-чанк до `data`, 64-битные размеры | ✅ эталон fb2k совпал |
| RF64 | `.rf64` | как WAV, 64-битные размеры | ✅ `internal/riff` + фикстура |
| AIFF | `.aif`, `.aiff` | NAME/AUTH/ANNO + ID3v2-чанк | ✅ свой IFF-ридер, фикстура с ID3 |
| WMA | `.wma` | ASF Content Description | ✅ |
| WMV / ASF | `.wmv`, `.asf` | ASF Content Description | ✅ |
| Matroska / WebM | `.mkv`, `.webm` | Matroska Tags (с учётом TargetTypeValue), вложения | ✅ |
| Speex | `.spx` | Vorbis comments в Ogg | ✅ свой Ogg-ридер, эталон fb2k совпал |
| AC3 | `.ac3` | APEv2 в конце файла (как fb2k) | ✅ |
| DTS | `.dts` | APEv2; `.dts` может быть Matroska с DTS | ✅ расширение приоритетнее магии |
| AU / SND | `.au`, `.snd` | нет стандартных тегов | ⏸ вне области |
| MPEG Transport Stream | `.ts` | нет тег-контейнера | ⏸ fb2k не читает даже заполненный внешний `.tag` (только обложку) — не поддерживаем |
| Audio CD | — | CD-TEXT через Windows API | ⏸ вне области CLI |

### Приоритеты

1. **P1**: ✅ все форматы «в области» закрыты — Musepack/TAK, Ogg FLAC/Speex, MP2/MP1/raw AAC/ALAC, RF64, Wave64.
2. **P2**: обратный импорт тегов в файлы (раздел ниже).
3. **P3**: AU, TS, Audio CD — стандартных тегов нет или нужен Windows API.

Фикстуры для всех перечисленных форматов уже лежат в `testdata/formats/`
(`make fixtures`), так что новый ридер можно сразу проверять тестом.

## Обратный импорт тегов в файлы

- [x] MP3: запись ID3v1/2.2/2.3/2.4 (`ext-tags import`, свой writer `internal/id3`).
- [ ] FLAC / Ogg Vorbis / Opus: запись Vorbis comments.
- [ ] MP4/M4A (AAC/ALAC): запись iTunes-атомов.
- [ ] APEv2 (APE, WavPack, TAK, Musepack).
- [ ] Matroska/WebM: запись Tags.
- [ ] ASF/WMA: запись Content Description (осторожно, Windows-only сложности).
- [ ] Общая инфраструктура: `--backup`, `--dry-run`, атомарная запись, проверка после записи.

## Инфраструктура

- [ ] `make compat` на фикстурах в CI (сейчас нужны `bin/examples`, которых нет в git).
- [ ] Тесты на реальных файлах для P1-форматов.
- [ ] `merge_existing` для бинарных полей, кроме обложки.
