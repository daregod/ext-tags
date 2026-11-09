package album

import "strings"

// Merge supplements existing album.yaml values with generated ones. Values
// already present in existing (including explicit empty strings) win.
func Merge(existing, generated *Config) *Config {
	if existing == nil {
		return generated
	}
	if generated == nil {
		return existing
	}
	out := &Config{
		Album:   cloneFields(existing.Album),
		Tracks:  cloneTracks(existing.Tracks),
		Artwork: cloneFields(existing.Artwork),
		dir:     existing.dir,
		path:    existing.path,
	}
	if out.Album == nil {
		out.Album = map[string]string{}
	}
	if out.Artwork == nil {
		out.Artwork = map[string]string{}
	}
	for key, value := range generated.Album {
		if _, ok := out.Album[key]; !ok {
			out.Album[key] = value
		}
	}
	for key, value := range generated.Artwork {
		if _, ok := out.Artwork[key]; !ok {
			out.Artwork[key] = value
		}
	}
	if out.Tracks == nil && len(generated.Tracks) > 0 {
		out.Tracks = map[string]map[string]string{}
	}
	actualKeys := make(map[string]string, len(out.Tracks))
	for key := range out.Tracks {
		actualKeys[canonicalTrackKey(key)] = key
	}
	for key, fields := range generated.Tracks {
		actual, ok := actualKeys[canonicalTrackKey(key)]
		if !ok {
			out.Tracks[key] = cloneFields(fields)
			actualKeys[canonicalTrackKey(key)] = key
			continue
		}
		current := out.Tracks[actual]
		for field, value := range fields {
			if _, exists := current[field]; !exists {
				current[field] = value
			}
		}
	}
	return out
}

func canonicalTrackKey(key string) string {
	key = strings.TrimSpace(key)
	if key == "" {
		return key
	}
	for _, r := range key {
		if r < '0' || r > '9' {
			return key
		}
	}
	return trimNumber(key)
}

func cloneFields(in map[string]string) map[string]string {
	if in == nil {
		return nil
	}
	out := make(map[string]string, len(in))
	for key, value := range in {
		out[key] = value
	}
	return out
}

func cloneTracks(in map[string]map[string]string) map[string]map[string]string {
	if in == nil {
		return nil
	}
	out := make(map[string]map[string]string, len(in))
	for key, fields := range in {
		out[key] = cloneFields(fields)
	}
	return out
}
