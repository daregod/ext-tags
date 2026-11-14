package matroska

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

// Cover is one artwork attachment.
type Cover struct {
	Filename string
	MIME     string
	Data     []byte
}

const (
	idSeekID             = 0x53AB
	idSeekPosition       = 0x53AC
	idSeek               = 0x4DBB
	idFileUID            = 0x46AE
	idCuePoint           = 0xBB
	idCueTrackPositions  = 0xB7
	idCueClusterPosition = 0xF1
)

var coverOrder = []string{"front", "back", "artist", "disc", "icon"}

// Write appends Tags and Attachments to a Matroska/WebM file. Existing
// trailing Tags/Attachments are replaced; clusters are not moved.
func Write(path string, fields map[string]string, covers map[string]*Cover) error {
	return write(path, fields, covers, false)
}

// WriteKeepAttachments replaces Tags but preserves all existing Attachments.
func WriteKeepAttachments(path string, fields map[string]string) error {
	return write(path, fields, nil, true)
}

func write(path string, fields map[string]string, covers map[string]*Cover, keepAttachments bool) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	tags, okTags := buildTags(fields)
	attachments, okAtt := buildAttachments(covers)
	if !okTags && !okAtt {
		return nil
	}

	header, err := readHeader(data, 0)
	if err != nil {
		return err
	}
	offset := header.idLen + header.sizeLen + int(header.size)
	seg, err := readHeader(data, offset)
	if err != nil {
		return err
	}
	if seg.id != idSegment {
		return fmt.Errorf("matroska: segment not found")
	}
	segmentSizeStart := offset + seg.idLen
	contentStart := segmentSizeStart + seg.sizeLen
	contentEnd := len(data)
	if !seg.unknown {
		contentEnd = contentStart + int(seg.size)
		if contentEnd > len(data) {
			return fmt.Errorf("matroska: truncated segment")
		}
	}

	old := data[contentStart:contentEnd]
	children, err := childRanges(old, 0, len(old))
	if err != nil {
		return err
	}
	content := make([]byte, 0, len(old)+len(tags)+len(attachments))
	mapping := make(map[int]int, len(children))
	for _, c := range children {
		if c.id == idTags || c.id == idSeekHead || (!keepAttachments && c.id == idAttachments) {
			continue
		}
		mapping[c.start] = len(content)
		content = append(content, old[c.start:c.end]...)
	}
	if err := patchCues(content, mapping); err != nil {
		return err
	}
	content = append(content, tags...)
	if !keepAttachments {
		content = append(content, attachments...)
	}

	out := append([]byte(nil), data[:segmentSizeStart]...)
	if seg.unknown {
		out = append(out, data[segmentSizeStart:contentStart]...)
	} else {
		out = append(out, encodeSize(len(content))...)
	}
	out = append(out, content...)
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	return os.WriteFile(path, out, info.Mode().Perm())
}

type childRange struct {
	id    uint64
	start int
	end   int
}

func childRanges(data []byte, start, end int) ([]childRange, error) {
	var out []childRange
	off := start
	for off < end {
		h, err := readHeader(data, off)
		if err != nil {
			return nil, err
		}
		next := off + h.idLen + h.sizeLen + int(h.size)
		if h.unknown || next > end {
			next = end
		}
		out = append(out, childRange{id: h.id, start: off, end: next})
		if next <= off {
			break
		}
		off = next
	}
	return out, nil
}

type elementHeader struct {
	id      uint64
	size    int64
	idLen   int
	sizeLen int
	unknown bool
}

func readHeader(data []byte, off int) (elementHeader, error) {
	h := elementHeader{}
	if off >= len(data) {
		return h, fmt.Errorf("matroska: eof")
	}
	idLen, ok := vintLen(data[off], true)
	if !ok || idLen == 0 || off+idLen > len(data) {
		return h, fmt.Errorf("matroska: bad id")
	}
	for i := 0; i < idLen; i++ {
		h.id = h.id<<8 | uint64(data[off+i])
	}
	h.idLen = idLen
	sizeOff := off + idLen
	if sizeOff >= len(data) {
		return h, fmt.Errorf("matroska: bad size")
	}
	h.sizeLen, ok = vintLen(data[sizeOff], false)
	if !ok || h.sizeLen == 0 || sizeOff+h.sizeLen > len(data) {
		return h, fmt.Errorf("matroska: bad size")
	}
	var value uint64
	for i := 0; i < h.sizeLen; i++ {
		value = value<<8 | uint64(data[sizeOff+i])
	}
	value &= (uint64(1) << (7 * h.sizeLen)) - 1
	if value == (uint64(1)<<(7*h.sizeLen))-1 {
		h.unknown = true
		h.size = int64(len(data) - (sizeOff + h.sizeLen))
	} else {
		h.size = int64(value)
	}
	return h, nil
}

