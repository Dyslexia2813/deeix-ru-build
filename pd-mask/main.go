// deeix-pd-mask: a small proxy between DEEIX Chat and LLM providers that replaces personal data
// with labels before a request leaves the server and puts the real values back in the answer.
// Masking is switched per request by a flag in the JSON body (deeix_pd_mask: true), which DEEIX
// adds through a model control ("Маскировка ПД" in the chat parameters panel).
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
	"net"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"
)

type route struct {
	name     string
	upstream *url.URL
	always   bool // mask every request on this route, flag or not (task models: titles, summaries)
}

type config struct {
	addr        string
	flag        string
	secret      []byte
	hint        bool
	rejectFiles bool
	maxBody     int64
	enabled     map[string]bool
	routes      map[string]route
}

var allTypes = []string{typeFIO, typeTel, typeEmail, typeINN, typeSNILS, typeOGRN, typePasport, typeCard, typeAccount, typeBirth, typeAddress}

func env(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}

func envBool(key string, def bool) bool {
	switch strings.ToLower(env(key, "")) {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	}
	return def
}

// parseRoutes reads PD_ROUTES: "name=https://host[ always]" entries separated by ";" or newlines.
func parseRoutes(raw string) (map[string]route, error) {
	routes := map[string]route{}
	for _, entry := range strings.FieldsFunc(raw, func(r rune) bool { return r == ';' || r == '\n' }) {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		name, target, ok := strings.Cut(entry, "=")
		name = strings.TrimSpace(name)
		if !ok || name == "" || strings.ContainsAny(name, "/ ") {
			return nil, fmt.Errorf("bad route %q (want name=https://host)", entry)
		}
		fields := strings.Fields(target)
		if len(fields) == 0 {
			return nil, fmt.Errorf("route %q has no upstream", name)
		}
		u, err := url.Parse(fields[0])
		if err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" {
			return nil, fmt.Errorf("route %q: bad upstream %q", name, fields[0])
		}
		u.Path = strings.TrimRight(u.Path, "/")
		r := route{name: name, upstream: u}
		for _, opt := range fields[1:] {
			if opt == "always" {
				r.always = true
			} else {
				return nil, fmt.Errorf("route %q: unknown option %q", name, opt)
			}
		}
		routes[name] = r
	}
	if len(routes) == 0 {
		return nil, errors.New("PD_ROUTES is empty")
	}
	return routes, nil
}

func loadConfig() (config, error) {
	cfg := config{
		addr:        env("PD_ADDR", ":8090"),
		flag:        env("PD_FLAG", "deeix_pd_mask"),
		hint:        envBool("PD_HINT", true),
		rejectFiles: envBool("PD_REJECT_FILES", false),
		maxBody:     64 << 20,
	}
	secret := env("PD_SECRET", "")
	if len(secret) < 16 {
		return cfg, errors.New("PD_SECRET must be at least 16 characters (openssl rand -hex 16)")
	}
	cfg.secret = []byte(secret)
	cfg.enabled = map[string]bool{}
	types := env("PD_TYPES", strings.Join(allTypes, ","))
	for _, t := range strings.Split(types, ",") {
		t = strings.ToUpper(strings.TrimSpace(t))
		if t == "" {
			continue
		}
		known := false
		for _, k := range allTypes {
			known = known || k == t
		}
		if !known {
			return cfg, fmt.Errorf("PD_TYPES: unknown type %q (known: %s)", t, strings.Join(allTypes, ","))
		}
		cfg.enabled[t] = true
	}
	routes, err := parseRoutes(env("PD_ROUTES", ""))
	if err != nil {
		return cfg, err
	}
	cfg.routes = routes
	return cfg, nil
}

var hopHeaders = map[string]bool{
	"Connection": true, "Keep-Alive": true, "Proxy-Authenticate": true, "Proxy-Authorization": true,
	"Te": true, "Trailer": true, "Transfer-Encoding": true, "Upgrade": true, "Host": true,
	"Content-Length": true, "Accept-Encoding": true,
}

type server struct {
	cfg    config
	client *http.Client
}

