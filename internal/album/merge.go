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
		Album:      mergeFieldMaps(existing.Album, generated.Album),
		Tracks:     cloneTracks(existing.Tracks),
		Artwork:    mergeFieldMaps(existing.Artwork, generated.Artwork),
		dir:        existing.dir,
		path:       existing.path,
		albumOrder: generated.albumOrder,
		trackOrder: generated.trackOrder,
		multiline:  generated.multiline,
		quote:      generated.quote,
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
			if !hasFieldFold(current, field) {
				current[field] = value
			}
		}
	}
	return out
}

func mergeFieldMaps(base, extra map[string]string) map[string]string {
	out := make(map[string]string, len(base)+len(extra))
	index := make(map[string]string, len(base)+len(extra))
	add := func(key, value string) {
		lower := strings.ToLower(strings.TrimSpace(key))
		if _, ok := index[lower]; ok {
			return
		}
		index[lower] = key
		out[key] = value
	}
	for key, value := range base {
		add(key, value)
	}
	for key, value := range extra {
		add(key, value)
	}
	return out
}

func hasFieldFold(fields map[string]string, name string) bool {
	lower := strings.ToLower(strings.TrimSpace(name))
	for key := range fields {
		if strings.ToLower(strings.TrimSpace(key)) == lower {
			return true
		}
	}
	return false
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
