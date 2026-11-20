// deeix-office: an MCP server for DEEIX Chat that lets a model make Word (.docx) and Excel (.xlsx) files.
//
// The model calls make_docx or make_xlsx with the content; the server builds the file, keeps it for
// OFFICE_TTL_DAYS days and answers with a download link. DEEIX (patch office-files) passes
// GET /api/v1/office-files/<token>/<name> on to GET /files/<token>/<name> here.
package main

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	protocolVersion = "2025-03-26"
	maxBody         = 4 << 20
	serverVersion   = "1.0.0"
)

type config struct {
	Addr       string
	Dir        string
	LinkPrefix string
	TTL        time.Duration
	MaxBytes   int64
}

func loadConfig() config {
	cfg := config{
		Addr:       envOr("OFFICE_ADDR", ":8095"),
		Dir:        envOr("OFFICE_DIR", "/data"),
		LinkPrefix: strings.TrimRight(envOr("OFFICE_LINK_PREFIX", "/api/v1/office-files"), "/"),
		TTL:        7 * 24 * time.Hour,
		MaxBytes:   200 << 20,
	}
	if days, err := strconv.ParseFloat(os.Getenv("OFFICE_TTL_DAYS"), 64); err == nil && days > 0 {
		cfg.TTL = time.Duration(days * float64(24*time.Hour))
	}
	if mb, err := strconv.ParseInt(os.Getenv("OFFICE_MAX_MB"), 10, 64); err == nil && mb > 0 {
		cfg.MaxBytes = mb << 20
	}
	return cfg
}

func envOr(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}

// ---- file store ----

type fileMeta struct {
	Name    string    `json:"name"`
	Created time.Time `json:"created"`
	UserID  string    `json:"user_id,omitempty"`
	Size    int       `json:"size"`
}

type store struct {
	cfg config
	mu  sync.Mutex
}

var tokenPattern = regexp.MustCompile(`^[0-9a-f]{32}$`)

func newToken() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// cleanFileName keeps letters of any script and drops characters Windows refuses.
func cleanFileName(name, ext, fallback string) string {
	name = strings.TrimSpace(name)
	lower := strings.ToLower(name)
	for _, e := range []string{".docx", ".xlsx", ".doc", ".xls"} {
		if strings.HasSuffix(lower, e) {
			name = name[:len(name)-len(e)]
			break
		}
	}
	name = strings.Map(func(r rune) rune {
		if r < 32 || strings.ContainsRune(`\/:*?"<>|`, r) {
			return ' '
		}
		return r
	}, name)
	name = strings.Trim(strings.Join(strings.Fields(name), " "), ". ")
	if runes := []rune(name); len(runes) > 100 {
		name = string(runes[:100])
	}
	if name == "" {
		name = fallback
	}
	return name + ext
}

