package ogg

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/daregod/ext-tags/internal/vorbis"
)

type page struct {
	headerType byte
	granule    uint64
	serial     uint32
	seq        uint32
	lacing     []byte
	payload    []byte
}

type packet struct {
	data      []byte
	startPage int
	endPage   int
}

// WriteComments replaces/merges Vorbis comments and pictures in an Ogg
// Vorbis, Opus or Speex file, rewriting only the header pages.
func WriteComments(path string, comments []vorbis.Comment, pictures []vorbis.Picture, replace bool) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	pages, err := parsePages(data)
	if err != nil {
		return err
	}
	if len(pages) == 0 {
		return fmt.Errorf("no Ogg pages")
	}
	packets, err := collectPackets(pages)
	if err != nil {
		return err
	}
	if len(packets) < 2 {
		return fmt.Errorf("too few Ogg packets")
	}
	if isOggFLAC(packets) {
		return writeOggFLAC(path, pages, packets, comments, pictures, replace)
	}

	codec, commentIndex, headerCount, err := detectCodec(packets)
	if err != nil {
		return err
	}
	if commentIndex >= len(packets) || headerCount > len(packets) {
		return fmt.Errorf("bad %s header layout", codec)
	}

	vendor, existing, err := parseCommentPacket(codec, packets[commentIndex].data)
	if err != nil {
		return err
	}
	if vendor == "" {
		vendor = "ext-tags"
	}

	merged := mergeComments(existing, comments, replace)
	mergedPictures, err := mergePictures(codec, existing, pictures, replace)
	if err != nil {
		return err
	}
	merged = append(merged, encodePictures(mergedPictures)...)
	packetData := buildCommentPacket(codec, vendor, merged)

	headerPackets := make([][]byte, 0, headerCount)
	for i := 0; i < headerCount; i++ {
		if i == commentIndex {
			headerPackets = append(headerPackets, packetData)
		} else {
			headerPackets = append(headerPackets, packets[i].data)
		}
	}

	audioStart := packets[headerCount].startPage
	if packets[headerCount-1].endPage >= audioStart {
		return fmt.Errorf("header packets share a page with audio; unsupported layout")
	}

	headerPages := paginateHeader(headerPackets, pages[0].serial, codec == "opus")
	for i := range headerPages {
		headerPages[i].seq = uint32(i)
	}
	out := make([]byte, 0, len(data)+len(packetData))
	for _, p := range headerPages {
		out = append(out, encodePage(p)...)
	}
	seq := uint32(len(headerPages))
	for i := audioStart; i < len(pages); i++ {
		p := pages[i]
		p.seq = seq
		seq++
		out = append(out, encodePage(p)...)
	}
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	return os.WriteFile(path, out, info.Mode().Perm())
}

// ReadComments returns the vendor and comments of an Ogg Vorbis/Opus/Speex file.
func ReadComments(path string) (string, []vorbis.Comment, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", nil, err
	}
	pages, err := parsePages(data)
	if err != nil {
		return "", nil, err
	}
	packets, err := collectPackets(pages)
	if err != nil {
		return "", nil, err
	}
	codec, idx, _, err := detectCodec(packets)
	if err != nil {
		return "", nil, err
	}
	if idx >= len(packets) {
		return "", nil, fmt.Errorf("missing comment packet")
	}
	return parseCommentPacket(codec, packets[idx].data)
}

func detectCodec(packets []packet) (string, int, int, error) {
	first := packets[0].data
	switch {
	case len(first) >= 7 && first[0] == 1 && string(first[1:7]) == "vorbis":
		return "vorbis", 1, 3, nil
	case len(first) >= 8 && string(first[:8]) == "OpusHead":
		return "opus", 1, 2, nil
	case len(first) >= 8 && string(first[:8]) == "Speex   ":
		return "speex", 1, 2, nil
	}
	return "", 0, 0, fmt.Errorf("unsupported Ogg codec")
}