func encodeSize(n int) []byte {
	for width := 1; width <= 8; width++ {
		max := (1 << (7 * width)) - 1
		if n < max {
			value := uint64(n) | (uint64(1) << (7 * width))
			out := make([]byte, width)
			for i := width - 1; i >= 0; i-- {
				out[i] = byte(value)
				value >>= 8
			}
			return out
		}
	}
	panic("matroska: size too large")
}

func mkElem(id uint64, parts ...[]byte) []byte {
	idBytes := encodeID(id)
	payloadLen := 0
	for _, p := range parts {
		payloadLen += len(p)
	}
	out := append([]byte(nil), idBytes...)
	out = append(out, encodeSize(payloadLen)...)
	for _, p := range parts {
		out = append(out, p...)
	}
	return out
}

func encodeID(id uint64) []byte {
	width := 1
	for width < 4 && id >= (uint64(1)<<(8*width)) {
		width++
	}
	out := make([]byte, width)
	for i := width - 1; i >= 0; i-- {
		out[i] = byte(id)
		id >>= 8
	}
	return out
}

func buildTags(fields map[string]string) ([]byte, bool) {
	if len(fields) == 0 {
		return nil, false
	}
	var simple []byte
	keys := make([]string, 0, len(fields))
	for key := range fields {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		value := fields[key]
		if value == "" {
			continue
		}
		switch key {
		case "Track":
			num, total := split(value)
			simple = append(simple, mkSimpleTag("PART_NUMBER", num)...)
			if total != "" {
				simple = append(simple, mkSimpleTag("TOTAL_PARTS", total)...)
			}
		case "Disc":
			num, total := split(value)
			simple = append(simple, mkSimpleTag("DISC", num)...)
			if total != "" {
				simple = append(simple, mkSimpleTag("DISC_TOTAL", total)...)
			}
		default:
			name := strings.ToUpper(strings.ReplaceAll(key, " ", "_"))
			simple = append(simple, mkSimpleTag(name, value)...)
		}
	}
	if len(simple) == 0 {
		return nil, false
	}
	tag := mkElem(idTag, mkElem(idTargets), simple)
	return mkElem(idTags, tag), true
}

func mkSimpleTag(name, value string) []byte {
	return mkElem(idSimpleTag, mkElem(idTagName, []byte(name)), mkElem(idTagString, []byte(value)))
}

func buildAttachments(covers map[string]*Cover) ([]byte, bool) {
	var files []byte
	uid := 1
	for _, kind := range coverOrder {
		cover := covers[kind]
		if cover == nil || len(cover.Data) == 0 {
			continue
		}
		name := cover.Filename
		if name == "" {
			name = kind + ".jpg"
		}
		mime := cover.MIME
		if mime == "" {
			mime = "image/jpeg"
		}
		files = append(files, mkElem(idAttached,
			mkElem(idFileUID, uintBytes(uid)),
			mkElem(idFileName, []byte(name)),
			mkElem(idFileMime, []byte(mime)),
			mkElem(idFileData, cover.Data),
		)...)
		uid++
	}
	if len(files) == 0 {
		return nil, false
	}
	return mkElem(idAttachments, files), true
}

func uintBytes(n int) []byte {
	if n == 0 {
		return []byte{0}
	}
	var out []byte
	for n > 0 {
		out = append([]byte{byte(n)}, out...)
		n >>= 8
	}
	return out
}

func split(value string) (string, string) {
	parts := strings.SplitN(value, "/", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return parts[0], ""
}
func patchCues(content []byte, mapping map[int]int) error {
	children, err := childRanges(content, 0, len(content))
	if err != nil {
		return err
	}
	for _, c := range children {
		if c.id != idCues {
			continue
		}
		h, err := readHeader(content, c.start)
		if err != nil {
			return err
		}
		points, err := childRanges(content, c.start+h.idLen+h.sizeLen, c.end)
		if err != nil {
			return err
		}
		for _, point := range points {
			if point.id != idCuePoint {
				continue
			}
			ph, err := readHeader(content, point.start)
			if err != nil {
				return err
			}
			tracks, err := childRanges(content, point.start+ph.idLen+ph.sizeLen, point.end)
			if err != nil {
				return err
			}
			for _, track := range tracks {
				if track.id != idCueTrackPositions {
					continue
				}
				th, err := readHeader(content, track.start)
				if err != nil {
					return err
				}
				positions, err := childRanges(content, track.start+th.idLen+th.sizeLen, track.end)
				if err != nil {
					return err
				}
				for _, pos := range positions {
					if pos.id != idCueClusterPosition {
						continue
					}
					h, err := readHeader(content, pos.start)
					if err != nil {
						return err
					}
					dataStart := pos.start + h.idLen + h.sizeLen
					var oldPos uint64
					for i := dataStart; i < pos.end; i++ {
						oldPos = oldPos<<8 | uint64(content[i])
					}
					newPos, ok := mapping[int(oldPos)]
					if !ok {
						continue
					}
					size := h.sizeLen
					value := uint64(newPos)
					for i := size - 1; i >= 0; i-- {
						content[dataStart+i] = byte(value)
						value >>= 8
					}
				}
			}
		}
	}
	return nil
}