func (s *store) save(name string, data []byte, userID string) (string, error) {
	token, err := newToken()
	if err != nil {
		return "", err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := os.WriteFile(filepath.Join(s.cfg.Dir, token+".bin"), data, 0o600); err != nil {
		return "", err
	}
	meta, _ := json.Marshal(fileMeta{Name: name, Created: time.Now().UTC(), UserID: userID, Size: len(data)})
	if err := os.WriteFile(filepath.Join(s.cfg.Dir, token+".json"), meta, 0o600); err != nil {
		_ = os.Remove(filepath.Join(s.cfg.Dir, token+".bin"))
		return "", err
	}
	s.cleanupLocked()
	return token, nil
}

func (s *store) open(token string) (*os.File, fileMeta, error) {
	var meta fileMeta
	if !tokenPattern.MatchString(token) {
		return nil, meta, os.ErrNotExist
	}
	raw, err := os.ReadFile(filepath.Join(s.cfg.Dir, token+".json"))
	if err != nil {
		return nil, meta, err
	}
	if err := json.Unmarshal(raw, &meta); err != nil {
		return nil, meta, err
	}
	if time.Since(meta.Created) > s.cfg.TTL {
		return nil, meta, os.ErrNotExist
	}
	f, err := os.Open(filepath.Join(s.cfg.Dir, token+".bin"))
	return f, meta, err
}

// cleanupLocked removes files older than the TTL, then the oldest ones while the total is over the cap.
func (s *store) cleanupLocked() {
	entries, err := os.ReadDir(s.cfg.Dir)
	if err != nil {
		return
	}
	type item struct {
		token   string
		created time.Time
		size    int64
	}
	var items []item
	var total int64
	for _, e := range entries {
		name := e.Name()
		if !strings.HasSuffix(name, ".json") {
			continue
		}
		token := strings.TrimSuffix(name, ".json")
		raw, err := os.ReadFile(filepath.Join(s.cfg.Dir, name))
		var meta fileMeta
		if err != nil || json.Unmarshal(raw, &meta) != nil || time.Since(meta.Created) > s.cfg.TTL {
			s.removeLocked(token)
			continue
		}
		info, err := os.Stat(filepath.Join(s.cfg.Dir, token+".bin"))
		if err != nil {
			s.removeLocked(token)
			continue
		}
		items = append(items, item{token, meta.Created, info.Size()})
		total += info.Size()
	}
	sort.Slice(items, func(i, j int) bool { return items[i].created.Before(items[j].created) })
	for _, it := range items {
		if total <= s.cfg.MaxBytes {
			break
		}
		s.removeLocked(it.token)
		total -= it.size
	}
	// Leftover .bin files without metadata.
	for _, e := range entries {
		if t := strings.TrimSuffix(e.Name(), ".bin"); t != e.Name() {
			if _, err := os.Stat(filepath.Join(s.cfg.Dir, t+".json")); errors.Is(err, os.ErrNotExist) {
				_ = os.Remove(filepath.Join(s.cfg.Dir, e.Name()))
			}
		}
	}
}

func (s *store) removeLocked(token string) {
	_ = os.Remove(filepath.Join(s.cfg.Dir, token+".bin"))
	_ = os.Remove(filepath.Join(s.cfg.Dir, token+".json"))
}

func (s *store) cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cleanupLocked()
}

// ---- MCP ----

type rpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type rpcResponse struct {
	JSONRPC string    `json:"jsonrpc"`
	ID      any       `json:"id"`
	Result  any       `json:"result,omitempty"`
	Error   *rpcError `json:"error,omitempty"`
}

const inlineHelp = "В тексте можно выделять: **жирный**, *курсив*, __подчёркнутый__. Перенос строки внутри абзаца — \\n."

