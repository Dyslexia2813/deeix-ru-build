package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var testSecret = []byte("test-secret-0123456789")

func TestMaskContract(t *testing.T) {
	m := NewMasker(testSecret, nil)
	in := "Индивидуальный предприниматель Иванов Иван Иванович (ИНН 500100732259), тел. +7 (916) 123-45-67, " +
		"ivanov@mail.ru, ООО «Ромашка» (ИНН 7707083893, ОГРН 1027700132195) в лице директора Петровой Марии Сергеевны. " +
		"Подписи: /Иванов И.И./ /М.С. Петрова/. ИВАНОВУ ИВАНУ ИВАНОВИЧУ. Сумма 1234567890 рублей, договор № 15 от 01.02.2024."
	out := m.Mask(in)
	for _, secret := range []string{"Иванов", "ИВАНОВ", "500100732259", "916", "ivanov@", "7707083893", "1027700132195", "Петров", "Марии"} {
		if strings.Contains(out, secret) {
			t.Errorf("%q left in %s", secret, out)
		}
	}
	for _, keep := range []string{"1234567890 рублей", "01.02.2024", "ООО «Ромашка»", "№ 15"} {
		if !strings.Contains(out, keep) {
			t.Errorf("%q should stay: %s", keep, out)
		}
	}
	full := m.label(typeFIO, surnameStem("Иванов")+"|и|и")
	if strings.Count(out, full) != 3 {
		t.Errorf("full name, initials and caps should share %s: %s", full, out)
	}
	if back := m.Unmask(out, false); !strings.Contains(back, "Иванов Иван Иванович") || !strings.Contains(back, "500100732259") {
		t.Errorf("unmask: %s", back)
	}
}

func TestLabelsStableAcrossRequests(t *testing.T) {
	a := NewMasker(testSecret, nil).Mask("Петров Пётр Петрович")
	b := NewMasker(testSecret, nil).Mask("с Петровым Петром Петровичем")
	if !strings.Contains(b, a) {
		t.Fatalf("labels differ: %q vs %q", a, b)
	}
}

func TestNoFalsePositives(t *testing.T) {
	m := NewMasker(testSecret, nil)
	in := "Статья 432 Гражданского кодекса Российской Федерации. Москва, 2024 год. Сумма 100 000 рублей. Номер 1234567890. Сергей пришёл."
	if out := m.Mask(in); out != in {
		t.Fatalf("changed: %s", out)
	}
}

func TestDisabledTypes(t *testing.T) {
	m := NewMasker(testSecret, map[string]bool{typeFIO: true})
	out := m.Mask("Иванов Иван Иванович, ИНН 500100732259")
	if !strings.Contains(out, "500100732259") || strings.Contains(out, "Иванов") {
		t.Fatalf("got %s", out)
	}
}

func TestJSONEscapeOnUnmask(t *testing.T) {
	m := NewMasker(testSecret, nil)
	lbl := m.Mask("ivanov@mail.ru")
	m.values[lbl] = `a"b\c`
	if got := m.Unmask(`{"q":"`+lbl+`"}`, true); got != `{"q":"a\"b\\c"}` {
		t.Fatalf("got %s", got)
	}
}

func TestSplitHold(t *testing.T) {
	cases := map[string][2]string{
		"привет [FIO_ab": {"привет ", "[FIO_ab"},
		"текст [":        {"текст ", "["},
		"ссылка [1]":     {"ссылка [1]", ""},
		"[Текст":         {"[Текст", ""},
		"[FIO_abcde]":    {"[FIO_abcde]", ""},
	}
	for in, want := range cases {
		out, hold := splitHold(in)
		if out != want[0] || hold != want[1] {
			t.Errorf("%q: got %q/%q", in, out, hold)
		}
	}
}

func sse(events ...string) string {
	var b strings.Builder
	for _, e := range events {
		b.WriteString(e)
		b.WriteString("\n\n")
	}
	return b.String()
}

