package main

import (
	"archive/zip"
	"bytes"
	"strings"
	"time"
	"unicode/utf8"
)

const xmlHeader = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` + "\n"

type zipFile struct {
	Name string
	Body string
}

// writeZip packs Office Open XML parts. Deflate keeps the files small.
func writeZip(files []zipFile) ([]byte, error) {
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	stamp := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	for _, f := range files {
		h := &zip.FileHeader{Name: f.Name, Method: zip.Deflate, Modified: stamp}
		part, err := w.CreateHeader(h)
		if err != nil {
			return nil, err
		}
		if _, err := part.Write([]byte(f.Body)); err != nil {
			return nil, err
		}
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// xmlText escapes text for XML and drops characters XML 1.0 does not allow.
func xmlText(value string) string {
	var b strings.Builder
	for _, r := range value {
		if r == utf8.RuneError || (r < 32 && r != '\t' && r != '\n' && r != '\r') || r == 0xFFFE || r == 0xFFFF {
			continue
		}
		switch r {
		case '&':
			b.WriteString("&amp;")
		case '<':
			b.WriteString("&lt;")
		case '>':
			b.WriteString("&gt;")
		case '"':
			b.WriteString("&quot;")
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}
