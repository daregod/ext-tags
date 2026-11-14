// Package mp4 rewrites iTunes metadata (moov/udta/meta/ilst) in MP4 files.
package mp4

import (
	"encoding/binary"
	"fmt"
	"os"
	"sort"
	"strconv"
)

// Cover is the front cover written to the covr atom.
type Cover struct {
	Filename string
	MIME     string
	Data     []byte
}

type box struct {
	typ  string
	raw  []byte
	data []byte
}

var textAtoms = map[string]string{
	"Title":        "\xa9nam",
	"Artist":       "\xa9ART",
	"Album":        "\xa9alb",
	"Album artist": "aART",
	"Year":         "\xa9day",
	"Genre":        "\xa9gen",
	"Composer":     "\xa9wrt",
	"Comment":      "\xa9cmt",
	"Copyright":    "cprt",
	"Encoder":      "\xa9too",
}

var containerTypes = map[string]bool{
	"moov": true, "trak": true, "mdia": true, "minf": true,
	"stbl": true, "edts": true, "dinf": true, "mvex": true,
}

// Write merges fields and the front cover into an MP4 file.
func Write(path string, fields map[string]string, cover *Cover, replace bool) error {
	return write(path, fields, cover, replace, nil)
}

// WriteOnly writes selected fields and removes selected atoms that are
// missing. It rejects fields/artwork that MP4 cannot represent.
func WriteOnly(path string, fields map[string]string, cover *Cover, replace bool, only []string) error {
	remove := make(map[string]bool)
	for _, field := range only {
		switch field {
		case "Title", "Artist", "Album", "Album artist", "Year", "Genre", "Composer", "Comment", "Copyright", "Encoder":
			remove[textAtoms[field]] = true
		case "Track":
			remove["trkn"] = true
		case "Disc":
			remove["disk"] = true
		case "Cover Art (Front)":
			remove["covr"] = true
		default:
			return fmt.Errorf("--only field %q is not supported for MP4", field)
		}
	}
	return write(path, fields, cover, replace, remove)
}

func write(path string, fields map[string]string, cover *Cover, replace bool, remove map[string]bool) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	boxes, err := parseBoxes(data)
	if err != nil {
		return err
	}
	moovIndex := -1
	offset := int64(0)
	moovOffset := int64(0)
	for i, b := range boxes {
		if b.typ == "moov" {
			moovIndex = i
			moovOffset = offset
			break
		}
		offset += int64(len(b.raw))
	}
	if moovIndex < 0 {
		return fmt.Errorf("mp4: moov box not found")
	}
	oldMoov := boxes[moovIndex].raw
	oldEnd := moovOffset + int64(len(oldMoov))

	payload, err := rewriteMoov(boxes[moovIndex].data, fields, cover, replace, remove)
	if err != nil {
		return err
	}
	newMoov := makeBox("moov", payload)
	delta := int64(len(newMoov) - len(oldMoov))
	if delta != 0 {
		if err := shiftOffsets(payload, oldEnd, delta); err != nil {
			return err
		}
		newMoov = makeBox("moov", payload)
	}

	out := make([]byte, 0, len(data)+int(delta))
	for i, b := range boxes {
		if i == moovIndex {
			out = append(out, newMoov...)
			continue
		}
		out = append(out, b.raw...)
	}
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	return os.WriteFile(path, out, info.Mode().Perm())
}

func rewriteMoov(payload []byte, fields map[string]string, cover *Cover, replace bool, remove map[string]bool) ([]byte, error) {
	children, err := parseBoxes(payload)
	if err != nil {
		return nil, err
	}
	found := false
	var out []byte
	for _, b := range children {
		if b.typ == "udta" {
			udta, err := rewriteUdta(b.data, fields, cover, replace, remove)
			if err != nil {
				return nil, err
			}
			out = append(out, makeBox("udta", udta)...)
			found = true
			continue
		}
		out = append(out, b.raw...)
	}
	if !found {
		udta, err := rewriteUdta(nil, fields, cover, replace, remove)
		if err != nil {
			return nil, err
		}
		out = append(out, makeBox("udta", udta)...)
	}
	return out, nil
}

func rewriteUdta(payload []byte, fields map[string]string, cover *Cover, replace bool, remove map[string]bool) ([]byte, error) {
	var children []box
	var err error
	if len(payload) > 0 {
		children, err = parseBoxes(payload)
		if err != nil {
			return nil, err
		}
	}
	found := false
	var out []byte
	for _, b := range children {
		if b.typ == "meta" {
			meta, err := rewriteMeta(b.data, fields, cover, replace, remove)
			if err != nil {
				return nil, err
			}
			out = append(out, makeBox("meta", meta)...)
			found = true
			continue
		}
		out = append(out, b.raw...)
	}
	if !found {
		meta, err := rewriteMeta(nil, fields, cover, replace, remove)
		if err != nil {
			return nil, err
		}
		out = append(out, makeBox("meta", meta)...)
	}
	return out, nil
}

func rewriteMeta(payload []byte, fields map[string]string, cover *Cover, replace bool, remove map[string]bool) ([]byte, error) {
	prefix := make([]byte, 4)
	if len(payload) >= 4 {
		copy(prefix, payload[:4])
		payload = payload[4:]
	}
	children, err := parseBoxes(payload)
	if err != nil {
		return nil, err
	}
	var hdlr []byte
	hasIlst := false
	var outChildren []byte
	for _, b := range children {
		switch b.typ {
		case "hdlr":
			hdlr = b.raw
		case "ilst":
			ilst, err := rewriteIlst(b.data, fields, cover, replace, remove)
			if err != nil {
				return nil, err
			}
			outChildren = append(outChildren, makeBox("ilst", ilst)...)
			hasIlst = true
		default:
			outChildren = append(outChildren, b.raw...)
		}
	}
	if !hasIlst {
		ilst, err := rewriteIlst(nil, fields, cover, replace, remove)
		if err != nil {
			return nil, err
		}
		outChildren = append(outChildren, makeBox("ilst", ilst)...)
	}
	if len(hdlr) == 0 {
		hdlr = makeHdlr()
	}
	out := append([]byte(nil), prefix...)
	out = append(out, hdlr...)
	return append(out, outChildren...), nil
}

