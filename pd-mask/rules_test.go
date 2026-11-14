package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const demo = "Претензия от ИП Иванов Иван Иванович (ИНН 500100732259) к ООО «Ромашка» (ИНН 7707083893). " +
	"Подпись: /Иванов И.И./. Сошлись на позицию Менделеева и Петрову. ИНН 123456789 указан с ошибкой. Российской Федерации."

func TestPreviewSuspectsAndRules(t *testing.T) {
	resp := NewMasker(testSecret, nil).preview(previewRequest{Texts: []previewText{{ID: "message", Text: demo}}})
	values := map[string]bool{}
	for _, r := range resp.Replacements {
		values[r.Value] = true
		if r.Kept {
			t.Errorf("nothing should be kept: %+v", r)
		}
	}
	if !values["Иванов Иван Иванович"] || !values["7707083893"] {
		t.Fatalf("replacements: %+v", resp.Replacements)
	}
	sus := map[string]string{}
	for _, s := range resp.Suspects {
		sus[s.Value] = s.Reason
	}
	if sus["Менделеева"] != "surname" || sus["Петрову"] != "surname" || sus["123456789"] != "inn" {
		t.Fatalf("suspects: %+v", resp.Suspects)
	}
	for _, bad := range []string{"Российской", "Ромашка", "Претензия", "Подпись"} {
		if _, ok := sus[bad]; ok {
			t.Errorf("false suspect %q", bad)
		}
	}

	// The user keeps the company's INN and Ivanov, and hides Mendeleev.
	resp = NewMasker(testSecret, nil).preview(previewRequest{
		Texts: []previewText{{ID: "message", Text: demo + " Менделееву тоже."}},
		Keep:  []string{"7707083893", "Иванов Иван Иванович"},
		Hide:  []string{"Менделеева"},
	})
	masked := resp.Texts[0].Masked
	if !strings.Contains(masked, "7707083893") || !strings.Contains(masked, "Иванов Иван Иванович") || !strings.Contains(masked, "/Иванов И.И./") {
		t.Errorf("kept values must stay: %s", masked)
	}
	if strings.Contains(masked, "Менделеев") {
		t.Errorf("hidden word left: %s", masked)
	}
	kept := 0
	for _, r := range resp.Replacements {
		if r.Kept {
			kept++
		}
	}
	if kept != 2 {
		t.Errorf("kept rows = %d: %+v", kept, resp.Replacements)
	}
	for _, s := range resp.Suspects {
		if strings.HasPrefix(s.Value, "Менделеев") || s.Value == "Иванов" {
			t.Errorf("handled word still suspect: %+v", s)
		}
	}
}

func TestRulesInRequest(t *testing.T) {
	var got []byte
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got, _ = io.ReadAll(r.Body)
		_, _ = io.WriteString(w, `{"choices":[]}`)
	}))
	defer up.Close()
	routes, _ := parseRoutes("p=" + up.URL)
	proxy := httptest.NewServer(newServer(config{flag: "deeix_pd_mask", secret: testSecret, maxBody: 1 << 20, routes: routes}))
	defer proxy.Close()
	body := `{"deeix_pd_mask":true,"deeix_pd_mask_rules":{"keep":["7707083893"],"hide":["Менделеев"]},"messages":[{"role":"user","content":"` + demo + `"}]}`
	post(t, proxy.URL+"/p/v1/chat/completions", body)
	sent := string(got)
	if strings.Contains(sent, "deeix_pd_mask") || !strings.Contains(sent, "7707083893") || strings.Contains(sent, "Менделеев") || strings.Contains(sent, "500100732259") {
		t.Fatalf("rules not applied: %s", sent)
	}
	// Rules without the flag are only removed.
	post(t, proxy.URL+"/p/v1/chat/completions", `{"deeix_pd_mask_rules":{"keep":["x"]},"messages":[{"role":"user","content":"Менделеев"}]}`)
	if strings.Contains(string(got), "deeix_pd_mask") || !strings.Contains(string(got), "Менделеев") {
		t.Fatalf("without flag: %s", got)
	}
}

func TestPreviewEndpoint(t *testing.T) {
	routes, _ := parseRoutes("p=https://example.com")
	proxy := httptest.NewServer(newServer(config{flag: "deeix_pd_mask", secret: testSecret, maxBody: 1 << 20, routes: routes}))
	defer proxy.Close()
	resp, out := post(t, proxy.URL+"/preview", `{"texts":[{"id":"message","text":"Иванов Иван Иванович"},{"id":"file:1","text":"тел. 8 916 123 45 67"}]}`)
	var parsed previewResponse
	if resp.StatusCode != 200 || json.Unmarshal([]byte(out), &parsed) != nil || len(parsed.Texts) != 2 || len(parsed.Replacements) != 2 {
		t.Fatalf("%d %s", resp.StatusCode, out)
	}
	if _, err := parseRoutes("preview=https://x"); err == nil {
		t.Fatal("route name preview must be refused")
	}
}