func parseCommentPacket(codec string, data []byte) (string, []vorbis.Comment, error) {
	var block []byte
	switch codec {
	case "vorbis":
		if len(data) < 8 || data[0] != 3 || string(data[1:7]) != "vorbis" {
			return "", nil, fmt.Errorf("bad Vorbis comment header")
		}
		block = data[7:]
	case "opus":
		if len(data) < 8 || string(data[:8]) != "OpusTags" {
			return "", nil, fmt.Errorf("bad OpusTags header")
		}
		block = data[8:]
	case "speex":
		block = data
	}
	vendor, comments, err := vorbis.ParseComments(block, false)
	return vendor, comments, err
}

func buildCommentPacket(codec, vendor string, comments []vorbis.Comment) []byte {
	switch codec {
	case "vorbis":
		out := []byte{3}
		out = append(out, []byte("vorbis")...)
		out = append(out, vorbis.MarshalComments(vendor, comments, true)...)
		return out
	case "opus":
		out := []byte("OpusTags")
		return append(out, vorbis.MarshalComments(vendor, comments, false)...)
	default:
		return vorbis.MarshalComments(vendor, comments, false)
	}
}

func mergeComments(existing, incoming []vorbis.Comment, replace bool) []vorbis.Comment {
	if replace {
		return incoming
	}
	overridden := make(map[string]bool, len(incoming))
	for _, c := range incoming {
		overridden[strings.ToUpper(c.Name)] = true
	}
	out := make([]vorbis.Comment, 0, len(existing)+len(incoming))
	for _, c := range existing {
		upper := strings.ToUpper(c.Name)
		if upper == "METADATA_BLOCK_PICTURE" || overridden[upper] {
			continue
		}
		out = append(out, c)
	}
	return append(out, incoming...)
}

func mergePictures(codec string, existing []vorbis.Comment, incoming []vorbis.Picture, replace bool) ([]vorbis.Picture, error) {
	var old []vorbis.Picture
	for _, c := range existing {
		if !strings.EqualFold(c.Name, "METADATA_BLOCK_PICTURE") {
			continue
		}
		raw, err := base64.StdEncoding.DecodeString(c.Value)
		if err != nil {
			continue
		}
		p, err := vorbis.ParsePicture(raw)
		if err == nil {
			old = append(old, p)
		}
	}
	if replace {
		return incoming, nil
	}
	overridden := make(map[uint32]bool, len(incoming))
	for _, p := range incoming {
		overridden[p.Type] = true
	}
	out := make([]vorbis.Picture, 0, len(old)+len(incoming))
	for _, p := range old {
		if !overridden[p.Type] {
			out = append(out, p)
		}
	}
	return append(out, incoming...), nil
}

func encodePictures(pictures []vorbis.Picture) []vorbis.Comment {
	out := make([]vorbis.Comment, 0, len(pictures))
	for _, p := range pictures {
		out = append(out, vorbis.Comment{
			Name:  "METADATA_BLOCK_PICTURE",
			Value: base64.StdEncoding.EncodeToString(vorbis.MarshalPicture(p)),
		})
	}
	return out
}

func parsePages(data []byte) ([]page, error) {
	var pages []page
	offset := 0
	for offset+27 <= len(data) {
		if string(data[offset:offset+4]) != "OggS" {
			return nil, fmt.Errorf("bad Ogg page at %d", offset)
		}
		if data[offset+4] != 0 {
			return nil, fmt.Errorf("unsupported Ogg version %d", data[offset+4])
		}
		p := page{
			headerType: data[offset+5],
			granule:    le64(data[offset+6 : offset+14]),
			serial:     le32(data[offset+14 : offset+18]),
			seq:        le32(data[offset+18 : offset+22]),
		}
		nseg := int(data[offset+26])
		offset += 27
		if offset+nseg > len(data) {
			return nil, fmt.Errorf("truncated Ogg segment table")
		}
		p.lacing = append([]byte(nil), data[offset:offset+nseg]...)
		offset += nseg
		payloadLen := 0
		for _, l := range p.lacing {
			payloadLen += int(l)
		}
		if offset+payloadLen > len(data) {
			return nil, fmt.Errorf("truncated Ogg payload")
		}
		p.payload = append([]byte(nil), data[offset:offset+payloadLen]...)
		offset += payloadLen
		pages = append(pages, p)
	}
	if offset != len(data) {
		return nil, fmt.Errorf("trailing bytes after Ogg pages")
	}
	return pages, nil
}

