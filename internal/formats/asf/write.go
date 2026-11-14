package asf

import (
	"encoding/binary"
	"fmt"
	"os"
	"sort"
	"strings"
	"unicode/utf16"
)

var guidFileProperties = [16]byte{0xA1, 0xDC, 0xAB, 0x8C, 0x47, 0xA9, 0xCF, 0x11, 0x8E, 0xE4, 0x00, 0xC0, 0x0C, 0x20, 0x53, 0x65}

var pictureTypes = map[string]byte{
	"front":  3,
	"back":   4,
	"artist": 8,
	"disc":   6,
	"icon":   1,
}

var descriptorFields = map[string]string{
	"Album":        "WM/AlbumTitle",
	"Album artist": "WM/AlbumArtist",
	"Genre":        "WM/Genre",
	"Year":         "date",
	"Track":        "WM/TrackNumber",
	"Disc":         "WM/PartOfSet",
	"Composer":     "WM/Composer",
	"Publisher":    "WM/Publisher",
	"Encoder":      "WM/EncodedBy",
}

// Cover is one WM/Picture image.
type Cover struct {
	Filename string
	MIME     string
	Data     []byte
}

type object struct {
	guid [16]byte
	raw  []byte
}

// Write replaces ASF content descriptions and pictures, preserving the rest
// of the header and all data objects.
func Write(path string, fields map[string]string, covers map[string]*Cover, replace bool) error {
	return write(path, fields, covers, replace, nil)
}

// WriteOnly synchronises selected fields/artwork and removes selected values
// that are absent from the new data.
func WriteOnly(path string, fields map[string]string, covers map[string]*Cover, only []string) error {
	content := map[string]string{}
	md, err := ReadFile(path)
	if err == nil {
		content["Title"] = md.Title
		content["Artist"] = md.Author
		content["Copyright"] = md.Copyright
		content["Comment"] = md.Description
		content["Rating"] = md.Rating
	}
	removeDesc := make(map[string]bool)
	removePictures := make(map[byte]bool)
	for _, field := range only {
		switch field {
		case "Title", "Artist", "Copyright", "Comment", "Rating":
			if fields[field] != "" {
				content[field] = fields[field]
			} else {
				delete(content, field)
			}
		case "Album", "Album artist", "Genre", "Year", "Track", "Disc", "Composer", "Publisher", "Encoder":
			removeDesc[strings.ToLower(descriptorFields[field])] = true
		case "Cover Art (Front)", "Cover Art (Back)", "Artist Picture", "Cover Art (Disc)", "Cover Art (Icon)":
			if typ, ok := pictureTypes[asfArtworkKind(field)]; ok {
				removePictures[typ] = true
			}
		default:
			return fmt.Errorf("--only field %q is not supported for ASF", field)
		}
	}
	return write(path, fields, covers, false, &writeOptions{content: content, removeDesc: removeDesc, removePictures: removePictures})
}

type writeOptions struct {
	content        map[string]string
	removeDesc     map[string]bool
	removePictures map[byte]bool
}