func toolList() []map[string]any {
	cellColor := map[string]any{"type": "string", "description": "Заливка цветом #RRGGBB, например #FFF2CC (жёлтый), #E2EFDA (зелёный), #FCE4D6 (оранжевый)."}
	align := map[string]any{"type": "string", "enum": []string{"left", "center", "right", "justify"}}
	// Cells take either a plain value or an object; the schema says so in words, without anyOf,
	// which some model providers do not accept.
	docCell := map[string]any{
		"description": "Ячейка: строка с текстом или объект {\"text\": \"...\", \"fill\": \"#FFF2CC\", \"bold\": true, \"align\": \"right\"}.",
	}
	sheetCell := map[string]any{
		"description": "Ячейка: число, текст, true/false, null (пусто), дата «2026-11-20» или «20.11.2026», формула «=SUM(D2:D10)»; или объект {\"value\": ..., \"fill\": \"#FFF2CC\", \"bold\": true, \"format\": \"money\"} (format: money, percent, date, integer, number, text).",
	}
	return []map[string]any{
		{
			"name": "make_docx",
			"description": "Создаёт документ Word (.docx) и возвращает ссылку на скачивание (файл хранится 7 дней). " +
				"Вызывай, когда пользователь просит документ, договор, письмо и т.п. в виде файла Word. " +
				"Содержимое передаётся списком блоков: заголовки (уровни 1–3), абзацы, списки, таблицы (с заливкой ячеек), разрыв страницы. " +
				inlineHelp + " После вызова вставь в ответ поле markdown из результата как есть.",
			"inputSchema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"file_name":    map[string]any{"type": "string", "description": "Имя файла без расширения, например «Договор поставки»."},
					"font":         map[string]any{"type": "string", "enum": []string{"Times New Roman", "Arial", "Calibri", "Cambria", "Georgia", "Verdana", "Tahoma", "Segoe UI", "Courier New", "PT Serif", "PT Sans"}, "description": "Шрифт всего документа. По умолчанию Times New Roman."},
					"font_size":    map[string]any{"type": "number", "description": "Размер шрифта в пунктах. По умолчанию 12."},
					"margins_mm":   map[string]any{"type": "object", "description": "Поля страницы в мм. По умолчанию слева 30, справа 15, сверху и снизу 20.", "properties": map[string]any{"top": map[string]any{"type": "number"}, "bottom": map[string]any{"type": "number"}, "left": map[string]any{"type": "number"}, "right": map[string]any{"type": "number"}}},
					"page_numbers": map[string]any{"type": "boolean", "description": "Номера страниц внизу по центру. По умолчанию да."},
					"blocks": map[string]any{
						"type":        "array",
						"description": "Блоки документа по порядку.",
						"items": map[string]any{
							"type": "object",
							"properties": map[string]any{
								"type":        map[string]any{"type": "string", "enum": []string{"heading", "paragraph", "list", "table", "page_break"}},
								"text":        map[string]any{"type": "string", "description": "Текст для heading и paragraph."},
								"level":       map[string]any{"type": "integer", "minimum": 1, "maximum": 3, "description": "Уровень заголовка."},
								"align":       align,
								"ordered":     map[string]any{"type": "boolean", "description": "Для list: true — нумерованный 1. 2. 3., false — маркированный."},
								"items":       map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "Пункты списка."},
								"rows":        map[string]any{"type": "array", "items": map[string]any{"type": "array", "items": docCell}, "description": "Строки таблицы, первая — заголовок (если header не false)."},
								"header":      map[string]any{"type": "boolean", "description": "Первая строка таблицы — заголовок (жирный, с заливкой, повторяется на каждой странице). По умолчанию да."},
								"header_fill": cellColor,
								"widths":      map[string]any{"type": "array", "items": map[string]any{"type": "number"}, "description": "Относительная ширина колонок, например [1, 4, 2]."},
							},
							"required": []string{"type"},
						},
					},
				},
				"required": []string{"file_name", "blocks"},
			},
		},
		{
			"name": "make_xlsx",
			"description": "Создаёт таблицу Excel (.xlsx) и возвращает ссылку на скачивание (файл хранится 7 дней). " +
				"Вызывай, когда пользователь просит таблицу, реестр, расчёт в виде файла Excel. " +
				"Можно несколько листов; строка заголовков жирная, закреплена и с фильтром; ширина колонок подбирается по тексту. " +
				"Числа передавай числами (1250000.5), даты — «2026-11-20», формулы — строкой с «=». ИНН, номера счетов и телефонов передавай строкой. " +
				"После вызова вставь в ответ поле markdown из результата как есть.",
			"inputSchema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"file_name": map[string]any{"type": "string", "description": "Имя файла без расширения, например «Реестр договоров»."},
					"sheets": map[string]any{
						"type": "array",
						"items": map[string]any{
							"type": "object",
							"properties": map[string]any{
								"name": map[string]any{"type": "string", "description": "Название листа (до 31 знака)."},
								"columns": map[string]any{"type": "array", "description": "Колонки: заголовок, ширина в знаках (необязательно), формат (необязательно).", "items": map[string]any{
									"type": "object", "properties": map[string]any{
										"title":  map[string]any{"type": "string"},
										"width":  map[string]any{"type": "number"},
										"format": map[string]any{"type": "string", "enum": []string{"money", "percent", "date", "integer", "number", "text"}},
									}, "required": []string{"title"},
								}},
								"rows":          map[string]any{"type": "array", "items": map[string]any{"type": "array", "items": sheetCell}, "description": "Строки данных (без строки заголовков). Формулы ссылаются на ячейки с учётом заголовка в строке 1: данные начинаются со строки 2."},
								"header_fill":   cellColor,
								"freeze_header": map[string]any{"type": "boolean", "description": "Закрепить строку заголовков. По умолчанию да."},
								"filter":        map[string]any{"type": "boolean", "description": "Фильтр на строке заголовков. По умолчанию да."},
							},
							"required": []string{"rows"},
						},
					},
				},
				"required": []string{"file_name", "sheets"},
			},
		},
	}
}

