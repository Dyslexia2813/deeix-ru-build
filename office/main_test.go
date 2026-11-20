package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"encoding/xml"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const sampleDocx = `{
  "file_name": "Договор: поставки/2026",
  "font": "Arial", "font_size": 11,
  "blocks": [
    {"type": "heading", "level": 1, "text": "Договор поставки № 15", "align": "center"},
    {"type": "paragraph", "text": "ООО «Ромашка» & АО «Вектор» заключили **настоящий договор** о *нижеследующем* и __важном__.\nВторая строка.", "align": "justify"},
    {"type": "heading", "level": 2, "text": "1. Предмет"},
    {"type": "list", "ordered": true, "items": ["Поставка товара", "Оплата"]},
    {"type": "list", "ordered": false, "items": ["пункт", "ещё пункт"]},
    {"type": "list", "ordered": true, "items": ["снова с единицы"]},
    {"type": "table", "widths": [1, 4, 2], "rows": [["№", "Товар", "Сумма"], ["1", {"text": "Бумага <A4>", "fill": "#FFF2CC"}, {"text": "1 250 000", "align": "right", "bold": true}]]},
    {"type": "page_break"},
    {"type": "unknown", "text": "неизвестный блок станет абзацем"}
  ]
}`

const sampleXlsx = `{
  "file_name": "Реестр",
  "sheets": [
    {"name": "Договоры: 2026", "columns": [{"title": "№"}, {"title": "Контрагент"}, {"title": "ИНН"}, {"title": "Сумма", "format": "money"}, {"title": "Дата"}, {"title": "Подписан"}],
     "rows": [
       ["Д-1", "ООО «Ромашка»", "7701234567", 1250000, "2026-11-20", true],
       ["Д-2", {"value": "АО «Вектор»", "fill": "#FCE4D6"}, "5003456789", "84 300,50", "21.11.2026", false],
       [null, "Итого", null, "=SUM(D2:D3)", null, null]
     ]},
    {"name": "Договоры: 2026", "rows": [["без заголовков", 12.5, "15%"]]}
  ]
}`

func zipParts(t *testing.T, data []byte) map[string]string {
	t.Helper()
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("not a zip: %v", err)
	}
	parts := map[string]string{}
	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			t.Fatal(err)
		}
		b, _ := io.ReadAll(rc)
		rc.Close()
		parts[f.Name] = string(b)
		if strings.HasSuffix(f.Name, ".xml") || strings.HasSuffix(f.Name, ".rels") {
			dec := xml.NewDecoder(bytes.NewReader(b))
			for {
				if _, err := dec.Token(); err == io.EOF {
					break
				} else if err != nil {
					t.Fatalf("%s is not well-formed XML: %v", f.Name, err)
				}
			}
		}
	}
	return parts
}

func TestBuildDocx(t *testing.T) {
	var args docxArgs
	if err := json.Unmarshal([]byte(sampleDocx), &args); err != nil {
		t.Fatal(err)
	}
	data, err := buildDocx(args)
	if err != nil {
		t.Fatal(err)
	}
	parts := zipParts(t, data)
	doc := parts["word/document.xml"]
	for _, want := range []string{"Heading1", `<w:b/><w:bCs/></w:rPr><w:t xml:space="preserve">настоящий договор`, "<w:i/>", `<w:u w:val="single"/>`, "<w:br/>", `w:fill="FFF2CC"`, `w:fill="E7EEF3"`, "<w:tblHeader/>", `w:type="page"`, "неизвестный блок", "&amp;", "&lt;A4&gt;", `w:numId w:val="3"`, "footerReference"} {
		if !strings.Contains(doc, want) {
			t.Errorf("document.xml has no %q", want)
		}
	}
	if !strings.Contains(parts["word/styles.xml"], `w:ascii="Arial"`) || !strings.Contains(parts["word/styles.xml"], `<w:sz w:val="22"/>`) {
		t.Error("font or size not applied")
	}
	if !strings.Contains(parts["word/footer1.xml"], "PAGE") {
		t.Error("no page number field")
	}
	if os.Getenv("OFFICE_SAMPLES") != "" {
		_ = os.WriteFile(filepath.Join(os.Getenv("OFFICE_SAMPLES"), "sample.docx"), data, 0o644)
	}
}

func TestBuildXlsx(t *testing.T) {
	var args xlsxArgs
	if err := json.Unmarshal([]byte(sampleXlsx), &args); err != nil {
		t.Fatal(err)
	}
	data, err := buildXlsx(args)
	if err != nil {
		t.Fatal(err)
	}
	parts := zipParts(t, data)
	s1 := parts["xl/worksheets/sheet1.xml"]
	for _, want := range []string{`state="frozen"`, `<autoFilter ref="A1:F4"/>`, `t="inlineStr"><is><t xml:space="preserve">7701234567`, `<v>1250000</v>`, `<v>84300.5</v>`, `<f>SUM(D2:D3)</f>`, `t="b"><v>1</v>`, `<v>46346</v>`} {
		if !strings.Contains(s1, want) {
			t.Errorf("sheet1.xml has no %q", want)
		}
	}
	if !strings.Contains(parts["xl/workbook.xml"], `name="Договоры  2026 (2)"`) && !strings.Contains(parts["xl/workbook.xml"], `name="Договоры 2026 (2)"`) {
		t.Errorf("duplicate sheet name not made unique: %s", parts["xl/workbook.xml"])
	}
	if !strings.Contains(parts["xl/styles.xml"], `rgb="FFFCE4D6"`) {
		t.Error("cell fill missing")
	}
	s2 := parts["xl/worksheets/sheet2.xml"]
	if !strings.Contains(s2, `<v>0.15</v>`) || strings.Contains(s2, "autoFilter") || strings.Contains(s2, "frozen") {
		t.Errorf("sheet2 wrong: %s", s2)
	}
	if os.Getenv("OFFICE_SAMPLES") != "" {
		_ = os.WriteFile(filepath.Join(os.Getenv("OFFICE_SAMPLES"), "sample.xlsx"), data, 0o644)
	}
}