func write(path string, fields map[string]string, covers map[string]*Cover, replace bool, opts *writeOptions) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	headerSize := int(binary.LittleEndian.Uint64(data[16:24]))
	if headerSize < 30 || headerSize > len(data) {
		return fmt.Errorf("asf: invalid header size")
	}
	objects, err := parseObjectsRaw(data[30:headerSize])
	if err != nil {
		return err
	}

	var kept []object
	var existingDesc []Descriptor
	insertAt := -1
	for _, obj := range objects {
		switch obj.guid {
		case guidContentDesc, guidExtendedContent:
			if insertAt < 0 {
				insertAt = len(kept)
			}
			if obj.guid == guidExtendedContent {
				existingDesc = parseDescriptorList(obj.raw[24:])
			}
			continue
		default:
			kept = append(kept, obj)
		}
	}
	if insertAt < 0 {
		insertAt = len(kept)
	}

	contentFields := fields
	if opts != nil && opts.content != nil {
		contentFields = opts.content
	}
	content := buildContentDescription(contentFields)
	ext := buildExtendedContentOpts(fields, covers, existingDesc, replace, opts)
	newObjects := make([]object, 0, len(kept)+2)
	newObjects = append(newObjects, kept[:insertAt]...)
	newObjects = append(newObjects, object{guid: guidContentDesc, raw: makeObject(guidContentDesc, content)})
	if len(ext) > 0 {
		newObjects = append(newObjects, object{guid: guidExtendedContent, raw: makeObject(guidExtendedContent, ext)})
	}
	newObjects = append(newObjects, kept[insertAt:]...)
	kept = newObjects

	total := 30
	for _, obj := range kept {
		total += len(obj.raw)
	}
	newHeaderSize := total
	newFileSize := newHeaderSize + (len(data) - headerSize)
	for i := range kept {
		if kept[i].guid != guidFileProperties {
			continue
		}
		payload := append([]byte(nil), kept[i].raw[24:]...)
		if len(payload) >= 24 {
			binary.LittleEndian.PutUint64(payload[16:24], uint64(newFileSize))
		}
		kept[i].raw = makeObject(guidFileProperties, payload)
	}

	out := make([]byte, 30, newFileSize)
	copy(out[0:16], guidHeader[:])
	binary.LittleEndian.PutUint64(out[16:24], uint64(newHeaderSize))
	binary.LittleEndian.PutUint32(out[24:28], uint32(len(kept)))
	out[28] = 1
	out[29] = 2
	for _, obj := range kept {
		out = append(out, obj.raw...)
	}
	out = append(out, data[headerSize:]...)
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	return os.WriteFile(path, out, info.Mode().Perm())
}

func parseObjectsRaw(data []byte) ([]object, error) {
	var objects []object
	offset := 0
	for offset+24 <= len(data) {
		guid := [16]byte(data[offset : offset+16])
		size := int(binary.LittleEndian.Uint64(data[offset+16 : offset+24]))
		if size < 24 || offset+size > len(data) {
			return nil, fmt.Errorf("asf: bad object size")
		}
		objects = append(objects, object{guid: guid, raw: append([]byte(nil), data[offset:offset+size]...)})
		offset += size
	}
	if offset != len(data) {
		return nil, fmt.Errorf("asf: trailing header bytes")
	}
	return objects, nil
}

func makeObject(guid [16]byte, payload []byte) []byte {
	out := make([]byte, 24, 24+len(payload))
	copy(out[0:16], guid[:])
	binary.LittleEndian.PutUint64(out[16:24], uint64(24+len(payload)))
	return append(out, payload...)
}

func buildContentDescription(fields map[string]string) []byte {
	values := []string{
		fields["Title"], fields["Artist"], fields["Copyright"],
		fields["Comment"], fields["Rating"],
	}
	var out []byte
	for _, value := range values {
		encoded := utf16Z(value)
		if len(encoded) > 0xFFFF {
			encoded = encoded[:0xFFFF]
		}
		length := make([]byte, 2)
		binary.LittleEndian.PutUint16(length, uint16(len(encoded)))
		out = append(out, length...)
	}
	for _, value := range values {
		out = append(out, utf16Z(value)...)
	}
	return out
}

type extDescriptor struct {
	name  string
	typ   uint16
	value []byte
}

func buildExtendedContent(fields map[string]string, covers map[string]*Cover, existing []Descriptor, replace bool) []byte {
	return buildExtendedContentOpts(fields, covers, existing, replace, nil)
}