func newServer(cfg config) *server {
	transport := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           (&net.Dialer{Timeout: 15 * time.Second, KeepAlive: 30 * time.Second}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          32,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   15 * time.Second,
		ExpectContinueTimeout: time.Second,
	}
	return &server{cfg: cfg, client: &http.Client{Transport: transport, CheckRedirect: func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}}}
}

func writeError(w http.ResponseWriter, status int, proto protocol, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	var body any
	if proto == protoAnthropic {
		body = map[string]any{"type": "error", "error": map[string]any{"type": "invalid_request_error", "message": message}}
	} else {
		body = map[string]any{"error": map[string]any{"message": message, "type": "pd_mask_error", "code": "pd_mask"}}
	}
	_ = json.NewEncoder(w).Encode(body)
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/healthz" {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status":"ok"}`)
		return
	}
	trimmed := strings.TrimPrefix(r.URL.Path, "/")
	name, rest, _ := strings.Cut(trimmed, "/")
	rt, ok := s.cfg.routes[name]
	if !ok {
		http.Error(w, "unknown route "+name, http.StatusNotFound)
		return
	}
	proto := detectProtocol("/" + rest)

	var body []byte
	if r.Body != nil {
		var err error
		body, err = io.ReadAll(io.LimitReader(r.Body, s.cfg.maxBody+1))
		if err != nil {
			writeError(w, http.StatusBadGateway, proto, "pd-mask: cannot read request: "+err.Error())
			return
		}
		if int64(len(body)) > s.cfg.maxBody {
			writeError(w, http.StatusRequestEntityTooLarge, proto, "pd-mask: request is too large")
			return
		}
	}

	var masker *Masker
	if r.Method == http.MethodPost && len(body) > 0 && bytes.Contains(body, []byte(s.cfg.flag)) || rt.always && r.Method == http.MethodPost && len(body) > 0 {
		parsed, err := decodeJSON(body)
		obj, isObj := parsed.(map[string]any)
		if err == nil && isObj {
			flag, hasFlag := obj[s.cfg.flag]
			delete(obj, s.cfg.flag)
			want := rt.always || flag == true || flag == "true" || flag == "on"
			if want {
				if proto == protoUnknown {
					writeError(w, http.StatusBadRequest, proto, "Маскировка ПД работает только с протоколами OpenAI Chat Completions и Anthropic Messages; этот запрос не отправлен.")
					return
				}
				started := time.Now()
				masker = NewMasker(s.cfg.secret, s.cfg.enabled)
				stats := &requestStats{}
				if proto == protoOpenAIChat {
					masker.maskOpenAIChat(obj, s.cfg.hint, stats)
				} else {
					masker.maskAnthropic(obj, s.cfg.hint, stats)
				}
				if stats.binaryParts > 0 && s.cfg.rejectFiles {
					writeError(w, http.StatusBadRequest, proto, fmt.Sprintf("Маскировка ПД: в запросе %d изображений или файлов, их содержимое замаскировать нельзя, поэтому запрос не отправлен (PD_REJECT_FILES=1).", stats.binaryParts))
					return
				}
				log.Printf("route=%s masked=%d %s binary=%d bytes=%d took=%s", rt.name, masker.Total(), countsString(masker.counts), stats.binaryParts, len(body), time.Since(started).Round(time.Millisecond))
			}
			if want || hasFlag {
				encoded, err := encodeJSON(obj)
				if err != nil {
					writeError(w, http.StatusInternalServerError, proto, "pd-mask: encode: "+err.Error())
					return
				}
				body = encoded
			}
		}
	}

	target := *rt.upstream
	target.Path = rt.upstream.Path + "/" + rest
	target.RawPath = ""
	target.RawQuery = r.URL.RawQuery
	outReq, err := http.NewRequestWithContext(r.Context(), r.Method, target.String(), bytes.NewReader(body))
	if err != nil {
		writeError(w, http.StatusBadGateway, proto, "pd-mask: "+err.Error())
		return
	}
	for key, values := range r.Header {
		if hopHeaders[http.CanonicalHeaderKey(key)] {
			continue
		}
		for _, v := range values {
			outReq.Header.Add(key, v)
		}
	}
	outReq.ContentLength = int64(len(body))

	resp, err := s.client.Do(outReq)
	if err != nil {
		writeError(w, http.StatusBadGateway, proto, "pd-mask: upstream: "+err.Error())
		return
	}
	defer resp.Body.Close()

	for key, values := range resp.Header {
		if hopHeaders[http.CanonicalHeaderKey(key)] || key == "Content-Encoding" {
			continue
		}
		for _, v := range values {
			w.Header().Add(key, v)
		}
	}
	if masker != nil {
		w.Header().Set("X-PD-Masked", fmt.Sprint(masker.Total()))
	}
	flusher, _ := w.(http.Flusher)
	flush := func() {
		if flusher != nil {
			flusher.Flush()
		}
	}
	isStream := strings.Contains(resp.Header.Get("Content-Type"), "text/event-stream")

	if masker == nil || len(masker.values) == 0 || resp.StatusCode >= 300 {
		w.WriteHeader(resp.StatusCode)
		copyStream(w, resp.Body, flush)
		return
	}
	if isStream {
		w.WriteHeader(resp.StatusCode)
		if err := newStreamUnmasker(masker, proto, w, flush).Run(resp.Body); err != nil {
			log.Printf("route=%s stream: %v", rt.name, err)
		}
		return
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		writeError(w, http.StatusBadGateway, proto, "pd-mask: upstream read: "+err.Error())
		return
	}
	if parsed, err := decodeJSON(data); err == nil {
		if obj, ok := parsed.(map[string]any); ok {
			if proto == protoOpenAIChat {
				masker.unmaskOpenAIResponse(obj)
			} else {
				masker.unmaskAnthropicResponse(obj)
			}
			if encoded, err := encodeJSON(obj); err == nil {
				data = encoded
			}
		}
	}
	w.WriteHeader(resp.StatusCode)
	_, _ = w.Write(data)
}

func copyStream(w io.Writer, r io.Reader, flush func()) {
	buf := make([]byte, 32*1024)
	for {
		n, err := r.Read(buf)
		if n > 0 {
			if _, werr := w.Write(buf[:n]); werr != nil {
				return
			}
			flush()
		}
		if err != nil {
			return
		}
	}
}

func countsString(counts map[string]int) string {
	keys := make([]string, 0, len(counts))
	for k := range counts {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s:%d", k, counts[k]))
	}
	return "[" + strings.Join(parts, " ") + "]"
}

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "healthcheck":
			resp, err := http.Get("http://127.0.0.1" + env("PD_ADDR", ":8090") + "/healthz")
			if err != nil || resp.StatusCode != http.StatusOK {
				os.Exit(1)
			}
			os.Exit(0)
		case "secret":
			b := make([]byte, 16)
			_, _ = rand.Read(b)
			fmt.Println(hex.EncodeToString(b))
			os.Exit(0)
		case "mask":
			// Debug: mask stdin with PD_SECRET and print the result (no network).
			data, _ := io.ReadAll(os.Stdin)
			secret := env("PD_SECRET", "debug-secret-0123456789")
			m := NewMasker([]byte(secret), nil)
			fmt.Println(m.Mask(string(data)))
			fmt.Fprintln(os.Stderr, "replaced:", m.Total(), countsString(m.counts))
			os.Exit(0)
		}
	}
	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	names := make([]string, 0, len(cfg.routes))
	for name, r := range cfg.routes {
		mode := "by flag"
		if r.always {
			mode = "always"
		}
		names = append(names, fmt.Sprintf("/%s -> %s (%s)", name, r.upstream, mode))
	}
	sort.Strings(names)
	log.Printf("deeix-pd-mask on %s, flag %q, routes: %s", cfg.addr, cfg.flag, strings.Join(names, ", "))
	srv := &http.Server{Addr: cfg.addr, Handler: newServer(cfg), ReadHeaderTimeout: 30 * time.Second}
	log.Fatal(srv.ListenAndServe())
}