func collectOpenAI(t *testing.T, stream string) (string, string) {
	t.Helper()
	var content, args strings.Builder
	for _, line := range strings.Split(stream, "\n") {
		if !strings.HasPrefix(line, "data: ") || strings.Contains(line, "[DONE]") {
			continue
		}
		var chunk struct {
			Choices []struct {
				Delta struct {
					Content   string `json:"content"`
					ToolCalls []struct {
						Function struct {
							Arguments string `json:"arguments"`
						} `json:"function"`
					} `json:"tool_calls"`
				} `json:"delta"`
			} `json:"choices"`
		}
		if err := json.Unmarshal([]byte(line[6:]), &chunk); err != nil {
			t.Fatalf("bad chunk %s: %v", line, err)
		}
		for _, c := range chunk.Choices {
			content.WriteString(c.Delta.Content)
			for _, tc := range c.Delta.ToolCalls {
				args.WriteString(tc.Function.Arguments)
			}
		}
	}
	return content.String(), args.String()
}

func TestOpenAIStreamSplitLabels(t *testing.T) {
	m := NewMasker(testSecret, nil)
	lbl := m.Mask("Иванов Иван Иванович")
	half := len(lbl) / 2
	stream := sse(
		`data: {"id":"1","choices":[{"index":0,"delta":{"content":"Это `+lbl[:half]+`"}}]}`,
		`data: {"id":"1","choices":[{"index":0,"delta":{"content":"`+lbl[half:]+`, всё. ["}}]}`,
		`data: {"id":"1","choices":[{"index":0,"delta":{"tool_calls":[{"index":0,"id":"c1","function":{"name":"egrul","arguments":"{\"q\":\"`+lbl[:3]+`"}}]}}]}`,
		`data: {"id":"1","choices":[{"index":0,"delta":{"tool_calls":[{"index":0,"function":{"arguments":"`+lbl[3:]+`\"}"}}]}}]}`,
		`data: {"id":"1","choices":[{"index":0,"delta":{},"finish_reason":"tool_calls"}]}`,
		`data: [DONE]`,
	)
	var out bytes.Buffer
	if err := newStreamUnmasker(m, protoOpenAIChat, &out, nil).Run(strings.NewReader(stream)); err != nil {
		t.Fatal(err)
	}
	content, args := collectOpenAI(t, out.String())
	if content != "Это Иванов Иван Иванович, всё. [" {
		t.Errorf("content %q", content)
	}
	if args != `{"q":"Иванов Иван Иванович"}` {
		t.Errorf("args %q", args)
	}
	if !strings.HasSuffix(strings.TrimSpace(out.String()), "data: [DONE]") {
		t.Errorf("DONE must stay last: %s", out.String())
	}
}

