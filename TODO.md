# TODO

## Форматы foobar2000

Статус: ✅ поддерживается, 🟡 частично / нужен реальный файл, ❌ пока нет.

| Формат | Расширения | Где лежат теги | Статус |
|---|---|---|---|
| MP3 | `.mp3` | ID3v1, ID3v2.2–2.4 | ✅ |
| MP2 / MP1 | `.mp2`, `.mp1` | ID3v1/ID3v2 | 🟡 расширения включены, нужен реальный файл |
| M4A + AAC | `.m4a`, `.mp4` | iTunes-атомы | ✅ |
| M4A + ALAC | `.m4a`, `.alac` | iTunes-атомы | ✅ (структура контейнера та же, проверено на переименованном файле) |
| AAC (raw, ADTS) | `.aac` | ID3v2 в начале/конце | ❌ проверить, читается ли через ID3-снайфер |
| Ogg Vorbis | `.ogg` | Vorbis comments | ✅ |
| Opus | `.opus` | Vorbis comments | ✅ |
| FLAC | `.flac` | Vorbis comments | ✅ |
| Ogg FLAC | `.oga`, `.ogg` | Vorbis comments | ❌ проверить поддержку `dhowden/tag` |
| WavPack | `.wv` | APEv2 в конце файла | ✅ новый |
| Monkey's Audio | `.ape` | APEv2 в конце файла | ✅ |
| Musepack | `.mpc` | APEv2 в конце файла | ❌ легко: `ape.ReadFromFileEnd` (магия `MPCK`/`MP+`) |
| TAK | `.tak` | APEv2 в конце файла | ❌ легко: `ape.ReadFromFileEnd` (магия `tBaK`) |
| WAV | `.wav` | LIST/INFO + ID3v2 (в т.ч. после RIFF) | ✅ новый |
| Wave64 | `.w64` | ID3v2-чанк, 64-битные размеры | ❌ нужен отдельный парсер |
| RF64 | `.rf64` | как WAV, 64-битные размеры | ❌ расширить `internal/riff` |
| AIFF | `.aif`, `.aiff` | ID3v2-чанк | ✅ |
| WMA | `.wma` | ASF Content Description | ✅ |
| WMV / ASF | `.wmv`, `.asf` | ASF Content Description | ✅ новый |
| Matroska / WebM | `.mkv`, `.webm` | Matroska Tags, вложения | ✅ |
| Speex | `.spx` | Vorbis comments в Ogg | ❌ Ogg-контейнер + комментарии |
| AC3 | `.ac3` | тегов обычно нет | ⏸ вне области (нет метаданных) |
| DTS | `.dts` | тегов обычно нет | ⏸ вне области |
| AU / SND | `.au`, `.snd` | нет стандартных тегов | ⏸ вне области |
| MPEG Transport Stream | `.ts` | ID3v2 в PES — редкость | ⏸ вне области |
| Audio CD | — | CD-TEXT через Windows API | ⏸ вне области CLI |

### Приоритеты

1. **P1, почти бесплатно**: Musepack и TAK (тот же APEv2-ридер, 3–5 строк на магию), проверить Ogg FLAC, MP2/MP1 и raw AAC на реальных файлах.
2. **P2, средний объём**: Wave64 и RF64 (расширить `internal/riff` на 64-битные размеры и GUID-чанки), Speex (Ogg-комментарии).
3. **P3**: AC3, DTS, AU, TS, Audio CD — стандартных тегов нет или нужен Windows API, ценность низкая.

## Обратный импорт тегов в файлы

- [ ] MP3: запись ID3v2.3/2.4 (приоритет, `bogem/id3v2` или свой writer).
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
