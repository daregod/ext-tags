package version

import (
	"strings"
	"testing"
)

func TestString(t *testing.T) {
	oldVersion, oldCommit, oldDate := Version, Commit, Date
	defer func() { Version, Commit, Date = oldVersion, oldCommit, oldDate }()
	Version, Commit, Date = "v1.2.3", "abc", "today"
	got := String()
	for _, part := range []string{"v1.2.3", "abc", "today"} {
		if !strings.Contains(got, part) {
			t.Errorf("String() = %q, missing %q", got, part)
		}
	}
}