func TestAnthropicStream(t *testing.T) {
	m := NewMasker(testSecret, nil)
	lbl := m.Mask("ivanov@mail.ru")
	stream := sse(
		"event: message_start\ndata: {\"type\":\"message_start\",\"message\":{\"id\":\"m\"}}",
		"event: content_block_start\ndata: {\"type\":\"content_block_start\",\"index\":0,\"content_block\":{\"type\":\"text\",\"text\":\"\"}}",
		"event: content_block_delta\ndata: {\"type\":\"content_block_delta\",\"index\":0,\"delta\":{\"type\":\"text_delta\",\"text\":\"Почта "+lbl[:4]+"\"}}",
		"event: content_block_delta\ndata: {\"type\":\"content_block_delta\",\"index\":0,\"delta\":{\"type\":\"text_delta\",\"text\":\""+lbl[4:]+" и ["+"\"}}",
		"event: content_block_stop\ndata: {\"type\":\"content_block_stop\",\"index\":0}",
		"event: content_block_start\ndata: {\"type\":\"content_block_start\",\"index\":1,\"content_block\":{\"type\":\"tool_use\",\"id\":\"t\",\"name\":\"x\",\"input\":{}}}",
		"event: content_block_delta\ndata: {\"type\":\"content_block_delta\",\"index\":1,\"delta\":{\"type\":\"input_json_delta\",\"partial_json\":\"{\\\"to\\\":\\\""+lbl+"\\\"}\"}}",
		"event: content_block_stop\ndata: {\"type\":\"content_block_stop\",\"index\":1}",
		"event: message_stop\ndata: {\"type\":\"message_stop\"}",
	)
	var out bytes.Buffer
	if err := newStreamUnmasker(m, protoAnthropic, &out, nil).Run(strings.NewReader(stream)); err != nil {
		t.Fatal(err)
	}
	var text, js strings.Builder
	events := strings.Split(out.String(), "\n\n")
	stopSeen := false
	for _, e := range events {
		if strings.Contains(e, "content_block_stop") && strings.Contains(e, `"index":0`) {
			stopSeen = true
		}
		i := strings.Index(e, "data: ")
		if i < 0 {
			continue
		}
		var ev struct {
			Type  string `json:"type"`
			Delta struct {
				Text        string `json:"text"`
				PartialJSON string `json:"partial_json"`
			} `json:"delta"`
		}
		if err := json.Unmarshal([]byte(e[i+6:]), &ev); err != nil {
			t.Fatalf("bad event %q: %v", e, err)
		}
		if ev.Type == "content_block_delta" && !strings.HasPrefix(e, "event: content_block_delta") {
			t.Fatalf("event line missing: %q", e)
		}
		if !stopSeen {
			text.WriteString(ev.Delta.Text)
		}
		js.WriteString(ev.Delta.PartialJSON)
	}
	if text.String() != "Почта ivanov@mail.ru и [" {
		t.Errorf("text %q", text.String())
	}
	if js.String() != `{"to":"ivanov@mail.ru"}` {
		t.Errorf("json %q", js.String())
	}
}

func post(t *testing.T, url, body string) (*http.Response, string) {
	t.Helper()
	req, _ := http.NewRequest(http.MethodPost, url, strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer sk-1")
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	return resp, string(data)
}

func TestProxyOpenAIRoundTrip(t *testing.T) {
	var sent []byte
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sent, _ = io.ReadAll(r.Body)
		if r.URL.Path != "/v1/chat/completions" || r.Header.Get("Authorization") != "Bearer sk-1" {
			t.Errorf("path %s / auth %q", r.URL.Path, r.Header.Get("Authorization"))
		}
		label := labelPattern.FindString(string(sent))
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"choices":[{"index":0,"message":{"role":"assistant","content":"Ответ для `+label+`","tool_calls":[{"id":"1","type":"function","function":{"name":"f","arguments":"{\"who\":\"`+label+`\"}"}}]}}]}`)
	}))
	defer up.Close()
	routes, _ := parseRoutes("p=" + up.URL)
	proxy := httptest.NewServer(newServer(config{flag: "deeix_pd_mask", secret: testSecret, hint: true, maxBody: 1 << 20, routes: routes}))
	defer proxy.Close()
	body := `{"model":"x","deeix_pd_mask":true,"messages":[{"role":"system","content":"Ты юрист."},{"role":"user","content":"Проверь Иванова Ивана Ивановича"},` +
		`{"role":"assistant","content":"","tool_calls":[{"id":"1","type":"function","function":{"name":"egrul","arguments":"{\"inn\":\"500100732259\"}"}}]},` +
		`{"role":"tool","tool_call_id":"1","content":"Найден ИП Иванов Иван Иванович, ИНН 500100732259"}]}`
	resp, out := post(t, proxy.URL+"/p/v1/chat/completions", body)
	for _, secret := range []string{"Иванов", "500100732259", "deeix_pd_mask"} {
		if strings.Contains(string(sent), secret) {
			t.Errorf("%q reached upstream: %s", secret, sent)
		}
	}
	if !strings.Contains(string(sent), "Служебная пометка") || !strings.Contains(string(sent), "Ты юрист.") {
		t.Errorf("hint missing: %s", sent)
	}
	if resp.Header.Get("X-PD-Masked") == "" {
		t.Errorf("no X-PD-Masked header")
	}
	if !strings.Contains(out, "Ответ для Иван") {
		t.Errorf("answer not unmasked: %s", out)
	}
	if !strings.Contains(out, `\"who\":\"Иван`) {
		t.Errorf("tool args not unmasked: %s", out)
	}
}

