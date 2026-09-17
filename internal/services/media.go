//go:build windows

package services

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

// Local media server for the screensaver/peek pages: WebView2 can block
// file:// subresources when the app is served from a custom scheme, so we
// serve allow-listed paths over loopback HTTP instead.

var (
	mediaOnce   sync.Once
	mediaPort   int
	mediaMu     sync.Mutex
	mediaPaths  []string // index → absolute path
)

func startMediaServer() {
	mediaOnce.Do(func() {
		ln, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			return
		}
		mediaPort = ln.Addr().(*net.TCPAddr).Port
		mux := http.NewServeMux()
		mux.HandleFunc("/m/", func(w http.ResponseWriter, r *http.Request) {
			var idx int
			if _, err := fmt.Sscanf(r.URL.Path, "/m/%d", &idx); err != nil {
				http.NotFound(w, r)
				return
			}
			mediaMu.Lock()
			var p string
			if idx >= 0 && idx < len(mediaPaths) {
				p = mediaPaths[idx]
			}
			mediaMu.Unlock()
			if p == "" {
				http.NotFound(w, r)
				return
			}
			http.ServeFile(w, r, p)
		})
		go func() { _ = http.Serve(ln, mux) }()
	})
}

// mediaURL registers a path and returns its loopback URL.
func mediaURL(path string) string {
	startMediaServer()
	if mediaPort == 0 {
		return ""
	}
	mediaMu.Lock()
	defer mediaMu.Unlock()
	// reuse existing registration
	for i, p := range mediaPaths {
		if p == path {
			return fmt.Sprintf("http://127.0.0.1:%d/m/%d", mediaPort, i)
		}
	}
	mediaPaths = append(mediaPaths, path)
	return fmt.Sprintf("http://127.0.0.1:%d/m/%d", mediaPort, len(mediaPaths)-1)
}

var slideshowExts = map[string]bool{
	".png": true, ".jpg": true, ".jpeg": true, ".gif": true,
	".bmp": true, ".webp": true,
}

// ScreensaverMedia resolves the configured source into displayable URLs:
// - http(s):// sources are passed through unchanged (web pages, remote video)
// - local files get a loopback URL
// - local folders return all image files inside
func (s *PolyToolsService) ScreensaverMedia(contentType, source string) []string {
	if source == "" {
		return nil
	}
	if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
		return []string{source}
	}
	info, err := os.Stat(source)
	if err != nil {
		return nil
	}
	if !info.IsDir() {
		return []string{mediaURL(source)}
	}
	entries, err := os.ReadDir(source)
	if err != nil {
		return nil
	}
	var paths []string
	for _, e := range entries {
		if e.IsDir() || !slideshowExts[strings.ToLower(filepath.Ext(e.Name()))] {
			continue
		}
		paths = append(paths, filepath.Join(source, e.Name()))
	}
	sort.Strings(paths)
	var urls []string
	for _, p := range paths {
		urls = append(urls, mediaURL(p))
	}
	return urls
}
