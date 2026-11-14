// Package metadata defines the format-independent document produced by
// readers and consumed by writers.
package metadata

// Kind names an artwork slot.
type Kind = string

const (
	Front  Kind = "front"
	Back   Kind = "back"
	Artist Kind = "artist"
	Disc   Kind = "disc"
	Icon   Kind = "icon"
)

// Cover is a picture attached to a document.
type Cover struct {
	// Filename is a suggested name for the APEv2 cover item ("cover.jpg").
	Filename string
	MIME     string
	Data     []byte
}

// Picture is one indexed artwork item from the container.
type Picture struct {
	Source   string
	Slot     string
	Filename string
	MIME     string
	Data     []byte
}

// Document is raw metadata, before any field rules are applied.
type Document struct {
	Format   string
	Fields   map[string]string
	Vars     map[string]string
	Cover    *Cover
	Artwork  map[string]*Cover // front, back, artist, disc, icon
	Pictures []Picture         // all container pictures in native order
}

// New returns a document with initialised maps.
func New() *Document {
	return &Document{
		Fields:  map[string]string{},
		Vars:    map[string]string{},
		Artwork: map[string]*Cover{},
	}
}

// SetArtwork stores a picture and keeps Cover as the front alias.
func (d *Document) SetArtwork(kind string, cover *Cover) {
	if d.Artwork == nil {
		d.Artwork = map[string]*Cover{}
	}
	d.Artwork[kind] = cover
	if kind == Front {
		d.Cover = cover
	}
}