func TestParseNumber(t *testing.T) {
	cases := map[string]bool{"1 234,56": true, "1234.5": true, "-7": true, "0,5": true, "1,234": false, "007": false, "7701234567": false, "12%": false, "abc": false}
	for in, ok := range cases {
		if _, got := parseNumber(in); got != ok {
			t.Errorf("parseNumber(%q) = %v, want %v", in, got, ok)
		}
	}
}

func TestInline(t *testing.T) {
	runs := parseInline("a **b** *c* __d__ 2*3")
	var bold, italic, under bool
	for _, r := range runs {
		bold = bold || (r.Bold && r.Text == "b")
		italic = italic || (r.Italic && r.Text == "c")
		under = under || (r.Underline && r.Text == "d")
	}
	if !bold || !italic || !under {
		t.Errorf("runs = %+v", runs)
	}
	if got := plainText("цена 2*3 руб."); got != "цена 2*3 руб." {
		t.Errorf("single star changed: %q", got)
	}
}

func TestCleanFileName(t *testing.T) {
	if got := cleanFileName(`Договор: "поставки"/2026.docx`, ".docx", "Документ"); got != "Договор поставки 2026.docx" {
		t.Errorf("got %q", got)
	}
	if got := cleanFileName("  ", ".xlsx", "Таблица"); got != "Таблица.xlsx" {
		t.Errorf("got %q", got)
	}
}

func TestMCPAndDownload(t *testing.T) {
	dir := t.TempDir()
	cfg := config{Dir: dir, LinkPrefix: "/api/v1/office-files", TTL: time.Hour, MaxBytes: 10 << 20}
	srv := &server{cfg: cfg, store: &store{cfg: cfg}}
	mux := http.NewServeMux()
	mux.HandleFunc("/mcp", srv.handleMCP)
	mux.HandleFunc("/files/", srv.handleFile)
	ts := httptest.NewServer(mux)
	defer ts.Close()

	post := func(body string) (*http.Response, map[string]any) {
		resp, err := http.Post(ts.URL+"/mcp", "application/json", strings.NewReader(body))
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		var out map[string]any
		_ = json.NewDecoder(resp.Body).Decode(&out)
		return resp, out
	}
	resp, out := post(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-03-26"}}`)
	if resp.Header.Get("Mcp-Session-Id") == "" || out["result"] == nil {
		t.Fatalf("initialize: %v", out)
	}
	if resp, _ := post(`{"jsonrpc":"2.0","method":"notifications/initialized"}`); resp.StatusCode != http.StatusAccepted {
		t.Fatalf("notification status %d", resp.StatusCode)
	}
	_, out = post(`{"jsonrpc":"2.0","id":2,"method":"tools/list"}`)
	tools := out["result"].(map[string]any)["tools"].([]any)
	if len(tools) != 2 {
		t.Fatalf("tools: %v", tools)
	}
	call, _ := json.Marshal(map[string]any{"jsonrpc": "2.0", "id": 3, "method": "tools/call", "params": map[string]any{
		"name": "make_xlsx", "arguments": json.RawMessage(sampleXlsx), "_meta": map[string]any{"user_id": 7},
	}})
	_, out = post(string(call))
	result := out["result"].(map[string]any)
	if result["isError"] == true {
		t.Fatalf("tool error: %v", result)
	}
	text := result["content"].([]any)[0].(map[string]any)["text"].(string)
	var payload map[string]any
	if err := json.Unmarshal([]byte(text), &payload); err != nil {
		t.Fatal(err)
	}
	link := payload["url"].(string)
	if !strings.HasPrefix(link, "/api/v1/office-files/") || !strings.Contains(payload["markdown"].(string), link) {
		t.Fatalf("payload: %v", payload)
	}
	dl, err := http.Get(ts.URL + "/files/" + strings.TrimPrefix(link, "/api/v1/office-files/"))
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(dl.Body)
	dl.Body.Close()
	if dl.StatusCode != 200 || !strings.Contains(dl.Header.Get("Content-Disposition"), "filename*=UTF-8''%D0%A0") || len(body) < 100 {
		t.Fatalf("download: %d %v", dl.StatusCode, dl.Header)
	}
	if nf, _ := http.Get(ts.URL + "/files/0123456789abcdef0123456789abcdef/x.xlsx"); nf.StatusCode != 404 {
		t.Fatalf("unknown token: %d", nf.StatusCode)
	}

	// Broken arguments come back as a tool error, not a crash.
	_, out = post(`{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"make_docx","arguments":{"file_name":"x","blocks":[]}}}`)
	if out["result"].(map[string]any)["isError"] != true {
		t.Fatalf("empty doc accepted: %v", out)
	}
}

func TestCleanupTTL(t *testing.T) {
	dir := t.TempDir()
	cfg := config{Dir: dir, TTL: time.Hour, MaxBytes: 10 << 20}
	st := &store{cfg: cfg}
	token, err := st.save("a.docx", []byte("data"), "1")
	if err != nil {
		t.Fatal(err)
	}
	old, _ := json.Marshal(fileMeta{Name: "a.docx", Created: time.Now().Add(-2 * time.Hour)})
	_ = os.WriteFile(filepath.Join(dir, token+".json"), old, 0o600)
	st.cleanup()
	if _, err := os.Stat(filepath.Join(dir, token+".bin")); !os.IsNotExist(err) {
		t.Fatal("expired file kept")
	}
}
