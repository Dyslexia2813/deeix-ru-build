package main

import (
	"regexp"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Rules are the user's choices from the preview window, sent with every request of the chat in
// "deeix_pd_mask_rules": values to leave as they are (Keep) and words to hide in addition (Hide).
type Rules struct {
	Keep []string `json:"keep"`
	Hide []string `json:"hide"`

	keepKeys map[string]bool
	keepText map[string]bool
	hide     []hideRule
}

type hideRule struct {
	re   *regexp.Regexp
	kind string
	key  string
	word string // as the user typed it
}

const maxRuleItems = 200

func normText(s string) string {
	return strings.ToLower(strings.Join(strings.Fields(strings.ReplaceAll(s, "ё", "е")), " "))
}

// parseRules reads the rules object from a request body value (nil when there is nothing to apply).
func parseRules(raw any) *Rules {
	obj, ok := raw.(map[string]any)
	if !ok {
		return nil
	}
	list := func(v any) []string {
		items, _ := v.([]any)
		var out []string
		for _, item := range items {
			if s, ok := item.(string); ok && strings.TrimSpace(s) != "" && len(out) < maxRuleItems && len(s) <= 300 {
				out = append(out, strings.TrimSpace(s))
			}
		}
		return out
	}
	return newRules(list(obj["keep"]), list(obj["hide"]))
}

func newRules(keep, hide []string) *Rules {
	if len(keep) == 0 && len(hide) == 0 {
		return nil
	}
	r := &Rules{Keep: keep, Hide: hide, keepKeys: map[string]bool{}, keepText: map[string]bool{}}
	for _, value := range keep {
		r.keepText[normText(value)] = true
		// «Иванов Иван Иванович» kept also keeps «Иванов И.И.»: same detector key.
		for _, s := range findSpans(value) {
			r.keepKeys[s.kind+"|"+s.key] = true
		}
	}
	for _, word := range hide {
		if rule, ok := makeHideRule(word); ok {
			r.hide = append(r.hide, rule)
		}
	}
	return r
}

func isCyrillicWord(s string) bool {
	if utf8.RuneCountInString(s) < 3 {
		return false
	}
	for _, r := range s {
		if !unicode.Is(unicode.Cyrillic, r) && r != '-' {
			return false
		}
	}
	return true
}

// makeHideRule: a single Cyrillic word is hidden in any case form («Менделеев», «Менделеева»,
// «Менделееву»); anything else is hidden as typed, ignoring letter case.
func makeHideRule(word string) (hideRule, bool) {
	word = strings.TrimSpace(word)
	if word == "" {
		return hideRule{}, false
	}
	if isCyrillicWord(word) {
		stem := surnameStem(word)
		if utf8.RuneCountInString(stem) < 3 {
			stem = strings.ToLower(word)
		}
		pattern := `(?i)` + regexp.QuoteMeta(stem) + `[а-яё]{0,4}`
		if strings.Contains(stem, "е") {
			pattern = `(?i)` + strings.ReplaceAll(regexp.QuoteMeta(stem), "е", "[её]") + `[а-яё]{0,4}`
		}
		kind := typeOther
		if r, _ := utf8.DecodeRuneInString(word); unicode.IsUpper(r) {
			kind = typeFIO
		}
		return hideRule{re: regexp.MustCompile(pattern), kind: kind, key: "hide|" + stem, word: word}, true
	}
	return hideRule{re: regexp.MustCompile(`(?i)` + regexp.QuoteMeta(word)), kind: typeOther, key: "hide|" + normText(word), word: word}, true
}

// hideWord returns the user's word behind a span added by Hide ("" for detected data).
func (r *Rules) hideWord(s span) string {
	if r == nil || !strings.HasPrefix(s.key, "hide|") {
		return ""
	}
	for _, rule := range r.hide {
		if rule.key == s.key {
			return rule.word
		}
	}
	return ""
}

func (r *Rules) keeps(s span, original string) bool {
	if r == nil {
		return false
	}
	return r.keepKeys[s.kind+"|"+s.key] || r.keepText[normText(original)]
}

// spans finds the personal data in text, plus the words the user asked to hide.
func (m *Masker) spans(text string) []span {
	raw := findRawSpans(text)
	if m.rules != nil {
		for _, rule := range m.rules.hide {
			for _, loc := range rule.re.FindAllStringIndex(text, -1) {
				if boundaryOK(text, loc[0], loc[1], false) {
					raw = append(raw, span{start: loc[0], end: loc[1], kind: rule.kind, key: rule.key})
				}
			}
		}
	}
	return resolveOverlaps(raw)
}

// ---------------------------------------------------------------------------
// Preview
// ---------------------------------------------------------------------------

type previewText struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

type previewRequest struct {
	Texts []previewText `json:"texts"`
	Keep  []string      `json:"keep"`
	Hide  []string      `json:"hide"`
}

type previewReplacement struct {
	Label string `json:"label"`
	Value string `json:"value"`
	Type  string `json:"type"`
	Count int    `json:"count"`
	Kept  bool   `json:"kept"`
	// Manual is the word from Hide that produced this label ("" for detected data).
	Manual string `json:"manual,omitempty"`
}

type previewSuspect struct {
	Value  string `json:"value"`
	Reason string `json:"reason"` // "surname", "inn"
	Count  int    `json:"count"`
}

type previewResult struct {
	ID       string `json:"id"`
	Masked   string `json:"masked"`
	Replaced int    `json:"replaced"`
}

type previewResponse struct {
	Texts        []previewResult      `json:"texts"`
	Replacements []previewReplacement `json:"replacements"`
	Suspects     []previewSuspect     `json:"suspects"`
}

func (m *Masker) preview(req previewRequest) previewResponse {
	m.rules = newRules(req.Keep, req.Hide)
	resp := previewResponse{Texts: []previewResult{}, Replacements: []previewReplacement{}, Suspects: []previewSuspect{}}
	suspects := map[string]*previewSuspect{}
	var order []string
	for _, item := range req.Texts {
		before := m.Total()
		masked := m.Mask(item.Text)
		resp.Texts = append(resp.Texts, previewResult{ID: item.ID, Masked: masked, Replaced: m.Total() - before})
		for _, sus := range findSuspects(item.Text, m.spans(item.Text)) {
			key := strings.ToLower(sus.Value)
			if prev, ok := suspects[key]; ok {
				prev.Count++
				continue
			}
			copy := sus
			suspects[key] = &copy
			order = append(order, key)
		}
	}
	for lbl, value := range m.values {
		resp.Replacements = append(resp.Replacements, previewReplacement{Label: lbl, Value: value, Type: m.labelKinds[lbl], Count: m.labelCounts[lbl], Manual: m.labelManual[lbl]})
	}
	for lbl, value := range m.kept {
		if _, masked := m.values[lbl]; masked {
			continue
		}
		resp.Replacements = append(resp.Replacements, previewReplacement{Label: lbl, Value: value, Type: m.labelKinds[lbl], Count: m.keptCounts[lbl], Kept: true})
	}
	sort.SliceStable(resp.Replacements, func(i, j int) bool {
		a, b := resp.Replacements[i], resp.Replacements[j]
		if a.Type != b.Type {
			return typeOrder(a.Type) < typeOrder(b.Type)
		}
		return a.Value < b.Value
	})
	for _, key := range order {
		resp.Suspects = append(resp.Suspects, *suspects[key])
	}
	return resp
}

func typeOrder(kind string) int {
	for i, k := range allTypes {
		if k == kind {
			return i
		}
	}
	return len(allTypes)
}

var (
	reCapWord       = regexp.MustCompile(`[А-ЯЁ][а-яё]{2,}(?:-[А-ЯЁ][а-яё]+)?`)
	reSurnameSuffix = regexp.MustCompile(`(?:ов|ев|ёв|ин|ын|ский|цкий|ской|цкой|енко|ко|ук|юк|чук|ян|янц|дзе|швили|ых|их)(?:а|у|ым|ой|ом|е|ы|ого|ому|ая|ую|ий|им|ими|ыми)?$`)
	reINNContext    = regexp.MustCompile(`(?i)инн[^\d\n]{0,12}(\d{9,13})`)
)

// Capitalised words that end like surnames but are ordinary words.
var suspectStopPrefixes = []string{
	"российск", "русск", "московск", "петербургск", "советск", "гражданск", "арбитражн", "федеральн", "конституц",
	"верховн", "областн", "районн", "городск", "налогов", "трудов", "уголовн", "семейн", "жилищн", "земельн",
	"европейск", "английск", "немецк", "французск", "китайск", "американск", "сибирск", "уральск", "кавказск",
	"казанск", "новосибирск", "екатеринбургск", "нижегородск", "краснодарск", "самарск", "ростовск", "воронежск",
	"росси", "москв", "договор", "статья", "статьи", "закон", "кодекс", "приложени", "заказчик", "исполнител",
	"продав", "покупател", "арендатор", "арендодател", "истц", "ответчик", "сторон", "директор", "генеральн",
	"председател", "министерств", "правительств", "президент", "дума", "совет", "суд", "банк", "сбербанк", "втб",
}

func findSuspects(text string, handled []span) []previewSuspect {
	var out []previewSuspect
	covered := func(start, end int) bool {
		for _, s := range handled {
			if start < s.end && s.start < end {
				return true
			}
		}
		return false
	}
	for _, loc := range reCapWord.FindAllStringIndex(text, -1) {
		if !boundaryOK(text, loc[0], loc[1], false) || covered(loc[0], loc[1]) {
			continue
		}
		word := text[loc[0]:loc[1]]
		lower := strings.ToLower(word)
		if !reSurnameSuffix.MatchString(lower) {
			continue
		}
		stop := false
		for _, prefix := range suspectStopPrefixes {
			if strings.HasPrefix(lower, prefix) {
				stop = true
				break
			}
		}
		if stop || insideQuotes(text, loc[0]) {
			continue
		}
		out = append(out, previewSuspect{Value: word, Reason: "surname", Count: 1})
	}
	for _, m := range reINNContext.FindAllStringSubmatchIndex(text, -1) {
		if covered(m[2], m[3]) {
			continue
		}
		out = append(out, previewSuspect{Value: text[m[2]:m[3]], Reason: "inn", Count: 1})
	}
	return out
}

// insideQuotes: «ООО «Петров и партнёры»» style names of organisations are not people.
func insideQuotes(text string, pos int) bool {
	open := strings.LastIndex(text[:pos], "«")
	if open < 0 {
		return false
	}
	return !strings.Contains(text[open:pos], "»")
}