type toolResult struct {
	Content []map[string]string `json:"content"`
	IsError bool                `json:"isError,omitempty"`
}

func textResult(text string, isError bool) toolResult {
	return toolResult{Content: []map[string]string{{"type": "text", "text": text}}, IsError: isError}
}

type server struct {
	cfg   config
	store *store
}

func (s *server) callTool(params json.RawMessage) toolResult {
	var call struct {
		Name      string          `json:"name"`
		Arguments json.RawMessage `json:"arguments"`
		Meta      struct {
			UserID json.RawMessage `json:"user_id"`
		} `json:"_meta"`
	}
	if err := json.Unmarshal(params, &call); err != nil {
		return textResult("Неверные параметры вызова: "+err.Error(), true)
	}
	userID := strings.Trim(string(call.Meta.UserID), `"`)
	var data []byte
	var name string
	var err error
	switch call.Name {
	case "make_docx":
		var args docxArgs
		if err = json.Unmarshal(call.Arguments, &args); err != nil {
			return textResult("Неверные аргументы make_docx: "+err.Error(), true)
		}
		name = cleanFileName(args.FileName, ".docx", "Документ")
		data, err = buildDocx(args)
	case "make_xlsx":
		var args xlsxArgs
		if err = json.Unmarshal(call.Arguments, &args); err != nil {
			return textResult("Неверные аргументы make_xlsx: "+err.Error(), true)
		}
		name = cleanFileName(args.FileName, ".xlsx", "Таблица")
		data, err = buildXlsx(args)
	default:
		return textResult("Нет такого инструмента: "+call.Name, true)
	}
	if err != nil {
		return textResult("Файл не собран: "+err.Error(), true)
	}
	token, err := s.store.save(name, data, userID)
	if err != nil {
		log.Printf("save failed: %v", err)
		return textResult("Файл собран, но не сохранён на сервере.", true)
	}
	link := fmt.Sprintf("%s/%s/%s", s.cfg.LinkPrefix, token, url.PathEscape(name))
	expires := time.Now().Add(s.cfg.TTL).UTC().Format("2006-01-02")
	log.Printf("made %s (%d bytes) for user %s", filepath.Ext(name), len(data), userID)
	out, _ := json.Marshal(map[string]any{
		"file":     name,
		"size_kb":  (len(data) + 1023) / 1024,
		"url":      link,
		"markdown": fmt.Sprintf("[📎 %s](%s)", strings.NewReplacer("[", "(", "]", ")").Replace(name), link),
		"expires":  expires,
		"note":     "Вставь поле markdown в ответ как есть, без изменений. Ссылка действует до " + expires + ".",
	})
	return textResult(string(out), false)
}

func (s *server) handleMCP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
	case http.MethodGet:
		// No server-initiated stream.
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	case http.MethodDelete:
		w.WriteHeader(http.StatusOK)
		return
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	body, err := io.ReadAll(io.LimitReader(r.Body, maxBody+1))
	if err != nil || len(body) > maxBody {
		writeJSON(w, http.StatusRequestEntityTooLarge, rpcResponse{JSONRPC: "2.0", Error: &rpcError{-32600, "request too large"}})
		return
	}
	body = bytes.TrimSpace(body)
	if len(body) > 0 && body[0] == '[' {
		var batch []rpcRequest
		if json.Unmarshal(body, &batch) != nil {
			writeJSON(w, http.StatusBadRequest, rpcResponse{JSONRPC: "2.0", Error: &rpcError{-32700, "parse error"}})
			return
		}
		var out []rpcResponse
		for _, req := range batch {
			if resp, ok := s.dispatch(req); ok {
				out = append(out, resp)
			}
		}
		if len(out) == 0 {
			w.WriteHeader(http.StatusAccepted)
			return
		}
		writeJSON(w, http.StatusOK, out)
		return
	}
	var req rpcRequest
	if json.Unmarshal(body, &req) != nil {
		writeJSON(w, http.StatusBadRequest, rpcResponse{JSONRPC: "2.0", Error: &rpcError{-32700, "parse error"}})
		return
	}
	resp, ok := s.dispatch(req)
	if !ok {
		w.WriteHeader(http.StatusAccepted)
		return
	}
	if req.Method == "initialize" {
		if token, err := newToken(); err == nil {
			w.Header().Set("Mcp-Session-Id", token)
		}
	}
	writeJSON(w, http.StatusOK, resp)
}

