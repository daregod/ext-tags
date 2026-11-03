# Generates tiny test fixtures for every format ext-tags knows about (or plans
# to). Requires ffmpeg with libmp3lame/libvorbis/libopus/libspeex/wavpack.
#
#   pwsh -File scripts/gen-fixtures.ps1
#   make fixtures
param(
    [string]$OutDir = "testdata/formats",
    [switch]$Force
)

$ErrorActionPreference = 'Stop'
New-Item -ItemType Directory -Force -Path $OutDir | Out-Null
$out = (Get-Item $OutDir).FullName
Get-ChildItem -Path $out -Filter '*.tag' -File -ErrorAction SilentlyContinue | Remove-Item -Force
Get-ChildItem -Path $out -Filter '*.tag.bak' -File -ErrorAction SilentlyContinue | Remove-Item -Force

$meta = @(
    '-metadata', 'title=Fixture Title',
    '-metadata', 'artist=Fixture Artist',
    '-metadata', 'album=Fixture Album',
    '-metadata', 'genre=Test',
    '-metadata', 'date=2026',
    '-metadata', 'track=3'
)

function Invoke-FFmpeg {
    param([string]$Name, [string[]]$Arguments)
    $path = Join-Path $out $Name
    if (-not $Force -and (Test-Path $path)) {
        Write-Host ("{0,-18} skip" -f $Name)
        return
    }
    & ffmpeg -y -hide_banner -loglevel error @Arguments
    if ($LASTEXITCODE -ne 0) { throw "ffmpeg failed for $Name" }
    if (-not (Test-Path $path)) { throw "ffmpeg produced no file $Name" }
    Write-Host ("{0,-18} {1,7} B" -f $Name, (Get-Item $path).Length)
}

$cover = Join-Path $out 'cover.jpg'
& ffmpeg -y -hide_banner -loglevel error -f lavfi -i "color=c=red:s=32x32:d=1" -frames:v 1 $cover
$folder = Join-Path $out 'folder.jpg'
& ffmpeg -y -hide_banner -loglevel error -f lavfi -i "color=c=green:s=32x32:d=1" -frames:v 1 $folder
$poster = Join-Path $out 'poster.jpg'
$box = Join-Path $out 'box.png'
& ffmpeg -y -hide_banner -loglevel error -f lavfi -i "color=c=yellow:s=32x32:d=1" -frames:v 1 $box
& ffmpeg -y -hide_banner -loglevel error -f lavfi -i "color=c=blue:s=32x32:d=1" -frames:v 1 $poster

$audio = @('-f', 'lavfi', '-i', 'sine=frequency=440:duration=0.4:sample_rate=22050', '-ac', '1')
$audioCover = @('-i', $cover, '-map', '0:a', '-map', '1:v', '-c:v', 'copy', '-disposition:v', 'attached_pic')