func buildExtendedContentOpts(fields map[string]string, covers map[string]*Cover, existing []Descriptor, replace bool, opts *writeOptions) []byte {
	var descriptors []extDescriptor
	replaced := make(map[string]bool)
	for field, name := range descriptorFields {
		if fields[field] == "" {
			continue
		}
		names := []string{name}
		if field == "Year" {
			names = []string{"date", "WM/Year"}
		}
		for _, n := range names {
			replaced[strings.ToLower(n)] = true
			descriptors = append(descriptors, extDescriptor{name: n, typ: 0, value: utf16Z(fields[field])})
		}
	}
	added := make(map[byte]bool)
	for kind, pictureType := range pictureTypes {
		cover := covers[kind]
		if cover == nil || len(cover.Data) == 0 {
			continue
		}
		value, err := encodePicture(pictureType, cover)
		if err != nil {
			continue
		}
		descriptors = append(descriptors, extDescriptor{name: "WM/Picture", typ: 1, value: value})
		added[pictureType] = true
	}
	if !replace {
		for _, d := range existing {
			lower := strings.ToLower(d.Name)
			if lower == "wm/picture" {
				if len(d.Value) > 0 && (added[d.Value[0]] || (opts != nil && opts.removePictures[d.Value[0]])) {
					continue
				}
				descriptors = append(descriptors, extDescriptor{name: d.Name, typ: d.Type, value: d.Value})
				continue
			}
			if replaced[lower] || (opts != nil && opts.removeDesc[lower]) {
				continue
			}
			descriptors = append(descriptors, extDescriptor{name: d.Name, typ: d.Type, value: d.Value})
		}
	}
	if len(descriptors) == 0 {
		return nil
	}
	sort.SliceStable(descriptors, func(i, j int) bool { return descriptors[i].name < descriptors[j].name })
	out := make([]byte, 2, 64)
	binary.LittleEndian.PutUint16(out, uint16(len(descriptors)))
	for _, d := range descriptors {
		name := utf16Z(d.name)
		if len(name) > 0xFFFF || len(d.value) > 0xFFFF {
			continue
		}
		length := make([]byte, 2)
		binary.LittleEndian.PutUint16(length, uint16(len(name)))
		out = append(out, length...)
		out = append(out, name...)
		meta := make([]byte, 4)
		binary.LittleEndian.PutUint16(meta[0:2], d.typ)
		binary.LittleEndian.PutUint16(meta[2:4], uint16(len(d.value)))
		out = append(out, meta...)
		out = append(out, d.value...)
	}
	return out
}

func parseDescriptorList(payload []byte) []Descriptor {
	if len(payload) < 2 {
		return nil
	}
	count := int(binary.LittleEndian.Uint16(payload[0:2]))
	offset := 2
	var out []Descriptor
	for i := 0; i < count; i++ {
		if offset+2 > len(payload) {
			break
		}
		nameLen := int(binary.LittleEndian.Uint16(payload[offset : offset+2]))
		offset += 2
		if offset+nameLen+4 > len(payload) {
			break
		}
		name := UTF16String(payload[offset : offset+nameLen])
		offset += nameLen
		valueType := binary.LittleEndian.Uint16(payload[offset : offset+2])
		valueLen := int(binary.LittleEndian.Uint16(payload[offset+2 : offset+4]))
		offset += 4
		if offset+valueLen > len(payload) {
			break
		}
		value := append([]byte(nil), payload[offset:offset+valueLen]...)
		offset += valueLen
		out = append(out, Descriptor{Name: name, Type: valueType, Value: value})
	}
	return out
}

func encodePicture(pictureType byte, cover *Cover) ([]byte, error) {
	mime := cover.MIME
	if mime == "" {
		mime = "image/jpeg"
	}
	out := []byte{pictureType}
	length := make([]byte, 4)
	binary.LittleEndian.PutUint32(length, uint32(len(cover.Data)))
	out = append(out, length...)
	out = append(out, utf16Z(mime)...)
	out = append(out, utf16Z("")...)
	out = append(out, cover.Data...)
	if len(out) > 0xFFFF {
		return nil, fmt.Errorf("asf: picture too large for extended content")
	}
	return out, nil
}

func utf16Z(value string) []byte {
	out := utf16Bytes(value)
	return append(out, 0, 0)
}

func utf16Bytes(value string) []byte {
	units := utf16.Encode([]rune(value))
	out := make([]byte, len(units)*2)
	for i, unit := range units {
		binary.LittleEndian.PutUint16(out[i*2:], unit)
	}
	return out
}
func asfArtworkKind(field string) string {
	switch strings.ToLower(strings.TrimSpace(field)) {
	case "cover art (front)":
		return "front"
	case "cover art (back)":
		return "back"
	case "artist picture":
		return "artist"
	case "cover art (disc)":
		return "disc"
	case "cover art (icon)":
		return "icon"
	}
	return ""
}