// dispatch answers one JSON-RPC message; notifications get no answer.
func (s *server) dispatch(req rpcRequest) (rpcResponse, bool) {
	if len(req.ID) == 0 || string(req.ID) == "null" {
		return rpcResponse{}, false
	}
	var id any
	_ = json.Unmarshal(req.ID, &id)
	resp := rpcResponse{JSONRPC: "2.0", ID: id}
	switch req.Method {
	case "initialize":
		version := protocolVersion
		var p struct {
			ProtocolVersion string `json:"protocolVersion"`
		}
		if json.Unmarshal(req.Params, &p) == nil && p.ProtocolVersion != "" {
			version = p.ProtocolVersion
		}
		resp.Result = map[string]any{
			"protocolVersion": version,
			"capabilities":    map[string]any{"tools": map[string]any{"listChanged": false}},
			"serverInfo":      map[string]any{"name": "deeix-office", "version": serverVersion},
			"instructions":    "Создание файлов Word и Excel: make_docx и make_xlsx возвращают ссылку на скачивание.",
		}
	case "ping":
		resp.Result = map[string]any{}
	case "tools/list":
		resp.Result = map[string]any{"tools": toolList()}
	case "tools/call":
		resp.Result = s.callTool(req.Params)
	default:
		resp.Error = &rpcError{-32601, "method not found: " + req.Method}
	}
	return resp, true
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

// handleFile serves /files/<token>/<any name>; the stored name is used for the download.
func (s *server) handleFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	rest := strings.TrimPrefix(r.URL.Path, "/files/")
	token, _, _ := strings.Cut(rest, "/")
	f, meta, err := s.store.open(token)
	if err != nil {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusNotFound)
		_, _ = io.WriteString(w, "Файл не найден или срок хранения (7 дней) истёк. Попросите модель создать его снова.\n")
		return
	}
	defer f.Close()
	contentType := "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	fallback := "document.docx"
	if strings.HasSuffix(strings.ToLower(meta.Name), ".xlsx") {
		contentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
		fallback = "table.xlsx"
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"; filename*=UTF-8''%s`, fallback, url.PathEscape(meta.Name)))
	w.Header().Set("Cache-Control", "private, no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	http.ServeContent(w, r, "", meta.Created, f)
}

func main() {
	cfg := loadConfig()
	if len(os.Args) > 1 && os.Args[1] == "healthcheck" {
		addr := cfg.Addr
		if strings.HasPrefix(addr, ":") {
			addr = "127.0.0.1" + addr
		}
		resp, err := http.Get("http://" + addr + "/healthz")
		if err != nil || resp.StatusCode != http.StatusOK {
			os.Exit(1)
		}
		os.Exit(0)
	}
	if err := os.MkdirAll(cfg.Dir, 0o700); err != nil {
		log.Fatalf("data dir: %v", err)
	}
	st := &store{cfg: cfg}
	st.cleanup()
	srv := &server{cfg: cfg, store: st}

	go func() {
		for range time.Tick(time.Hour) {
			st.cleanup()
		}
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("/mcp", srv.handleMCP)
	mux.HandleFunc("/files/", srv.handleFile)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "version": serverVersion})
	})
	httpServer := &http.Server{
		Addr:              cfg.Addr,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       60 * time.Second,
		WriteTimeout:      120 * time.Second,
	}
	log.Printf("deeix-office %s on %s, files in %s for %s", serverVersion, cfg.Addr, cfg.Dir, cfg.TTL)
	log.Fatal(httpServer.ListenAndServe())
}
