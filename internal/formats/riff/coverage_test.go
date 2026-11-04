package riff

import "testing"

func TestInfoField(t *testing.T) {
	cases := map[string]string{
		"INAM": "TITLE", "IART": "ARTIST", "IPRD": "ALBUM",
		"ICRD": "YEAR", "IGNR": "GENRE", "ICMT": "COMMENT",
		"ITRK": "TRACKNUMBER", "IPRT": "TRACKNUMBER",
		"ICOP": "COPYRIGHT", "ISFT": "ENCODER", "XXXX": "",
	}
	for id, want := range cases {
		if got := infoField(id); got != want {
			t.Errorf("%s: %q, want %q", id, got, want)
		}
	}
}
