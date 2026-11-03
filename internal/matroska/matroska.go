// Package matroska reads metadata from Matroska/WebM containers (EBML).
//
// Supported: Info/Title, Tags/SimpleTag (TagName/TagString) and
// Attachments (pictures usable as cover art). Cluster and Cues data is
// skipped, so large video files are not loaded into memory.
package matroska

import (
	"bytes"
	"fmt"
	"io"
	"os"
)

// Element IDs.
const (
	idEBML        = 0x1A45DFA3
	idSegment     = 0x18538067
	idSeekHead    = 0x114D9B74
	idInfo        = 0x1549A966
	idTitle       = 0x7BA9
	idMuxingApp   = 0x4D80
	idWritingApp  = 0x5741
	idTracks      = 0x1654AE6B
	idTrackEntry  = 0xAE
	idTags        = 0x1254C367
	idTag         = 0x7373
	idTargets     = 0x63C0
	idSimpleTag   = 0x67C8
	idTagName     = 0x45A3
	idTagString   = 0x4487
	idTagBinary   = 0x4485
	idTagLanguage = 0x447A
	idAttachments = 0x1941A469
	idAttached    = 0x61A7
	idFileName    = 0x466E
	idFileMime    = 0x4660
	idFileData    = 0x465C
	idFileDesc    = 0x467E
	idChapters    = 0x1043A770
	idEdition     = 0x45B9
	idChapterAtom = 0xB6
	idCluster     = 0x1F43B675
	idCues        = 0x1C53BB6B
)

// SimpleTag is a single Matroska tag.
type SimpleTag struct {
	Name  string
	Value string
}

// Attachment is a file embedded into the container (cover art).
type Attachment struct {
	Filename    string
	MIME        string
	Description string
	Data        []byte
}

// Metadata is the parsed metadata of a Matroska file.
type Metadata struct {
	Title       string
	MuxingApp   string
	WritingApp  string
	Tags        []SimpleTag
	Attachments []Attachment
}

// ReadFile parses a Matroska/WebM file.
func ReadFile(path string) (*Metadata, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return nil, err
	}
	p := &parser{r: f, size: info.Size(), meta: &Metadata{}}
	if err := p.walk(0, p.size, 0); err != nil {
		return nil, err
	}
	return p.meta, nil
}

// Parse parses Matroska data already in memory.
func Parse(data []byte) (*Metadata, error) {
	p := &parser{r: bytes.NewReader(data), size: int64(len(data)), meta: &Metadata{}}
	if err := p.walk(0, p.size, 0); err != nil {
		return nil, err
	}
	return p.meta, nil
}

type parser struct {
	r    io.ReadSeeker
	size int64
	meta *Metadata
}

const unknownSize = int64(-1)

func (p *parser) walk(start, end int64, depth int) error {
	if depth > 64 {
		return nil
	}
	off := start
	for off < end {
		id, size, headerLen, err := p.readHeader(off)
		if err != nil {
			// Tolerate trailing garbage and truncated files.
			return nil
		}
		dataStart := off + int64(headerLen)
		dataEnd := end
		if size != unknownSize {
			dataEnd = dataStart + size
			if dataEnd > end {
				dataEnd = end
			}
		}
		if dataEnd < dataStart {
			return nil
		}

		switch id {
		case idSegment, idEBML, idInfo, idTracks, idTrackEntry, idTags, idTag, idTargets,
			idAttachments, idChapters, idEdition, idChapterAtom:
			if err := p.walk(dataStart, dataEnd, depth+1); err != nil {
				return err
			}
		case idCluster, idCues, idSeekHead:
			// Skip large binary sections.
		case idTitle:
			p.meta.Title = p.readString(dataStart, size)
		case idMuxingApp:
			p.meta.MuxingApp = p.readString(dataStart, size)
		case idWritingApp:
			p.meta.WritingApp = p.readString(dataStart, size)
		case idSimpleTag:
			if err := p.readSimpleTag(dataStart, dataEnd); err != nil {
				return err
			}
		case idAttached:
			if err := p.readAttachment(dataStart, dataEnd); err != nil {
				return err
			}
		}
		off = dataEnd
	}
	return nil
}

