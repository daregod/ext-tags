// Package update implements self-update for standalone ext-tags binaries.
package update

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	githubRepo  = "daregod/ext-tags"
	maxDownload = 128 << 20
	maxBinary   = 256 << 20
)

// Release is the subset of GitHub release metadata we need.
type Release struct {
	TagName string  `json:"tag_name"`
	Assets  []Asset `json:"assets"`
}

// Asset is one GitHub release asset.
type Asset struct {
	Name string `json:"name"`
	URL  string `json:"browser_download_url"`
	Size int64  `json:"size"`
}

// Updater checks and downloads official releases.
type Updater struct {
	Client    *http.Client
	APIURL    string
	OS        string
	Arch      string
	Token     string
	UserAgent string
}

// New returns an updater for the current platform.
func New(current string) *Updater {
	return &Updater{
		Client:    &http.Client{Timeout: 2 * time.Minute},
		APIURL:    "https://api.github.com/repos/" + githubRepo + "/releases/latest",
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
		UserAgent: "ext-tags/" + current,
	}
}

// Check returns the latest release.
func (u *Updater) Check() (*Release, error) {
	req, err := http.NewRequest(http.MethodGet, u.APIURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", u.UserAgent)
	if u.Token != "" {
		req.Header.Set("Authorization", "Bearer "+u.Token)
	}
	resp, err := u.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("github releases: %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}
	var rel Release
	if err := json.NewDecoder(io.LimitReader(resp.Body, 4<<20)).Decode(&rel); err != nil {
		return nil, err
	}
	if rel.TagName == "" {
		return nil, fmt.Errorf("github release has no tag")
	}
	return &rel, nil
}

// AssetFor selects the archive for OS/architecture.
func (u *Updater) AssetFor(rel *Release) (Asset, error) {
	version := strings.TrimPrefix(rel.TagName, "v")
	ext := ".tar.gz"
	if u.OS == "windows" {
		ext = ".zip"
	}
	want := fmt.Sprintf("ext-tags_%s_%s_%s%s", version, u.OS, u.Arch, ext)
	for _, asset := range rel.Assets {
		if asset.Name == want {
			return asset, nil
		}
	}
	return Asset{}, fmt.Errorf("release %s has no asset %s", rel.TagName, want)
}

// Checksums downloads and parses SHA256SUMS.
func (u *Updater) Checksums(rel *Release) (map[string]string, error) {
	for _, asset := range rel.Assets {
		if asset.Name != "SHA256SUMS" {
			continue
		}
		data, err := u.download(asset.URL)
		if err != nil {
			return nil, err
		}
		return parseChecksums(data)
	}
	return nil, fmt.Errorf("release %s has no SHA256SUMS", rel.TagName)
}

// FetchBinary downloads the archive, verifies its checksum, and extracts the binary.
func (u *Updater) FetchBinary(rel *Release, asset Asset) ([]byte, error) {
	sums, err := u.Checksums(rel)
	if err != nil {
		return nil, err
	}
	want := sums[asset.Name]
	if want == "" {
		return nil, fmt.Errorf("no checksum for %s", asset.Name)
	}
	data, err := u.download(asset.URL)
	if err != nil {
		return nil, err
	}
	got := sha256.Sum256(data)
	if !strings.EqualFold(hex.EncodeToString(got[:]), want) {
		return nil, fmt.Errorf("checksum mismatch for %s", asset.Name)
	}
	return extractBinary(asset.Name, data)
}

func (u *Updater) download(url string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", u.UserAgent)
	if u.Token != "" {
		req.Header.Set("Authorization", "Bearer "+u.Token)
	}
	resp, err := u.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download %s: %s", url, resp.Status)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxDownload))
	if err != nil {
		return nil, err
	}
	return data, nil
}

func parseChecksums(data []byte) (map[string]string, error) {
	out := make(map[string]string)
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		name := strings.TrimPrefix(fields[1], "*")
		out[path.Base(name)] = fields[0]
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("empty SHA256SUMS")
	}
	return out, nil
}

func extractBinary(name string, data []byte) ([]byte, error) {
	if strings.HasSuffix(name, ".zip") {
		zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
		if err != nil {
			return nil, err
		}
		for _, f := range zr.File {
			if isBinaryName(f.Name) {
				rc, err := f.Open()
				if err != nil {
					return nil, err
				}
				defer rc.Close()
				return io.ReadAll(io.LimitReader(rc, maxBinary))
			}
		}
		return nil, fmt.Errorf("binary not found in %s", name)
	}
	gz, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if isBinaryName(hdr.Name) {
			return io.ReadAll(io.LimitReader(tr, maxBinary))
		}
	}
	return nil, fmt.Errorf("binary not found in %s", name)
}

func isBinaryName(name string) bool {
	base := path.Base(name)
	return strings.HasPrefix(base, "ext-tags") && !strings.HasSuffix(base, ".tag")
}

// Apply atomically replaces target with data.
func Apply(target string, data []byte, mode os.FileMode) error {
	dir := filepath.Dir(target)
	tmp, err := os.CreateTemp(dir, ".ext-tags-update-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmpName, mode.Perm()); err != nil {
		return err
	}
	if runtime.GOOS != "windows" {
		return os.Rename(tmpName, target)
	}
	old := target + ".old"
	_ = os.Remove(old)
	if err := os.Rename(target, old); err != nil {
		return err
	}
	if err := os.Rename(tmpName, target); err != nil {
		_ = os.Rename(old, target)
		return err
	}
	_ = os.Remove(old)
	return nil
}

// NormalizeVersion strips a leading v.
func NormalizeVersion(v string) string {
	return strings.TrimPrefix(strings.TrimSpace(v), "v")
}

// IsNewer reports whether latest is newer than current.
func IsNewer(latest, current string) bool {
	latest = NormalizeVersion(latest)
	current = NormalizeVersion(current)
	if current == "" || current == "0.0.0-dev" {
		return true
	}
	if latest == current {
		return false
	}
	lp := strings.Split(latest, ".")
	cp := strings.Split(current, ".")
	for i := 0; i < len(lp) || i < len(cp); i++ {
		var ln, cn int
		if i < len(lp) {
			ln, _ = strconv.Atoi(strings.Split(lp[i], "-")[0])
		}
		if i < len(cp) {
			cn, _ = strconv.Atoi(strings.Split(cp[i], "-")[0])
		}
		if ln != cn {
			return ln > cn
		}
	}
	return false
}

// Managed returns the package manager and update instructions if executable
// appears to be owned by one.
func Managed(executable string) (string, string) {
	p := strings.ToLower(filepath.ToSlash(executable))
	switch {
	case strings.HasPrefix(p, "/opt/bin/"):
		return "opkg", "opkg update && opkg upgrade ext-tags"
	case strings.Contains(p, "/cellar/") || strings.Contains(p, "/homebrew/"):
		return "Homebrew", "brew update && brew upgrade ext-tags"
	case strings.Contains(p, "/scoop/apps/"):
		return "Scoop", "scoop update && scoop update ext-tags"
	case strings.HasPrefix(p, "/usr/bin/"):
		return "apt", "sudo apt update && sudo apt upgrade ext-tags"
	}
	return "", ""
}