func TestProxyPassThroughWithoutFlag(t *testing.T) {
	var got []byte
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got, _ = io.ReadAll(r.Body)
		_, _ = io.WriteString(w, `{"ok":true}`)
	}))
	defer up.Close()
	routes, _ := parseRoutes("p=" + up.URL)
	proxy := httptest.NewServer(newServer(config{flag: "deeix_pd_mask", secret: testSecret, maxBody: 1 << 20, routes: routes}))
	defer proxy.Close()
	body := `{"messages":[{"role":"user","content":"Иванов Иван Иванович"}],  "x": 1}`
	_, out := post(t, proxy.URL+"/p/v1/chat/completions", body)
	if string(got) != body || out != `{"ok":true}` {
		t.Fatalf("not byte-for-byte: %s / %s", got, out)
	}
	body = `{"deeix_pd_mask":false,"messages":[{"role":"user","content":"Иванов Иван Иванович"}]}`
	_, _ = post(t, proxy.URL+"/p/v1/chat/completions", body)
	if strings.Contains(string(got), "deeix_pd_mask") || !strings.Contains(string(got), "Иванов") {
		t.Fatalf("flag off: %s", got)
	}
}

func TestProxyAnthropicStreamAndAlways(t *testing.T) {
	var got []byte
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got, _ = io.ReadAll(r.Body)
		label := labelPattern.FindString(string(got))
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(w, sse(
			"event: content_block_delta\ndata: {\"type\":\"content_block_delta\",\"index\":0,\"delta\":{\"type\":\"text_delta\",\"text\":\""+label[:5]+"\"}}",
			"event: content_block_delta\ndata: {\"type\":\"content_block_delta\",\"index\":0,\"delta\":{\"type\":\"text_delta\",\"text\":\""+label[5:]+"!\"}}",
			"event: content_block_stop\ndata: {\"type\":\"content_block_stop\",\"index\":0}",
		))
	}))
	defer up.Close()
	routes, _ := parseRoutes("p=" + up.URL + " always")
	proxy := httptest.NewServer(newServer(config{flag: "deeix_pd_mask", secret: testSecret, maxBody: 1 << 20, routes: routes}))
	defer proxy.Close()
	_, out := post(t, proxy.URL+"/p/anthropic/v1/messages", `{"system":[{"type":"text","text":"Клиент: Петров П.П."}],"messages":[{"role":"user","content":[{"type":"text","text":"Привет"}]}],"stream":true}`)
	if strings.Contains(string(got), "Петров") {
		t.Fatalf("always route leaked: %s", got)
	}
	if !strings.Contains(out, "Петров П.П.") || !strings.Contains(out, `"text":"!"`) && !strings.Contains(out, "П.П.!") {
		t.Fatalf("stream not restored: %s", out)
	}
}

func TestUnsupportedProtocolRefused(t *testing.T) {
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("request must not reach upstream")
	}))
	defer up.Close()
	routes, _ := parseRoutes("p=" + up.URL)
	proxy := httptest.NewServer(newServer(config{flag: "deeix_pd_mask", secret: testSecret, maxBody: 1 << 20, routes: routes}))
	defer proxy.Close()
	resp, _ := post(t, proxy.URL+"/p/v1/responses", `{"deeix_pd_mask":true,"input":"Иванов Иван Иванович"}`)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status %d", resp.StatusCode)
	}
}