func (p *parser) readSimpleTag(start, end int64) error {
	var (
		name  string
		value string
	)
	off := start
	for off < end {
		id, size, headerLen, err := p.readHeader(off)
		if err != nil {
			return nil
		}
		dataStart := off + int64(headerLen)
		dataEnd := end
		if size != unknownSize {
			dataEnd = dataStart + size
			if dataEnd > end {
				dataEnd = end
			}
		}
		switch id {
		case idTagName:
			name = p.readString(dataStart, size)
		case idTagString:
			value = p.readString(dataStart, size)
		case idSimpleTag:
			if err := p.readSimpleTag(dataStart, dataEnd); err != nil {
				return err
			}
		}
		off = dataEnd
	}
	if name != "" {
		p.meta.Tags = append(p.meta.Tags, SimpleTag{Name: name, Value: value})
	}
	return nil
}

func (p *parser) readAttachment(start, end int64) error {
	var att Attachment
	off := start
	for off < end {
		id, size, headerLen, err := p.readHeader(off)
		if err != nil {
			return nil
		}
		dataStart := off + int64(headerLen)
		dataEnd := end
		if size != unknownSize {
			dataEnd = dataStart + size
			if dataEnd > end {
				dataEnd = end
			}
		}
		switch id {
		case idFileName:
			att.Filename = p.readString(dataStart, size)
		case idFileMime:
			att.MIME = p.readString(dataStart, size)
		case idFileDesc:
			att.Description = p.readString(dataStart, size)
		case idFileData:
			att.Data = p.readBytes(dataStart, size)
		}
		off = dataEnd
	}
	if att.Filename != "" || len(att.Data) > 0 {
		p.meta.Attachments = append(p.meta.Attachments, att)
	}
	return nil
}

// readHeader decodes an element ID and size at offset.
func (p *parser) readHeader(offset int64) (id uint64, size int64, headerLen int, err error) {
	var buf [12]byte
	if _, err := p.r.Seek(offset, io.SeekStart); err != nil {
		return 0, 0, 0, err
	}
	n, err := io.ReadFull(p.r, buf[:])
	if err != nil && n == 0 {
		return 0, 0, 0, err
	}
	data := buf[:n]

	idLen, ok := vintLen(data[0], true)
	if !ok || idLen > len(data) || idLen > 4 {
		return 0, 0, 0, fmt.Errorf("invalid element id at %d", offset)
	}
	for i := 0; i < idLen; i++ {
		id = id<<8 | uint64(data[i])
	}

	if len(data) <= idLen {
		return 0, 0, 0, io.ErrUnexpectedEOF
	}
	sizeLen, ok := vintLen(data[idLen], false)
	if !ok || idLen+sizeLen > len(data) {
		return 0, 0, 0, fmt.Errorf("invalid element size at %d", offset)
	}
	first := data[idLen]
	value := uint64(first) & (0xFF >> uint(sizeLen))
	for i := 1; i < sizeLen; i++ {
		value = value<<8 | uint64(data[idLen+i])
	}
	if value == ^uint64(0)>>(uint(8-sizeLen)*8) && isAllOnes(data[idLen:idLen+sizeLen]) {
		size = unknownSize
	} else {
		size = int64(value)
	}
	return id, size, idLen + sizeLen, nil
}

func vintLen(first byte, id bool) (int, bool) {
	if first == 0 {
		return 0, false
	}
	for length := 1; length <= 8; length++ {
		mask := byte(0x80) >> uint(length-1)
		if first&mask != 0 {
			if id && length > 4 {
				return 0, false
			}
			return length, true
		}
	}
	return 0, false
}

func isAllOnes(b []byte) bool {
	first := b[0]
	length := len(b)
	if first != byte(0xFF>>uint(length-1)) {
		return false
	}
	for _, c := range b[1:] {
		if c != 0xFF {
			return false
		}
	}
	return true
}

func (p *parser) readString(offset, size int64) string {
	if size <= 0 || size > 1<<20 {
		return ""
	}
	data := p.readBytes(offset, size)
	return string(bytes.TrimRight(data, "\x00"))
}

func (p *parser) readBytes(offset, size int64) []byte {
	if size <= 0 || size > 64<<20 {
		return nil
	}
	if _, err := p.r.Seek(offset, io.SeekStart); err != nil {
		return nil
	}
	data := make([]byte, size)
	n, err := io.ReadFull(p.r, data)
	if err != nil && n == 0 {
		return nil
	}
	return data[:n]
}