Invoke-FFmpeg 'audio.mp3'  ($audio + $audioCover + @('-c:a', 'libmp3lame', '-b:a', '48k', '-id3v2_version', '3') + $meta + @((Join-Path $out 'audio.mp3')))
Invoke-FFmpeg 'audio-id3v2.4.mp3' ($audio + $audioCover + @('-c:a', 'libmp3lame', '-b:a', '48k', '-id3v2_version', '4') + $meta + @((Join-Path $out 'audio-id3v2.4.mp3')))
Invoke-FFmpeg 'audio-id3v1-src.mp3' ($audio + @('-c:a', 'libmp3lame', '-b:a', '48k', '-write_id3v1', '1') + $meta + @((Join-Path $out 'audio-id3v1-src.mp3')))
$v1src = [IO.File]::ReadAllBytes((Join-Path $out 'audio-id3v1-src.mp3'))
$v1tagLen = (($v1src[6] -band 0x7F) -shl 21) -bor (($v1src[7] -band 0x7F) -shl 14) -bor (($v1src[8] -band 0x7F) -shl 7) -bor ($v1src[9] -band 0x7F)
[IO.File]::WriteAllBytes((Join-Path $out 'audio-id3v1.mp3'), $v1src[(10 + $v1tagLen)..($v1src.Length - 1)])
Remove-Item (Join-Path $out 'audio-id3v1-src.mp3')
Write-Host ("{0,-18} {1,7} B" -f 'audio-id3v1.mp3', (Get-Item (Join-Path $out 'audio-id3v1.mp3')).Length)
Invoke-FFmpeg 'audio-raw.mp2' ($audio + @('-c:a', 'mp2', '-b:a', '64k', '-f', 'mp2') + @((Join-Path $out 'audio-raw.mp2')))
# Raw MP2 has no tag container: prepend the ID3v2 tag generated for audio.mp3.
$mp3 = [IO.File]::ReadAllBytes((Join-Path $out 'audio.mp3'))
$tagLen = (($mp3[6] -band 0x7F) -shl 21) -bor (($mp3[7] -band 0x7F) -shl 14) -bor (($mp3[8] -band 0x7F) -shl 7) -bor ($mp3[9] -band 0x7F)
$id3 = $mp3[0..(9 + $tagLen)]
$raw = [IO.File]::ReadAllBytes((Join-Path $out 'audio-raw.mp2'))
[IO.File]::WriteAllBytes((Join-Path $out 'audio.mp2'), ($id3 + $raw))
Remove-Item (Join-Path $out 'audio-raw.mp2')
Write-Host ("{0,-18} {1,7} B" -f 'audio.mp2', (Get-Item (Join-Path $out 'audio.mp2')).Length)
Invoke-FFmpeg 'audio-raw.aac' ($audio + @('-c:a', 'aac', '-b:a', '32k', '-f', 'adts') + @((Join-Path $out 'audio-raw.aac')))
$aacMp3 = [IO.File]::ReadAllBytes((Join-Path $out 'audio.mp3'))
$aacTagLen = (($aacMp3[6] -band 0x7F) -shl 21) -bor (($aacMp3[7] -band 0x7F) -shl 14) -bor (($aacMp3[8] -band 0x7F) -shl 7) -bor ($aacMp3[9] -band 0x7F)
$aacId3 = $aacMp3[0..(9 + $aacTagLen)]
$aacRaw = [IO.File]::ReadAllBytes((Join-Path $out 'audio-raw.aac'))
[IO.File]::WriteAllBytes((Join-Path $out 'audio.aac'), ($aacId3 + $aacRaw))
Remove-Item (Join-Path $out 'audio-raw.aac')
Write-Host ("{0,-18} {1,7} B" -f 'audio.aac', (Get-Item (Join-Path $out 'audio.aac')).Length)
$mp2enc = Join-Path (Split-Path $PSScriptRoot -Parent) 'bin\mjpegtools-2.1.0-mingw-bin64\bin\mp2enc.exe'
if (($Force -or -not (Test-Path (Join-Path $out 'audio.mp1'))) -and (Test-Path $mp2enc)) {
    $mp1wav = Join-Path $out 'audio-mp1-src.wav'
    & ffmpeg -y -hide_banner -loglevel error -f lavfi -i "sine=frequency=440:duration=0.4:sample_rate=44100" -ac 2 -c:a pcm_s16le $mp1wav
    $rawMp1 = Join-Path $out 'audio-mp1-raw.mp1'
    $argLine = '-l 1 -b 384 -r 44100 -o "' + $rawMp1 + '"'
    $proc = Start-Process -FilePath $mp2enc -ArgumentList $argLine -RedirectStandardInput $mp1wav -Wait -PassThru -NoNewWindow
    if ((Test-Path $rawMp1) -and ($null -eq $proc.ExitCode -or $proc.ExitCode -eq 0)) {
        $mp3 = [IO.File]::ReadAllBytes((Join-Path $out 'audio.mp3'))
        $tagLen = (($mp3[6] -band 0x7F) -shl 21) -bor (($mp3[7] -band 0x7F) -shl 14) -bor (($mp3[8] -band 0x7F) -shl 7) -bor ($mp3[9] -band 0x7F)
        $id3 = $mp3[0..(9 + $tagLen)]
        $raw = [IO.File]::ReadAllBytes($rawMp1)
        [IO.File]::WriteAllBytes((Join-Path $out 'audio.mp1'), ($id3 + $raw))
        Remove-Item $rawMp1, $mp1wav
        Write-Host ("{0,-18} {1,7} B" -f 'audio.mp1', (Get-Item (Join-Path $out 'audio.mp1')).Length)
    }
    else {
        Remove-Item $rawMp1, $mp1wav -ErrorAction SilentlyContinue
        Write-Host 'audio.mp1          skip (mp2enc failed)'
    }
}
$flac = Get-Command flac -ErrorAction SilentlyContinue
if ($flac -and ($Force -or -not (Test-Path (Join-Path $out 'audio.w64')))) {
    $tmpFlac = Join-Path $out 'audio-w64.tmp.flac'
    $rawW64 = Join-Path $out 'audio-w64.tmp.w64'
    & flac -s -f -o $tmpFlac (Join-Path $out 'audio.wav')
    $encOK = $LASTEXITCODE -eq 0
    if ($encOK) {
        & flac -s -d -f --force-wave64-format -o $rawW64 $tmpFlac
        $encOK = $LASTEXITCODE -eq 0
    }
    if ($encOK -and (Test-Path $rawW64)) {
        $mp3 = [IO.File]::ReadAllBytes((Join-Path $out 'audio.mp3'))
        $tagLen = (($mp3[6] -band 0x7F) -shl 21) -bor (($mp3[7] -band 0x7F) -shl 14) -bor (($mp3[8] -band 0x7F) -shl 7) -bor ($mp3[9] -band 0x7F)
        $id3 = [byte[]]$mp3[0..(9 + $tagLen)]
        $base = [IO.File]::ReadAllBytes($rawW64)
        # Insert the ID3 chunk right after fmt, before data: foobar2000 only
        # picks up tag chunks in the header area.
        $fmtSize = [BitConverter]::ToUInt64($base, 56)
        $insertAt = 40 + [int]$fmtSize
        $guid = [Text.Encoding]::ASCII.GetBytes('ID3 ') + [byte[]](0xF3, 0xAC, 0xD3, 0x11, 0x8C, 0xD1, 0x00, 0xC0, 0x4F, 0x8E, 0xDB, 0x8A)
        $chunkSize = 24 + $id3.Length
        $chunkLen = $chunkSize
        while (($chunkLen % 8) -ne 0) { $chunkLen++ }
        $chunk = New-Object byte[] $chunkLen
        [Array]::Copy($guid, 0, $chunk, 0, 16)
        [Array]::Copy([BitConverter]::GetBytes([uint64]$chunkSize), 0, $chunk, 16, 8)
        [Array]::Copy($id3, 0, $chunk, 24, $id3.Length)

        $outBytes = New-Object byte[] ($base.Length + $chunk.Length)
        [Array]::Copy($base, 0, $outBytes, 0, $insertAt)
        [Array]::Copy($chunk, 0, $outBytes, $insertAt, $chunk.Length)
        [Array]::Copy($base, $insertAt, $outBytes, $insertAt + $chunk.Length, $base.Length - $insertAt)
        [Array]::Copy([BitConverter]::GetBytes([uint64]$outBytes.Length), 0, $outBytes, 16, 8)
        [IO.File]::WriteAllBytes((Join-Path $out 'audio.w64'), $outBytes)
        Remove-Item $tmpFlac, $rawW64
        Write-Host ("{0,-18} {1,7} B" -f 'audio.w64', (Get-Item (Join-Path $out 'audio.w64')).Length)
    }
    else {
        Remove-Item $tmpFlac, $rawW64 -ErrorAction SilentlyContinue
        Write-Host 'audio.w64          skip (flac failed)'
    }
}
Invoke-FFmpeg 'audio.flac' ($audio + $audioCover + @('-c:a', 'flac') + $meta + @((Join-Path $out 'audio.flac')))
Invoke-FFmpeg 'audio.ogg'  ($audio + @('-c:a', 'libvorbis', '-q:a', '0') + $meta + @((Join-Path $out 'audio.ogg')))
Invoke-FFmpeg 'audio.opus' ($audio + @('-c:a', 'libopus', '-b:a', '24k') + $meta + @((Join-Path $out 'audio.opus')))
Invoke-FFmpeg 'audio-aac.m4a'  ($audio + $audioCover + @('-c:a', 'aac', '-b:a', '32k') + $meta + @((Join-Path $out 'audio-aac.m4a')))
Invoke-FFmpeg 'audio-alac.m4a' ($audio + $audioCover + @('-c:a', 'alac') + $meta + @((Join-Path $out 'audio-alac.m4a')))
Invoke-FFmpeg 'audio.wav'  ($audio + @('-c:a', 'pcm_s16le') + $meta + @((Join-Path $out 'audio.wav')))
Invoke-FFmpeg 'audio-empty.wav' ($audio + @('-fflags', '+bitexact', '-flags:a', '+bitexact', '-c:a', 'pcm_s16le') + @((Join-Path $out 'audio-empty.wav')))
Invoke-FFmpeg 'audio.aiff' ($audio + $audioCover + @('-c:a', 'pcm_s16be', '-write_id3v2', '1', '-id3v2_version', '3') + $meta + @((Join-Path $out 'audio.aiff')))
Invoke-FFmpeg 'audio-unicode.flac' ($audio + $audioCover + @('-c:a', 'flac') + @('-metadata', 'title=Тестовый заголовок', '-metadata', 'artist=Исполнитель', '-metadata', 'album=Альбом', '-metadata', 'genre=Тест', '-metadata', 'date=2026', '-metadata', 'track=3') + @((Join-Path $out 'audio-unicode.flac')))
Invoke-FFmpeg 'audio.wv'   ($audio + @('-c:a', 'wavpack') + $meta + @((Join-Path $out 'audio.wv')))
Invoke-FFmpeg 'audio.wma'  ($audio + @('-c:a', 'wmav2', '-b:a', '32k', '-f', 'asf') + $meta + @((Join-Path $out 'audio.wma')))
Invoke-FFmpeg 'audio.mkv'  ($audio + @('-c:a', 'flac') + $meta + @('-attach', $cover, '-metadata:s:t', 'mimetype=image/jpeg', '-metadata:s:t', 'filename=cover.jpg', (Join-Path $out 'audio.mkv')))
Invoke-FFmpeg 'audio-track.mkv' ($audio + @('-c:a', 'flac') + @('-metadata', 'title=Album Title', '-metadata', 'album=Some Album', '-metadata', 'genre=Test', '-metadata', 'date=2026', '-metadata', 'track=3', '-metadata:s:a:0', 'ARTIST=Track Artist') + @((Join-Path $out 'audio-track.mkv')))
Invoke-FFmpeg 'audio-both.mkv' ($audio + @('-c:a', 'flac') + @('-metadata', 'title=Album Title', '-metadata', 'ALBUM=Some Album', '-metadata', 'ARTIST=Album Artist', '-metadata', 'genre=Test', '-metadata', 'date=2026', '-metadata', 'track=3', '-metadata:s:a:0', 'ARTIST=Track Artist') + @((Join-Path $out 'audio-both.mkv')))
Invoke-FFmpeg 'audio-attachments.mkv' ($audio + @('-c:a', 'flac') + $meta + @('-attach', $poster, '-metadata:s:t:0', 'mimetype=image/jpeg', '-metadata:s:t:0', 'filename=poster.jpg', '-attach', $folder, '-metadata:s:t:1', 'mimetype=image/jpeg', '-metadata:s:t:1', 'filename=folder.jpg', (Join-Path $out 'audio-attachments.mkv')))
Invoke-FFmpeg 'audio-cover.png.mkv' ($audio + @('-c:a', 'flac') + $meta + @('-attach', $box, '-metadata:s:t', 'mimetype=image/png', '-metadata:s:t', 'filename=box.png', (Join-Path $out 'audio-cover.png.mkv')))
Invoke-FFmpeg 'audio.webm' ($audio + @('-c:a', 'libopus', '-b:a', '24k') + $meta + @((Join-Path $out 'audio.webm')))
Invoke-FFmpeg 'audio-flac.oga' ($audio + @('-c:a', 'flac', '-f', 'ogg') + $meta + @((Join-Path $out 'audio-flac.oga')))
Invoke-FFmpeg 'audio.spx'  ($audio + @('-c:a', 'libspeex', '-f', 'ogg') + $meta + @((Join-Path $out 'audio.spx')))
Invoke-FFmpeg 'audio.ac3'  ($audio + @('-c:a', 'ac3', '-b:a', '64k', '-f', 'ac3') + @((Join-Path $out 'audio.ac3')))
Invoke-FFmpeg 'audio.dts'  ($audio + @('-c:a', 'dca', '-strict', '-2', '-f', 'dts') + @((Join-Path $out 'audio.dts')))
Invoke-FFmpeg 'audio.au'   ($audio + @('-c:a', 'pcm_s16be', '-f', 'au') + @((Join-Path $out 'audio.au')))

Invoke-FFmpeg 'audio.rf64' ($audio + @('-c:a', 'pcm_s16le', '-f', 'wav', '-rf64', 'always') + $meta + @((Join-Path $out 'audio.rf64')))
Invoke-FFmpeg 'audio.ts'   ($audio + @('-c:a', 'mp2', '-b:a', '64k', '-f', 'mpegts') + $meta + @((Join-Path $out 'audio.ts')))

$video = @('-f', 'lavfi', '-i', 'color=c=black:s=32x32:d=0.4', '-f', 'lavfi', '-i', 'sine=frequency=440:duration=0.4:sample_rate=22050')
Invoke-FFmpeg 'video.wmv' ($video + @('-c:v', 'wmv2', '-c:a', 'wmav2', '-b:a', '32k', '-f', 'asf') + $meta + @((Join-Path $out 'video.wmv')))
