// deeix-ocr: a small OCR service for DEEIX Chat built on the Tesseract command line.
//
// It speaks the protocol of the DEEIX "tesseract" OCR engine:
//
//	GET  /healthz  -> {"status":"ok",...}
//	POST /ocr      multipart: file, page_ranges ("1-3,5"), prompt (ignored)
//	               -> {"rendered_pages":N,"pages":[{"page_number":1,"text":"..."}]}
//	               -> 204 when no text was found
//
// It is written for a small VPS: one job at a time, one page at a time, one Tesseract
// thread, low CPU and disk priority (set by the entrypoint), page and size limits, and
// a time budget after which it returns the pages recognised so far.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

type settings struct {
	Addr        string
	Lang        string
	PSM         string
	DPI         int
	MaxSide     int
	MaxPages    int
	MaxFileMB   int
	MaxPixels   int
	MaxQueue    int
	PageTimeout time.Duration
	Budget      time.Duration
	Token       string
	TempDir     string
}

func loadSettings() settings {
	return settings{
		Addr:        envString("OCR_ADDR", ":8004"),
		Lang:        envString("OCR_LANG", "rus+eng"),
		PSM:         envString("OCR_PSM", "3"),
		DPI:         envInt("OCR_DPI", 200, 72, 400),
		MaxSide:     envInt("OCR_MAX_SIDE", 3000, 800, 8000),
		MaxPages:    envInt("OCR_MAX_PAGES", 20, 1, 500),
		MaxFileMB:   envInt("OCR_MAX_FILE_MB", 25, 1, 500),
		MaxPixels:   envInt("OCR_MAX_PIXELS", 6_000_000, 500_000, 40_000_000),
		MaxQueue:    envInt("OCR_MAX_QUEUE", 8, 0, 100),
		PageTimeout: time.Duration(envInt("OCR_PAGE_TIMEOUT_SECONDS", 120, 10, 3600)) * time.Second,
		Budget:      time.Duration(envInt("OCR_BUDGET_SECONDS", 540, 10, 7200)) * time.Second,
		Token:       strings.TrimSpace(os.Getenv("OCR_TOKEN")),
		TempDir:     envString("OCR_TMP", os.TempDir()),
	}
}

func envString(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}

func envInt(key string, def, min, max int) int {
	v, err := strconv.Atoi(strings.TrimSpace(os.Getenv(key)))
	if err != nil {
		return def
	}
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

type server struct {
	cfg     settings
	slot    chan struct{} // one OCR job at a time
	waiting atomic.Int32
	// runner executes external commands; replaced in tests.
	runner func(ctx context.Context, name string, args ...string) ([]byte, error)
}

func newServer(cfg settings) *server {
	return &server{cfg: cfg, slot: make(chan struct{}, 1), runner: runCommand}
}

func runCommand(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	// One Tesseract thread: several threads cost much more memory for little gain on a small VPS.
	cmd.Env = append(os.Environ(), "OMP_THREAD_LIMIT=1")
	var stderr strings.Builder
	cmd.Stderr = &limitedWriter{w: &stderr, n: 2048}
	out, err := cmd.Output()
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, fmt.Errorf("%s: %w: %s", name, err, strings.TrimSpace(stderr.String()))
	}
	return out, nil
}

type limitedWriter struct {
	w io.Writer
	n int
}

func (l *limitedWriter) Write(p []byte) (int, error) {
	if l.n > 0 {
		k := min(len(p), l.n)
		_, _ = l.w.Write(p[:k])
		l.n -= k
	}
	return len(p), nil
}

func (s *server) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", s.handleHealth)
	mux.HandleFunc("/ocr", s.handleOCR)
	return mux
}

func (s *server) authorized(r *http.Request) bool {
	if s.cfg.Token == "" {
		return true
	}
	auth := strings.TrimSpace(r.Header.Get("Authorization"))
	return auth == "Bearer "+s.cfg.Token || strings.TrimSpace(r.Header.Get("X-API-Key")) == s.cfg.Token
}