func rewriteIlst(payload []byte, fields map[string]string, cover *Cover, replace bool, remove map[string]bool) ([]byte, error) {
	var children []box
	var err error
	if len(payload) > 0 {
		children, err = parseBoxes(payload)
		if err != nil {
			return nil, err
		}
	}
	wanted := make(map[string]bool)
	for field := range fields {
		if atom, ok := textAtoms[field]; ok {
			wanted[atom] = true
		}
	}
	if _, ok := fields["Track"]; ok {
		wanted["trkn"] = true
	}
	if _, ok := fields["Disc"]; ok {
		wanted["disk"] = true
	}
	if cover != nil && len(cover.Data) > 0 {
		wanted["covr"] = true
	}

	var out []byte
	for _, b := range children {
		if replace || wanted[b.typ] || remove[b.typ] {
			continue
		}
		out = append(out, b.raw...)
	}

	keys := make([]string, 0, len(fields))
	for key := range fields {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, field := range keys {
		value := fields[field]
		if value == "" {
			continue
		}
		switch field {
		case "Track", "Disc":
			atom := "trkn"
			if field == "Disc" {
				atom = "disk"
			}
			num, total := splitPair(value)
			data := make([]byte, 8)
			setBE16(data[2:4], num)
			setBE16(data[4:6], total)
			out = append(out, makeBox(atom, makeData(data, 0))...)
		default:
			atom, ok := textAtoms[field]
			if !ok {
				continue
			}
			out = append(out, makeBox(atom, makeData([]byte(value), 1))...)
		}
	}
	if cover != nil && len(cover.Data) > 0 {
		dataType := uint32(13)
		if cover.MIME == "image/png" {
			dataType = 14
		}
		out = append(out, makeBox("covr", makeData(cover.Data, dataType))...)
	}
	return out, nil
}

func makeData(value []byte, dataType uint32) []byte {
	buf := make([]byte, 8, 8+len(value))
	binary.BigEndian.PutUint32(buf[0:4], dataType)
	buf = append(buf, value...)
	return makeBox("data", buf)
}

func makeHdlr() []byte {
	payload := make([]byte, 24)
	copy(payload[8:12], "mdir")
	return makeBox("hdlr", payload)
}

func makeBox(typ string, payload []byte) []byte {
	size := 8 + len(payload)
	out := make([]byte, 8, size)
	binary.BigEndian.PutUint32(out[0:4], uint32(size))
	copy(out[4:8], typ)
	return append(out, payload...)
}

func parseBoxes(data []byte) ([]box, error) {
	var boxes []box
	offset := 0
	for offset+8 <= len(data) {
		size := int(binary.BigEndian.Uint32(data[offset : offset+4]))
		typ := string(data[offset+4 : offset+8])
		header := 8
		switch size {
		case 1:
			if offset+16 > len(data) {
				return nil, fmt.Errorf("mp4: truncated 64-bit box")
			}
			size = int(binary.BigEndian.Uint64(data[offset+8 : offset+16]))
			header = 16
		case 0:
			size = len(data) - offset
		}
		if size < header || offset+size > len(data) {
			return nil, fmt.Errorf("mp4: bad box %q size %d", typ, size)
		}
		raw := data[offset : offset+size]
		boxes = append(boxes, box{typ: typ, raw: raw, data: raw[header:]})
		offset += size
	}
	if offset != len(data) {
		return nil, fmt.Errorf("mp4: trailing bytes")
	}
	return boxes, nil
}

func shiftOffsets(payload []byte, threshold, delta int64) error {
	boxes, err := parseBoxes(payload)
	if err != nil {
		return err
	}
	for _, b := range boxes {
		switch b.typ {
		case "stco":
			if len(b.data) < 8 {
				continue
			}
			count := int(binary.BigEndian.Uint32(b.data[4:8]))
			for i := 0; i < count; i++ {
				pos := 8 + i*4
				if pos+4 > len(b.data) {
					return fmt.Errorf("mp4: truncated stco")
				}
				v := int64(binary.BigEndian.Uint32(b.data[pos : pos+4]))
				if v >= threshold {
					binary.BigEndian.PutUint32(b.data[pos:pos+4], uint32(v+delta))
				}
			}
		case "co64":
			if len(b.data) < 8 {
				continue
			}
			count := int(binary.BigEndian.Uint32(b.data[4:8]))
			for i := 0; i < count; i++ {
				pos := 8 + i*8
				if pos+8 > len(b.data) {
					return fmt.Errorf("mp4: truncated co64")
				}
				v := int64(binary.BigEndian.Uint64(b.data[pos : pos+8]))
				if v >= threshold {
					binary.BigEndian.PutUint64(b.data[pos:pos+8], uint64(v+delta))
				}
			}
		default:
			if containerTypes[b.typ] {
				if err := shiftOffsets(b.data, threshold, delta); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func splitPair(value string) (string, string) {
	for i := 0; i < len(value); i++ {
		if value[i] == '/' {
			return value[:i], value[i+1:]
		}
	}
	return value, ""
}

func setBE16(dst []byte, value string) {
	n, _ := strconv.Atoi(value)
	binary.BigEndian.PutUint16(dst, uint16(n))
}
