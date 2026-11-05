package update

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestVersions(t *testing.T) {
	if !IsNewer("v0.10.4", "v0.10.3") || !IsNewer("0.11.0", "0.10.9") {
		t.Fatal("newer versions not detected")
	}
	if IsNewer("0.10.3", "0.10.3") || IsNewer("0.10.2", "0.10.3") {
		t.Fatal("older version detected as newer")
	}
	if !IsNewer("0.10.3", "0.0.0-dev") {
		t.Fatal("dev build must update")
	}
}

func TestChecksumsAndAsset(t *testing.T) {
	sums, err := parseChecksums([]byte("abc  file.tar.gz\n"))
	if err != nil || sums["file.tar.gz"] != "abc" {
		t.Fatalf("sums = %v, %v", sums, err)
	}
	rel := &Release{TagName: "v0.10.4", Assets: []Asset{{Name: "ext-tags_0.10.4_linux_amd64.tar.gz"}}}
	u := &Updater{OS: "linux", Arch: "amd64"}
	asset, err := u.AssetFor(rel)
	if err != nil || asset.Name != rel.Assets[0].Name {
		t.Fatalf("asset = %+v, %v", asset, err)
	}
}

func TestManaged(t *testing.T) {
	cases := map[string]string{
		"/opt/bin/ext-tags": "opkg",
		"/opt/homebrew/Cellar/ext-tags/1.0/bin/ext-tags":      "Homebrew",
		"C:/Users/u/scoop/apps/ext-tags/current/ext-tags.exe": "Scoop",
		"/usr/bin/ext-tags":       "apt",
		"/usr/local/bin/ext-tags": "",
	}
	for path, want := range cases {
		got, _ := Managed(path)
		if got != want {
			t.Errorf("Managed(%q) = %q, want %q", path, got, want)
		}
	}
}

func TestExtractTarAndZip(t *testing.T) {
	var tarBuf bytes.Buffer
	gz := gzip.NewWriter(&tarBuf)
	tw := tar.NewWriter(gz)
	_ = tw.WriteHeader(&tar.Header{Name: "ext-tags_linux_amd64", Mode: 0755, Size: 3})
	_, _ = tw.Write([]byte("abc"))
	_ = tw.Close()
	_ = gz.Close()
	data, err := extractBinary("x.tar.gz", tarBuf.Bytes())
	if err != nil || string(data) != "abc" {
		t.Fatalf("tar = %q, %v", data, err)
	}

	var zipBuf bytes.Buffer
	zw := zip.NewWriter(&zipBuf)
	f, _ := zw.Create("ext-tags_windows_amd64.exe")
	_, _ = f.Write([]byte("xyz"))
	_ = zw.Close()
	data, err = extractBinary("x.zip", zipBuf.Bytes())
	if err != nil || string(data) != "xyz" {
		t.Fatalf("zip = %q, %v", data, err)
	}
}

func TestApply(t *testing.T) {
	target := filepath.Join(t.TempDir(), "ext-tags")
	if err := os.WriteFile(target, []byte("old"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := Apply(target, []byte("new"), 0755); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(target)
	if err != nil || string(data) != "new" {
		t.Fatalf("data = %q, %v", data, err)
	}
}

func TestCheckAndFetch(t *testing.T) {
	archive := tarBytes(t, "ext-tags_0.10.4_linux_amd64", []byte("binary"))
	sum := sha256.Sum256(archive)
	sums := []byte(hex.EncodeToString(sum[:]) + "  ext-tags_0.10.4_linux_amd64.tar.gz\n")
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/release":
			_ = json.NewEncoder(w).Encode(Release{TagName: "v0.10.4", Assets: []Asset{
				{Name: "ext-tags_0.10.4_linux_amd64.tar.gz", URL: srv.URL + "/archive"},
				{Name: "SHA256SUMS", URL: srv.URL + "/sums"},
			}})
		case "/sums":
			_, _ = w.Write(sums)
		case "/archive":
			_, _ = w.Write(archive)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()
	u := &Updater{Client: &http.Client{Timeout: 5 * time.Second}, APIURL: srv.URL + "/release", OS: "linux", Arch: "amd64", UserAgent: "test"}
	rel, err := u.Check()
	if err != nil {
		t.Fatal(err)
	}
	asset, _ := u.AssetFor(rel)
	bin, err := u.FetchBinary(rel, asset)
	if err != nil || string(bin) != "binary" {
		t.Fatalf("bin = %q, %v", bin, err)
	}
}

func tarBytes(t *testing.T, name string, data []byte) []byte {
	t.Helper()
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	tw := tar.NewWriter(gz)
	if err := tw.WriteHeader(&tar.Header{Name: name, Mode: 0755, Size: int64(len(data))}); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write(data); err != nil {
		t.Fatal(err)
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}