// handleHealth needs no token: it tells nothing about the files.
func (s *server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status":    "ok",
		"engine":    "tesseract",
		"lang":      s.cfg.Lang,
		"dpi":       s.cfg.DPI,
		"max_pages": s.cfg.MaxPages,
		"busy":      len(s.slot) > 0,
		"waiting":   s.waiting.Load(),
	})
}

type pageResult struct {
	PageNumber int    `json:"page_number"`
	Text       string `json:"text"`
}

type httpError struct {
	status int
	msg    string
}

func (e *httpError) Error() string { return e.msg }

func (s *server) handleOCR(w http.ResponseWriter, r *http.Request) {
	started := time.Now()
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !s.authorized(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	work, err := os.MkdirTemp(s.cfg.TempDir, "deeix-ocr-*")
	if err != nil {
		http.Error(w, "temp dir: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer os.RemoveAll(work)

	upload, pageRanges, err := s.receive(w, r, work)
	if err != nil {
		var he *httpError
		if errors.As(err, &he) {
			http.Error(w, he.msg, he.status)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Queue: wait for the single slot, but refuse when too many jobs are already waiting.
	if s.waiting.Add(1) > int32(s.cfg.MaxQueue)+1 {
		s.waiting.Add(-1)
		http.Error(w, "ocr busy, try again later", http.StatusServiceUnavailable)
		return
	}
	select {
	case s.slot <- struct{}{}:
		s.waiting.Add(-1)
	case <-r.Context().Done():
		s.waiting.Add(-1)
		return
	}
	defer func() { <-s.slot }()

	ctx, cancel := context.WithDeadline(r.Context(), started.Add(s.cfg.Budget))
	defer cancel()

	pages, err := s.recognize(ctx, upload, pageRanges, work)
	if err != nil && len(pages) == 0 {
		if r.Context().Err() != nil {
			return
		}
		var he *httpError
		if errors.As(err, &he) {
			http.Error(w, he.msg, he.status)
		} else {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		}
		log.Printf("ocr %s: failed after %s: %v", filepath.Base(upload.name), time.Since(started).Round(time.Millisecond), err)
		return
	}
	found := false
	for _, p := range pages {
		if strings.TrimSpace(p.Text) != "" {
			found = true
			break
		}
	}
	log.Printf("ocr %s: %d page(s) in %s (text: %v, partial error: %v)", filepath.Base(upload.name), len(pages), time.Since(started).Round(time.Millisecond), found, err)
	if !found {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"rendered_pages": len(pages), "pages": pages})
}

type uploadedFile struct {
	path string
	name string
	mime string
}

// receive streams the multipart body to disk without holding the file in memory.
func (s *server) receive(w http.ResponseWriter, r *http.Request, work string) (uploadedFile, string, error) {
	limit := int64(s.cfg.MaxFileMB)<<20 + 1<<20
	r.Body = http.MaxBytesReader(w, r.Body, limit)
	reader, err := r.MultipartReader()
	if err != nil {
		return uploadedFile{}, "", &httpError{http.StatusBadRequest, "expected multipart/form-data"}
	}
	var file uploadedFile
	pageRanges := ""
	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			var tooBig *http.MaxBytesError
			if errors.As(err, &tooBig) {
				return uploadedFile{}, "", &httpError{http.StatusRequestEntityTooLarge, fmt.Sprintf("file is larger than %d MB", s.cfg.MaxFileMB)}
			}
			return uploadedFile{}, "", err
		}
		switch part.FormName() {
		case "file":
			ext := strings.ToLower(filepath.Ext(part.FileName()))
			if len(ext) > 6 {
				ext = ""
			}
			path := filepath.Join(work, "input"+ext)
			out, err := os.Create(path)
			if err != nil {
				return uploadedFile{}, "", err
			}
			_, copyErr := io.Copy(out, part)
			closeErr := out.Close()
			if copyErr != nil {
				var tooBig *http.MaxBytesError
				if errors.As(copyErr, &tooBig) {
					return uploadedFile{}, "", &httpError{http.StatusRequestEntityTooLarge, fmt.Sprintf("file is larger than %d MB", s.cfg.MaxFileMB)}
				}
				return uploadedFile{}, "", copyErr
			}
			if closeErr != nil {
				return uploadedFile{}, "", closeErr
			}
			file = uploadedFile{path: path, name: part.FileName(), mime: part.Header.Get("Content-Type")}
		case "page_ranges":
			b, _ := io.ReadAll(io.LimitReader(part, 4096))
			pageRanges = string(b)
		default:
			_, _ = io.Copy(io.Discard, io.LimitReader(part, 1<<20))
		}
		_ = part.Close()
	}
	if file.path == "" {
		return uploadedFile{}, "", &httpError{http.StatusBadRequest, "missing file"}
	}
	return file, pageRanges, nil
}

func (s *server) recognize(ctx context.Context, file uploadedFile, pageRanges, work string) ([]pageResult, error) {
	head := make([]byte, 8)
	f, err := os.Open(file.path)
	if err != nil {
		return nil, err
	}
	n, _ := io.ReadFull(f, head)
	_ = f.Close()
	head = head[:n]
	if strings.HasPrefix(string(head), "%PDF") || strings.EqualFold(filepath.Ext(file.path), ".pdf") {
		return s.recognizePDF(ctx, file.path, pageRanges, work)
	}
	text, err := s.recognizeImage(ctx, file.path, work)
	if err != nil {
		return nil, err
	}
	return []pageResult{{PageNumber: 1, Text: text}}, nil
}

func (s *server) recognizePDF(ctx context.Context, path, pageRanges, work string) ([]pageResult, error) {
	info, err := s.runner(ctx, "pdfinfo", path)
	if err != nil {
		return nil, &httpError{http.StatusUnprocessableEntity, "not a readable PDF: " + err.Error()}
	}
	total := parsePDFPageCount(string(info))
	if total <= 0 {
		return nil, &httpError{http.StatusUnprocessableEntity, "PDF has no pages"}
	}
	pages := parsePageRanges(pageRanges, total)
	if len(pages) > s.cfg.MaxPages {
		pages = pages[:s.cfg.MaxPages]
	}
	results := make([]pageResult, 0, len(pages))
	var lastErr error
	for _, page := range pages {
		if ctx.Err() != nil {
			// Time budget is over: return what is ready instead of losing everything.
			if lastErr == nil {
				lastErr = fmt.Errorf("time budget reached before page %d", page)
			}
			break
		}
		text, err := s.recognizePDFPage(ctx, path, page, work)
		if err != nil {
			lastErr = fmt.Errorf("page %d: %w", page, err)
			log.Printf("ocr page %d: %v", page, err)
			continue
		}
		results = append(results, pageResult{PageNumber: page, Text: text})
	}
	if len(results) == 0 && lastErr != nil {
		return nil, lastErr
	}
	return results, lastErr
}

func (s *server) recognizePDFPage(ctx context.Context, path string, page int, work string) (string, error) {
	pageCtx, cancel := context.WithTimeout(ctx, s.cfg.PageTimeout)
	defer cancel()
	dpi := s.cfg.DPI
	p := strconv.Itoa(page)
	if info, err := s.runner(pageCtx, "pdfinfo", "-f", p, "-l", p, path); err == nil {
		if w, h := parsePDFPageSize(string(info), page); w > 0 && h > 0 {
			dpi = pageDPI(w, h, s.cfg.DPI, s.cfg.MaxSide)
		}
	}
	base := filepath.Join(work, "page")
	if _, err := s.runner(pageCtx, "pdftoppm", "-r", strconv.Itoa(dpi), "-gray", "-png", "-f", p, "-l", p, "-singlefile", path, base); err != nil {
		return "", err
	}
	png := base + ".png"
	defer os.Remove(png)
	return s.tesseract(pageCtx, png)
}

func (s *server) tesseract(ctx context.Context, imagePath string) (string, error) {
	out, err := s.runner(ctx, "tesseract", imagePath, "stdout", "-l", s.cfg.Lang, "--oem", "1", "--psm", s.cfg.PSM)
	if err != nil {
		return "", err
	}
	return cleanText(string(out)), nil
}

var pagesLine = regexp.MustCompile(`(?m)^Pages:\s+(\d+)`)

func parsePDFPageCount(info string) int {
	m := pagesLine.FindStringSubmatch(info)
	if m == nil {
		return 0
	}
	n, _ := strconv.Atoi(m[1])
	return n
}

var pageSizeLine = regexp.MustCompile(`(?m)^Page\s+(\d+)\s+size:\s+([\d.]+)\s+x\s+([\d.]+)`)
var plainSizeLine = regexp.MustCompile(`(?m)^Page size:\s+([\d.]+)\s+x\s+([\d.]+)`)

// parsePDFPageSize reads the page size in points from pdfinfo output; /Rotate does not matter
// because only the longer side is used.
func parsePDFPageSize(info string, page int) (float64, float64) {
	for _, m := range pageSizeLine.FindAllStringSubmatch(info, -1) {
		if n, _ := strconv.Atoi(m[1]); n == page {
			w, _ := strconv.ParseFloat(m[2], 64)
			h, _ := strconv.ParseFloat(m[3], 64)
			return w, h
		}
	}
	if m := plainSizeLine.FindStringSubmatch(info); m != nil {
		w, _ := strconv.ParseFloat(m[1], 64)
		h, _ := strconv.ParseFloat(m[2], 64)
		return w, h
	}
	return 0, 0
}

// pageDPI lowers the resolution for oversized pages so the longer side stays within maxSide pixels.
func pageDPI(widthPt, heightPt float64, dpi, maxSide int) int {
	longest := max(widthPt, heightPt)
	if longest <= 0 {
		return dpi
	}
	limit := int(float64(maxSide) * 72 / longest)
	if limit < dpi {
		dpi = limit
	}
	return max(dpi, 36)
}

// parsePageRanges follows the DEEIX format "1-3,5"; empty or invalid means every page.
func parsePageRanges(raw string, total int) []int {
	if total <= 0 {
		return nil
	}
	seen := map[int]bool{}
	var pages []int
	add := func(p int) {
		if p >= 1 && p <= total && !seen[p] {
			seen[p] = true
			pages = append(pages, p)
		}
	}
	for _, chunk := range strings.Split(raw, ",") {
		chunk = strings.TrimSpace(chunk)
		if chunk == "" {
			continue
		}
		if a, b, ok := strings.Cut(chunk, "-"); ok {
			start, err1 := strconv.Atoi(strings.TrimSpace(a))
			end, err2 := strconv.Atoi(strings.TrimSpace(b))
			if err1 != nil || err2 != nil || start <= 0 {
				continue
			}
			if end < start {
				end = start
			}
			for p := start; p <= min(end, total); p++ {
				add(p)
			}
			continue
		}
		if p, err := strconv.Atoi(chunk); err == nil {
			add(p)
		}
	}
	if len(pages) == 0 {
		for p := 1; p <= total; p++ {
			pages = append(pages, p)
		}
	}
	sortInts(pages)
	return pages
}

func sortInts(a []int) {
	for i := 1; i < len(a); i++ {
		for j := i; j > 0 && a[j-1] > a[j]; j-- {
			a[j-1], a[j] = a[j], a[j-1]
		}
	}
}

func cleanText(raw string) string {
	raw = strings.ReplaceAll(raw, "\f", "\n")
	lines := strings.Split(raw, "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		if v := strings.TrimSpace(line); v != "" {
			out = append(out, v)
		}
	}
	return strings.Join(out, "\n")
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", mime.FormatMediaType("application/json", map[string]string{"charset": "utf-8"}))
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func main() {
	cfg := loadSettings()
	if len(os.Args) > 1 && os.Args[1] == "healthcheck" {
		// Used by the Docker HEALTHCHECK: the image has no curl.
		resp, err := http.Get("http://127.0.0.1" + cfg.Addr[strings.LastIndex(cfg.Addr, ":"):] + "/healthz")
		if err != nil || resp.StatusCode != http.StatusOK {
			os.Exit(1)
		}
		os.Exit(0)
	}
	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           newServer(cfg).routes(),
		ReadHeaderTimeout: 30 * time.Second,
	}
	log.Printf("deeix-ocr listening on %s (lang %s, %d dpi, up to %d pages, one job at a time)", cfg.Addr, cfg.Lang, cfg.DPI, cfg.MaxPages)
	log.Fatal(srv.ListenAndServe())
}