func collectPackets(pages []page) ([]packet, error) {
	if len(pages) == 0 {
		return nil, fmt.Errorf("no pages")
	}
	serial := pages[0].serial
	var packets []packet
	var cur []byte
	start := -1
	for pi, p := range pages {
		if p.serial != serial {
			continue
		}
		offset := 0
		for _, l := range p.lacing {
			if start < 0 {
				start = pi
			}
			size := int(l)
			cur = append(cur, p.payload[offset:offset+size]...)
			offset += size
			if l < 255 {
				packets = append(packets, packet{data: cur, startPage: start, endPage: pi})
				cur = nil
				start = -1
			}
		}
	}
	return packets, nil
}

func paginateHeader(packets [][]byte, serial uint32, split bool) []page {
	var pages []page
	var lacing []byte
	var payload []byte
	first := true
	startsContinued := false
	flush := func(granule uint64, nextContinued bool) {
		ht := byte(0)
		if first {
			ht = 0x02
			first = false
		}
		if startsContinued {
			ht |= 0x01
		}
		pages = append(pages, page{headerType: ht, granule: granule, serial: serial, lacing: lacing, payload: payload})
		lacing = nil
		payload = nil
		startsContinued = nextContinued
	}

	for _, data := range packets {
		segments := lacingFor(len(data))
		for len(segments) > 0 {
			room := 255 - len(lacing)
			if room == 0 {
				flush(^uint64(0), true)
				continue
			}
			take := len(segments)
			if take > room {
				take = room
			}
			consumed := 0
			for _, s := range segments[:take] {
				consumed += int(s)
				lacing = append(lacing, s)
			}
			payload = append(payload, data[:consumed]...)
			data = data[consumed:]
			segments = segments[take:]
			if len(segments) > 0 {
				flush(^uint64(0), true)
			}
		}
		if split && len(lacing) > 0 {
			flush(0, false)
		}
	}
	if len(lacing) > 0 {
		flush(0, false)
	}
	for i := range pages {
		if pages[i].granule != ^uint64(0) {
			pages[i].granule = 0
		}
	}
	return pages
}
func lacingFor(n int) []byte {
	out := make([]byte, 0, n/255+1)
	for n >= 255 {
		out = append(out, 255)
		n -= 255
	}
	out = append(out, byte(n))
	return out
}

func encodePage(p page) []byte {
	out := make([]byte, 27, 27+len(p.lacing)+len(p.payload))
	copy(out, "OggS")
	out[4] = 0
	out[5] = p.headerType
	putLE64(out[6:14], p.granule)
	putLE32(out[14:18], p.serial)
	putLE32(out[18:22], p.seq)
	out[26] = byte(len(p.lacing))
	out = append(out, p.lacing...)
	out = append(out, p.payload...)
	crc := oggCRC(out)
	putLE32(out[22:26], crc)
	return out
}

func le32(b []byte) uint32 {
	return uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24
}

func le64(b []byte) uint64 {
	var v uint64
	for i := 0; i < 8; i++ {
		v |= uint64(b[i]) << (8 * i)
	}
	return v
}

func putLE32(b []byte, v uint32) {
	b[0], b[1], b[2], b[3] = byte(v), byte(v>>8), byte(v>>16), byte(v>>24)
}

func putLE64(b []byte, v uint64) {
	for i := 0; i < 8; i++ {
		b[i] = byte(v >> (8 * i))
	}
}

var crcTable = func() [256]uint32 {
	var t [256]uint32
	for i := 0; i < 256; i++ {
		r := uint32(i) << 24
		for j := 0; j < 8; j++ {
			if r&0x80000000 != 0 {
				r = (r << 1) ^ 0x04C11DB7
			} else {
				r <<= 1
			}
		}
		t[i] = r
	}
	return t
}()

func oggCRC(data []byte) uint32 {
	var crc uint32
	for _, b := range data {
		crc = (crc << 8) ^ crcTable[byte(crc>>24)^b]
	}
	return crc
}
