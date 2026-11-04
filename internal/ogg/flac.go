package ogg

import (
	"encoding/binary"
	"fmt"
	"os"

	"github.com/daregod/ext-tags/internal/vorbis"
)

type flacBlock struct {
	typ  byte
	data []byte
}

func isOggFLAC(packets []packet) bool {
	if len(packets) == 0 {
		return false
	}
	first := packets[0].data
	return len(first) >= 13 && first[0] == 0x7F && string(first[1:5]) == "FLAC" && string(first[9:13]) == "fLaC"
}

func writeOggFLAC(path string, pages []page, packets []packet, comments []vorbis.Comment, pictures []vorbis.Picture, replace bool) error {
	first := append([]byte(nil), packets[0].data...)
	count := int(binary.BigEndian.Uint16(first[7:9]))
	if count < 0 || 1+count > len(packets) {
		return fmt.Errorf("bad Ogg FLAC header count")
	}

	var blocks []flacBlock
	var existingComments []vorbis.Comment
	var existingPictures []vorbis.Picture
	vendor := "ext-tags"
	for i := 1; i <= count; i++ {
		p := packets[i].data
		if len(p) < 4 {
			return fmt.Errorf("truncated Ogg FLAC metadata packet")
		}
		length := int(p[1])<<16 | int(p[2])<<8 | int(p[3])
		if 4+length > len(p) {
			return fmt.Errorf("truncated Ogg FLAC metadata block")
		}
		typ := p[0] & 0x7F
		data := p[4 : 4+length]
		switch typ {
		case 4:
			v, cs, err := vorbis.ParseComments(data, false)
			if err != nil {
				return err
			}
			if v != "" {
				vendor = v
			}
			existingComments = cs
		case 6:
			pic, err := vorbis.ParsePicture(data)
			if err != nil {
				return err
			}
			existingPictures = append(existingPictures, pic)
		default:
			blocks = append(blocks, flacBlock{typ: typ, data: append([]byte(nil), data...)})
		}
	}

	merged := comments
	if !replace {
		merged = mergeComments(existingComments, comments, false)
	}
	mergedPictures := pictures
	if !replace {
		mergedPictures = mergeFLACPictures(existingPictures, pictures)
	}
	if len(merged) > 0 {
		blocks = append(blocks, flacBlock{typ: 4, data: vorbis.MarshalComments(vendor, merged, false)})
	}
	for _, p := range mergedPictures {
		blocks = append(blocks, flacBlock{typ: 6, data: vorbis.MarshalPicture(p)})
	}
	if len(blocks) == 0 {
		return fmt.Errorf("refusing to write empty Ogg FLAC metadata")
	}

	headerPackets := make([][]byte, 0, len(blocks)+1)
	binary.BigEndian.PutUint16(first[7:9], uint16(len(blocks)))
	headerPackets = append(headerPackets, first)
	for i, b := range blocks {
		header := []byte{b.typ, 0, 0, 0}
		if i == len(blocks)-1 {
			header[0] |= 0x80
		}
		header[1] = byte(len(b.data) >> 16)
		header[2] = byte(len(b.data) >> 8)
		header[3] = byte(len(b.data))
		packetData := append(header, b.data...)
		headerPackets = append(headerPackets, packetData)
	}

	audioStart := packets[1+count].startPage
	if packets[count].endPage >= audioStart {
		return fmt.Errorf("ogg FLAC headers share a page with audio")
	}
	headerPages := paginateHeader(headerPackets, pages[0].serial, true)
	for i := range headerPages {
		headerPages[i].seq = uint32(i)
	}
	out := make([]byte, 0, len(first)+len(headerPackets))
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

func mergeFLACPictures(existing, incoming []vorbis.Picture) []vorbis.Picture {
	overridden := make(map[uint32]bool, len(incoming))
	for _, p := range incoming {
		overridden[p.Type] = true
	}
	out := make([]vorbis.Picture, 0, len(existing)+len(incoming))
	for _, p := range existing {
		if !overridden[p.Type] {
			out = append(out, p)
		}
	}
	return append(out, incoming...)
}
