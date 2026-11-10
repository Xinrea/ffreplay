//go:build js && wasm

package asset

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"path"
	"strings"
	"sync"
	"syscall/js"
	"time"
)

// In the browser build the asset tree is served as individual static files
// instead of being embedded in the wasm module. That keeps the module small
// and lets browsers and CDNs cache images and data files across releases.
var AssetFS fs.FS = newWebFS()

// A stalled request must not block package initialization or a frame forever.
var webClient = &http.Client{Timeout: 5 * time.Minute}

type webFS struct {
	base    *url.URL
	entries sync.Map // name -> *webFSEntry
}

type webFSEntry struct {
	once sync.Once
	data []byte
	err  error
}

func newWebFS() *webFS {
	base := js.Global().Get("document").Get("baseURI").String()
	if base == "" {
		base = js.Global().Get("location").Get("href").String()
	}

	u, err := url.Parse(base)
	if err != nil {
		u = &url.URL{Path: "/"}
	}
	// Asset names are relative to the directory of the page.
	u.Path = path.Dir(u.Path)
	if !strings.HasSuffix(u.Path, "/") {
		u.Path += "/"
	}
	u.RawQuery = ""
	u.Fragment = ""

	return &webFS{base: u}
}

func (w *webFS) Open(name string) (fs.File, error) {
	if !fs.ValidPath(name) {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrInvalid}
	}

	data, err := w.ReadFile(name)
	if err != nil {
		return nil, &fs.PathError{Op: "open", Path: name, Err: err}
	}

	return &webFile{name: path.Base(name), Reader: bytes.NewReader(data)}, nil
}

func (w *webFS) ReadFile(name string) ([]byte, error) {
	if !fs.ValidPath(name) {
		return nil, &fs.PathError{Op: "read", Path: name, Err: fs.ErrInvalid}
	}

	value, _ := w.entries.LoadOrStore(name, &webFSEntry{})
	entry := value.(*webFSEntry)

	entry.once.Do(func() {
		entry.data, entry.err = w.fetch(name)
	})
	if entry.err != nil {
		return nil, entry.err
	}

	// Return a copy so callers cannot mutate the cached bytes.
	return bytes.Clone(entry.data), nil
}

func (w *webFS) fetch(name string) ([]byte, error) {
	u := w.base.ResolveReference(&url.URL{Path: name})

	// Transient connection failures happen under load and on flaky networks;
	// a texture that fails here would otherwise take the whole replay down.
	const attempts = 3

	var lastErr error

	for attempt := range attempts {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt) * 200 * time.Millisecond)
		}

		data, err := w.fetchOnce(u)
		if err == nil {
			return data, nil
		}
		if errors.Is(err, errHTTPStatus) {
			return nil, err
		}
		lastErr = err
	}

	return nil, lastErr
}

var errHTTPStatus = errors.New("unexpected HTTP status")

func (w *webFS) fetchOnce(u *url.URL) ([]byte, error) {
	resp, err := webClient.Get(u.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s: %w: %d", u.String(), errHTTPStatus, resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

type webFile struct {
	name string
	*bytes.Reader
}

func (f *webFile) Stat() (fs.FileInfo, error) { return webFileInfo{name: f.name, size: f.Size()}, nil }

func (f *webFile) Close() error { return nil }

type webFileInfo struct {
	name string
	size int64
}

func (i webFileInfo) Name() string       { return i.name }
func (i webFileInfo) Size() int64        { return i.size }
func (i webFileInfo) Mode() fs.FileMode  { return 0o444 }
func (i webFileInfo) ModTime() time.Time { return time.Time{} }
func (i webFileInfo) IsDir() bool        { return false }
func (i webFileInfo) Sys() any           { return nil }
