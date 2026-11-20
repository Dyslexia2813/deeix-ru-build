package main

import (
	"strings"
)

// run is a piece of text with the same character formatting.
type run struct {
	Text      string
	Bold      bool
	Italic    bool
	Underline bool
}

// parseInline splits text with light markdown marks into runs:
// **bold**, *italic*, __underline__. A mark without its pair stays as typed.
func parseInline(text string) []run {
	type mark struct {
		token string
		apply func(*run, bool)
	}
	marks := []mark{
		{"**", func(r *run, on bool) { r.Bold = on }},
		{"__", func(r *run, on bool) { r.Underline = on }},
		{"*", func(r *run, on bool) { r.Italic = on }},
	}
	// Only marks that have a closing pair are treated as formatting.
	paired := map[string]bool{}
	for _, m := range marks {
		count := strings.Count(text, m.token)
		if m.token == "*" {
			count = strings.Count(strings.ReplaceAll(text, "**", ""), "*")
		}
		paired[m.token] = count >= 2
	}

	var runs []run
	state := run{}
	var buf strings.Builder
	flush := func() {
		if buf.Len() == 0 {
			return
		}
		r := state
		r.Text = buf.String()
		runs = append(runs, r)
		buf.Reset()
	}
	open := map[string]bool{}
	remaining := map[string]int{}
	for _, m := range marks {
		if m.token == "*" {
			remaining[m.token] = strings.Count(strings.ReplaceAll(text, "**", ""), "*")
		} else {
			remaining[m.token] = strings.Count(text, m.token)
		}
	}
	for i := 0; i < len(text); {
		matched := false
		for _, m := range marks {
			if !paired[m.token] || !strings.HasPrefix(text[i:], m.token) {
				continue
			}
			// An opening mark needs a closing one later in the text.
			if !open[m.token] && remaining[m.token] < 2 {
				continue
			}
			flush()
			open[m.token] = !open[m.token]
			m.apply(&state, open[m.token])
			remaining[m.token]--
			i += len(m.token)
			matched = true
			break
		}
		if matched {
			continue
		}
		buf.WriteByte(text[i])
		i++
	}
	flush()
	return runs
}

// plainText drops the inline marks, for places where formatting is not kept.
func plainText(text string) string {
	var b strings.Builder
	for _, r := range parseInline(text) {
		b.WriteString(r.Text)
	}
	return b.String()
}
