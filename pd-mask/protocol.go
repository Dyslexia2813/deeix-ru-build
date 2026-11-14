package main

import (
	"bytes"
	"encoding/json"
	"strings"
)

type protocol int

const (
	protoUnknown protocol = iota
	protoOpenAIChat
	protoAnthropic
)

func detectProtocol(path string) protocol {
	p := strings.TrimRight(path, "/")
	switch {
	case strings.HasSuffix(p, "/chat/completions"):
		return protoOpenAIChat
	case strings.HasSuffix(p, "/messages"):
		return protoAnthropic
	}
	return protoUnknown
}

const hintText = "Служебная пометка: персональные данные в этом диалоге заменены метками вида [FIO_…], [INN_…], [TEL_…]. " +
	"Одна и та же метка всегда обозначает одно и то же значение. Когда нужно сослаться на эти данные, пиши метку как есть, " +
	"в том числе в аргументах инструментов; не пытайся угадать настоящие значения и не придумывай новых меток. " +
	"Пользователь и инструменты увидят настоящие данные: метки заменяются автоматически."

func decodeJSON(data []byte) (any, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	var v any
	if err := dec.Decode(&v); err != nil {
		return nil, err
	}
	return v, nil
}

func encodeJSON(v any) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	return bytes.TrimRight(buf.Bytes(), "\n"), nil
}

// walkStrings applies f to every string value (not keys) of a decoded JSON value.
func walkStrings(v any, f func(string) string) any {
	switch t := v.(type) {
	case string:
		return f(t)
	case map[string]any:
		for k, item := range t {
			t[k] = walkStrings(item, f)
		}
		return t
	case []any:
		for i, item := range t {
			t[i] = walkStrings(item, f)
		}
		return t
	}
	return v
}

// maskJSONText masks the string values of a JSON document kept as text (tool call arguments).
func (m *Masker) maskJSONText(s string) string {
	v, err := decodeJSON([]byte(s))
	if err != nil {
		return m.Mask(s)
	}
	before := m.Total()
	v = walkStrings(v, m.Mask)
	if m.Total() == before {
		return s
	}
	out, err := encodeJSON(v)
	if err != nil {
		return m.Mask(s)
	}
	return string(out)
}

type requestStats struct {
	binaryParts int // images, files and audio that cannot be masked
}

func isBinaryPartType(t string) bool {
	switch t {
	case "image", "image_url", "input_image", "file", "input_file", "input_audio", "audio", "document", "video", "video_url":
		return true
	}
	return false
}

// maskOpenAIChat masks a Chat Completions request in place.
func (m *Masker) maskOpenAIChat(body map[string]any, hint bool, stats *requestStats) {
	messages, _ := body["messages"].([]any)
	for _, raw := range messages {
		msg, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		switch content := msg["content"].(type) {
		case string:
			msg["content"] = m.Mask(content)
		case []any:
			for _, rawPart := range content {
				part, ok := rawPart.(map[string]any)
				if !ok {
					continue
				}
				typ, _ := part["type"].(string)
				if text, ok := part["text"].(string); ok {
					part["text"] = m.Mask(text)
				} else if isBinaryPartType(typ) {
					stats.binaryParts++
				}
			}
		}
		for _, key := range []string{"reasoning_content", "reasoning"} {
			if text, ok := msg[key].(string); ok {
				msg[key] = m.Mask(text)
			}
		}
		if calls, ok := msg["tool_calls"].([]any); ok {
			for _, rawCall := range calls {
				call, _ := rawCall.(map[string]any)
				fn, _ := call["function"].(map[string]any)
				if args, ok := fn["arguments"].(string); ok {
					fn["arguments"] = m.maskJSONText(args)
				}
			}
		}
	}
	if !hint || m.Total() == 0 {
		return
	}
	for _, raw := range messages {
		msg, _ := raw.(map[string]any)
		if msg == nil {
			continue
		}
		if role, _ := msg["role"].(string); role != "system" && role != "developer" {
			break
		}
		switch content := msg["content"].(type) {
		case string:
			msg["content"] = content + "\n\n" + hintText
			return
		case []any:
			msg["content"] = append(content, map[string]any{"type": "text", "text": hintText})
			return
		}
	}
	body["messages"] = append([]any{map[string]any{"role": "system", "content": hintText}}, messages...)
}

