package update

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewAndErrors(t *testing.T) {
	u := New("1.2.3")
	if u == nil || u.Client == nil || u.APIURL == "" || u.UserAgent != "ext-tags/1.2.3" {
		t.Fatalf("updater = %+v", u)
	}
	if _, err := u.AssetFor(&Release{TagName: "v1.0.0"}); err == nil {
		t.Fatal("expected missing asset")
	}
	if _, err := u.Checksums(&Release{TagName: "v1.0.0"}); err == nil {
		t.Fatal("expected missing checksums")
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/status":
			http.Error(w, "nope", http.StatusInternalServerError)
		case "/empty":
			_, _ = w.Write([]byte("{}"))
		default:
			_, _ = w.Write([]byte("not json"))
		}
	}))
	defer server.Close()

	for _, path := range []string{"/status", "/empty", "/json"} {
		u = &Updater{Client: server.Client(), APIURL: server.URL + path}
		if _, err := u.Check(); err == nil {
			t.Fatalf("expected error for %s", path)
		}
	}
}
