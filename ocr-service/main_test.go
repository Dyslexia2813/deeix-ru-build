package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestParsePageRanges(t *testing.T) {
	cases := []struct {
		raw   string
		total int
		want  []int
	}{
		{"", 3, []int{1, 2, 3}},
		{"2-3,1", 5, []int{1, 2, 3}},
		{"4-9", 5, []int{4, 5}},
		{"7", 5, []int{1, 2, 3, 4, 5}}, // nothing valid -> every page
		{"3,3,2-2", 5, []int{2, 3}},
	}
	for _, c := range cases {
		if got := parsePageRanges(c.raw, c.total); !reflect.DeepEqual(got, c.want) {
			t.Errorf("parsePageRanges(%q, %d) = %v, want %v", c.raw, c.total, got, c.want)
		}
	}
}

func TestPDFInfoParsing(t *testing.T) {
	info := "Title: x\nPages:           12\nPage    3 size: 595.276 x 841.89 pts (A4)\n"
	if n := parsePDFPageCount(info); n != 12 {
		t.Fatalf("pages = %d", n)
	}
	if w, h := parsePDFPageSize(info, 3); w < 595 || h < 841 {
		t.Fatalf("size = %v x %v", w, h)
	}
	if w, _ := parsePDFPageSize("Page size:      612 x 792 pts (letter)\n", 1); w != 612 {
		t.Fatalf("plain size = %v", w)
	}
}

func TestPageDPI(t *testing.T) {
	if d := pageDPI(595, 842, 200, 3000); d != 200 {
		t.Fatalf("A4 keeps 200 dpi, got %d", d)
	}
	// A huge drawing (A0) is rendered at a lower resolution: longer side <= 3000 px.
	if d := pageDPI(2384, 3370, 200, 3000); d*3370/72 > 3000 {
		t.Fatalf("A0 at %d dpi exceeds the side limit", d)
	}
}

func TestTargetSize(t *testing.T) {
	if w, h := targetSize(4000, 3000, 6_000_000); w*h > 6_100_000 || w < h {
		t.Fatalf("12 MP photo -> %dx%d", w, h)
	}
	if w, _ := targetSize(700, 300, 6_000_000); w != minLongSide {
		t.Fatalf("small screenshot is enlarged to %d, got %d", minLongSide, w)
	}
	if w, _ := targetSize(200, 100, 6_000_000); w != 600 {
		t.Fatalf("enlargement is capped at 3x, got %d", w)
	}
}

func TestExifOrientation(t *testing.T) {
	tiff := make([]byte, 8+2+12)
	copy(tiff, "II*\x00")
	binary.LittleEndian.PutUint32(tiff[4:], 8)
	binary.LittleEndian.PutUint16(tiff[8:], 1)
	binary.LittleEndian.PutUint16(tiff[10:], 0x0112)
	binary.LittleEndian.PutUint16(tiff[12:], 3)
	binary.LittleEndian.PutUint32(tiff[14:], 1)
	binary.LittleEndian.PutUint16(tiff[18:], 6)
	if o := exifOrientation(append([]byte("Exif\x00\x00"), tiff...)); o != 6 {
		t.Fatalf("orientation = %d, want 6", o)
	}
	if x, y := orientSource(0, 0, 40, 30, 6); x != 0 || y != 29 {
		t.Fatalf("orientation 6 maps the top-left corner to (%d,%d)", x, y)
	}
}

func multipartPDF(t *testing.T) (*bytes.Buffer, string) {
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	fw, _ := w.CreateFormFile("file", "scan.pdf")
	_, _ = fw.Write([]byte("%PDF-1.4 fake"))
	_ = w.WriteField("page_ranges", "1-2")
	_ = w.Close()
	return &body, w.FormDataContentType()
}

// fakeRunner imitates pdfinfo/pdftoppm/tesseract and records how many OCR runs overlap.
type fakeRunner struct {
	mu      sync.Mutex
	running int
	peak    int
}

func (f *fakeRunner) run(ctx context.Context, name string, args ...string) ([]byte, error) {
	switch name {
	case "pdfinfo":
		return []byte("Pages: 3\nPage    1 size: 595 x 842 pts\n"), nil
	case "pdftoppm":
		return nil, nil
	}
	f.mu.Lock()
	f.running++
	f.peak = max(f.peak, f.running)
	f.mu.Unlock()
	time.Sleep(20 * time.Millisecond)
	f.mu.Lock()
	f.running--
	f.mu.Unlock()
	return []byte("Текст страницы\n\n\f"), nil
}

func TestOCRProtocolAndOneJobAtATime(t *testing.T) {
	cfg := loadSettings()
	cfg.TempDir = t.TempDir()
	s := newServer(cfg)
	fake := &fakeRunner{}
	s.runner = fake.run
	ts := httptest.NewServer(s.routes())
	defer ts.Close()

	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			body, ct := multipartPDF(t)
			resp, err := http.Post(ts.URL+"/ocr", ct, body)
			if err != nil {
				t.Error(err)
				return
			}
			defer resp.Body.Close()
			var buf bytes.Buffer
			_, _ = buf.ReadFrom(resp.Body)
			if resp.StatusCode != 200 || !strings.Contains(buf.String(), `"page_number":2`) || !strings.Contains(buf.String(), "Текст страницы") {
				t.Errorf("status %d body %s", resp.StatusCode, buf.String())
			}
		}()
	}
	wg.Wait()
	if fake.peak != 1 {
		t.Fatalf("OCR runs overlapped: peak %d", fake.peak)
	}
}

func TestTokenAndLimits(t *testing.T) {
	cfg := loadSettings()
	cfg.TempDir = t.TempDir()
	cfg.Token = "secret"
	cfg.MaxFileMB = 1
	s := newServer(cfg)
	s.runner = (&fakeRunner{}).run
	ts := httptest.NewServer(s.routes())
	defer ts.Close()

	body, ct := multipartPDF(t)
	resp, _ := http.Post(ts.URL+"/ocr", ct, body)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("no token: status %d", resp.StatusCode)
	}
	resp, _ = http.Get(ts.URL + "/healthz")
	if resp.StatusCode != 200 {
		t.Fatalf("healthz must stay open: %d", resp.StatusCode)
	}

	var big bytes.Buffer
	w := multipart.NewWriter(&big)
	fw, _ := w.CreateFormFile("file", "big.pdf")
	_, _ = fw.Write(bytes.Repeat([]byte("x"), 3<<20))
	_ = w.Close()
	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/ocr", &big)
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("Authorization", "Bearer secret")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusRequestEntityTooLarge {
		t.Fatalf("oversized file: status %d", resp.StatusCode)
	}
}