// maskAnthropic masks a Messages API request in place.
func (m *Masker) maskAnthropic(body map[string]any, hint bool, stats *requestStats) {
	maskBlocks := func(blocks []any) {
		for _, rawBlock := range blocks {
			m.maskAnthropicBlock(rawBlock, stats)
		}
	}
	switch system := body["system"].(type) {
	case string:
		body["system"] = m.Mask(system)
	case []any:
		maskBlocks(system)
	}
	messages, _ := body["messages"].([]any)
	for _, raw := range messages {
		msg, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		switch content := msg["content"].(type) {
		case string:
			msg["content"] = m.Mask(content)
		case []any:
			maskBlocks(content)
		}
	}
	if !hint || m.Total() == 0 {
		return
	}
	switch system := body["system"].(type) {
	case string:
		body["system"] = system + "\n\n" + hintText
	case []any:
		body["system"] = append(system, map[string]any{"type": "text", "text": hintText})
	default:
		body["system"] = hintText
	}
}

func (m *Masker) maskAnthropicBlock(rawBlock any, stats *requestStats) {
	block, ok := rawBlock.(map[string]any)
	if !ok {
		return
	}
	typ, _ := block["type"].(string)
	switch typ {
	case "text":
		if text, ok := block["text"].(string); ok {
			block["text"] = m.Mask(text)
		}
	case "thinking":
		if text, ok := block["thinking"].(string); ok {
			block["thinking"] = m.Mask(text)
		}
	case "tool_use", "server_tool_use":
		if input, ok := block["input"]; ok {
			block["input"] = walkStrings(input, m.Mask)
		}
	case "tool_result", "web_search_tool_result":
		switch content := block["content"].(type) {
		case string:
			block["content"] = m.Mask(content)
		case []any:
			for _, inner := range content {
				m.maskAnthropicBlock(inner, stats)
			}
		}
	case "document":
		source, _ := block["source"].(map[string]any)
		if st, _ := source["type"].(string); st == "text" {
			if data, ok := source["data"].(string); ok {
				source["data"] = m.Mask(data)
				return
			}
		}
		stats.binaryParts++
	default:
		if isBinaryPartType(typ) {
			stats.binaryParts++
		}
	}
}

// unmaskOpenAIResponse restores values in a non-streamed Chat Completions response.
func (m *Masker) unmaskOpenAIResponse(body map[string]any) {
	choices, _ := body["choices"].([]any)
	for _, raw := range choices {
		choice, _ := raw.(map[string]any)
		msg, _ := choice["message"].(map[string]any)
		if msg == nil {
			continue
		}
		if text, ok := msg["content"].(string); ok {
			msg["content"] = m.Unmask(text, false)
		}
		if calls, ok := msg["tool_calls"].([]any); ok {
			for _, rawCall := range calls {
				call, _ := rawCall.(map[string]any)
				fn, _ := call["function"].(map[string]any)
				if args, ok := fn["arguments"].(string); ok {
					fn["arguments"] = m.Unmask(args, true)
				}
			}
		}
	}
}

// unmaskAnthropicResponse restores values in a non-streamed Messages response.
func (m *Masker) unmaskAnthropicResponse(body map[string]any) {
	content, _ := body["content"].([]any)
	for _, raw := range content {
		block, _ := raw.(map[string]any)
		switch block["type"] {
		case "text":
			if text, ok := block["text"].(string); ok {
				block["text"] = m.Unmask(text, false)
			}
		case "tool_use":
			if input, ok := block["input"]; ok {
				block["input"] = walkStrings(input, func(s string) string { return m.Unmask(s, false) })
			}
		}
	}
}
