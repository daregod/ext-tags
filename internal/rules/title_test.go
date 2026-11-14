package rules

import "testing"

func TestTitleSanitizeAndClearVars(t *testing.T) {
	if got := title("hello WORLD"); got != "Hello World" {
		t.Fatalf("title = %q", got)
	}
	vars := map[string]string{"tracknumber": "1", "totaltracks": "2", "discnumber": "1", "disctotal": "2"}
	clearNumberVars(vars, "track")
	if vars["tracknumber"] != "" || vars["totaltracks"] != "" || vars["discnumber"] != "1" {
		t.Fatalf("vars = %v", vars)
	}
	clearNumberVars(vars, "disc")
	if vars["discnumber"] != "" || vars["disctotal"] != "" {
		t.Fatalf("vars = %v", vars)
	}
	for _, input := range []string{"a/b", "..", "x:y", "a\x01b"} {
		if got := sanitize(input); got == "" {
			t.Fatalf("sanitize(%q) empty", input)
		}
	}
}
