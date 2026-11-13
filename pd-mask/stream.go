package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// splitHold splits streamed text into the part that can be sent now and an unfinished label at
// the end that must wait for the next chunk.
func splitHold(s string) (string, string) {
	if !strings.Contains(s, "[") {
		return s, ""
	}
	loc := labelPrefixPattern.FindStringIndex(s)
	if loc == nil || len(s)-loc[0] >= maxLabelLen {
		return s, ""
	}
	return s[:loc[0]], s[loc[0]:]
}

// streamUnmasker restores labels in an SSE stream (Chat Completions or Anthropic Messages).
type streamUnmasker struct {
	m       *Masker
	proto   protocol
	w       io.Writer
	flush   func()
	pending map[string]string
	order   []string
	last    map[string]any // last Chat Completions chunk, template for synthetic chunks
}

func newStreamUnmasker(m *Masker, proto protocol, w io.Writer, flush func()) *streamUnmasker {
	return &streamUnmasker{m: m, proto: proto, w: w, flush: flush, pending: map[string]string{}}
}

func (s *streamUnmasker) take(key, chunk string, jsonString bool) string {
	text := s.pending[key] + chunk
	out, hold := splitHold(text)
	if hold == "" {
		delete(s.pending, key)
	} else {
		if _, seen := s.pending[key]; !seen {
			s.order = append(s.order, key)
		}
		s.pending[key] = hold
	}
	return s.m.Unmask(out, jsonString)
}

func (s *streamUnmasker) drain(key string) (string, bool) {
	text, ok := s.pending[key]
	if ok {
		delete(s.pending, key)
	}
	return text, ok && text != ""
}

// Run copies events from r to the writer, rewriting the ones that carry model text.
func (s *streamUnmasker) Run(r io.Reader) error {
	reader := bufio.NewReaderSize(r, 64*1024)
	var event bytes.Buffer
	for {
		line, err := reader.ReadBytes('\n')
		if len(line) > 0 {
			event.Write(line)
			if len(bytes.TrimRight(line, "\r\n")) == 0 {
				if werr := s.handleEvent(event.Bytes()); werr != nil {
					return werr
				}
				event.Reset()
			}
		}
		if err != nil {
			if event.Len() > 0 {
				if werr := s.handleEvent(event.Bytes()); werr != nil {
					return werr
				}
			}
			s.flushAll()
			if err == io.EOF {
				return nil
			}
			return err
		}
	}
}

func (s *streamUnmasker) write(b []byte) error {
	if _, err := s.w.Write(b); err != nil {
		return err
	}
	if s.flush != nil {
		s.flush()
	}
	return nil
}

func (s *streamUnmasker) handleEvent(raw []byte) error {
	lines := strings.Split(string(raw), "\n")
	eventName := ""
	dataIndex := -1
	for i, line := range lines {
		trimmed := strings.TrimRight(line, "\r")
		if strings.HasPrefix(trimmed, "event:") {
			eventName = strings.TrimSpace(trimmed[len("event:"):])
		}
		if strings.HasPrefix(trimmed, "data:") && dataIndex < 0 {
			dataIndex = i
		}
	}
	if dataIndex < 0 {
		return s.write(raw)
	}
	data := strings.TrimSpace(strings.TrimRight(lines[dataIndex], "\r")[len("data:"):])
	if data == "[DONE]" {
		if err := s.flushAll(); err != nil {
			return err
		}
		return s.write(raw)
	}
	parsed, err := decodeJSON([]byte(data))
	obj, ok := parsed.(map[string]any)
	if err != nil || !ok {
		return s.write(raw)
	}
	var before [][]byte
	changed := false
	switch s.proto {
	case protoOpenAIChat:
		before, changed = s.openAIChunk(obj)
	case protoAnthropic:
		before, changed = s.anthropicEvent(eventName, obj)
	}
	for _, b := range before {
		if err := s.write(b); err != nil {
			return err
		}
	}
	if !changed {
		return s.write(raw)
	}
	encoded, err := encodeJSON(obj)
	if err != nil {
		return s.write(raw)
	}
	lines[dataIndex] = "data: " + string(encoded)
	return s.write([]byte(strings.Join(lines, "\n")))
}

func (s *streamUnmasker) openAIChunk(obj map[string]any) ([][]byte, bool) {
	s.last = obj
	changed := false
	var before [][]byte
	choices, _ := obj["choices"].([]any)
	for _, raw := range choices {
		choice, _ := raw.(map[string]any)
		if choice == nil {
			continue
		}
		idx := fmt.Sprint(choice["index"])
		delta, _ := choice["delta"].(map[string]any)
		if delta != nil {
			if text, ok := delta["content"].(string); ok && text != "" {
				out := s.take("c"+idx, text, false)
				if out != text {
					delta["content"] = out
					changed = true
				}
			}
			if calls, ok := delta["tool_calls"].([]any); ok {
				for _, rawCall := range calls {
					call, _ := rawCall.(map[string]any)
					fn, _ := call["function"].(map[string]any)
					args, ok := fn["arguments"].(string)
					if !ok || args == "" {
						continue
					}
					out := s.take("t"+idx+"."+fmt.Sprint(call["index"]), args, true)
					if out != args {
						fn["arguments"] = out
						changed = true
					}
				}
			}
		}
		if reason, ok := choice["finish_reason"]; ok && reason != nil {
			before = append(before, s.flushChoice(idx)...)
		}
	}
	return before, changed
}

func (s *streamUnmasker) syntheticOpenAI(idx string, delta map[string]any) []byte {
	chunk := map[string]any{}
	for k, v := range s.last {
		if k != "choices" && k != "usage" {
			chunk[k] = v
		}
	}
	chunk["choices"] = []any{map[string]any{"index": json.Number(idx), "delta": delta}}
	encoded, _ := encodeJSON(chunk)
	return []byte("data: " + string(encoded) + "\n\n")
}

func (s *streamUnmasker) flushChoice(idx string) [][]byte {
	var out [][]byte
	for _, key := range s.order {
		text, ok := s.drain(key)
		if !ok {
			continue
		}
		switch {
		case key == "c"+idx:
			out = append(out, s.syntheticOpenAI(idx, map[string]any{"content": text}))
		case strings.HasPrefix(key, "t"+idx+"."):
			callIdx := strings.TrimPrefix(key, "t"+idx+".")
			out = append(out, s.syntheticOpenAI(idx, map[string]any{"tool_calls": []any{map[string]any{
				"index": json.Number(callIdx), "function": map[string]any{"arguments": text},
			}}}))
		}
	}
	return out
}

func (s *streamUnmasker) anthropicEvent(name string, obj map[string]any) ([][]byte, bool) {
	typ, _ := obj["type"].(string)
	if name == "" {
		name = typ
	}
	idx := fmt.Sprint(obj["index"])
	switch typ {
	case "content_block_delta":
		delta, _ := obj["delta"].(map[string]any)
		switch delta["type"] {
		case "text_delta":
			if text, ok := delta["text"].(string); ok && text != "" {
				out := s.take("x"+idx, text, false)
				if out != text {
					delta["text"] = out
					return nil, true
				}
			}
		case "input_json_delta":
			if text, ok := delta["partial_json"].(string); ok && text != "" {
				out := s.take("j"+idx, text, true)
				if out != text {
					delta["partial_json"] = out
					return nil, true
				}
			}
		}
	case "content_block_stop":
		return s.flushAnthropic(idx), false
	}
	return nil, false
}

func (s *streamUnmasker) syntheticAnthropic(idx string, delta map[string]any) []byte {
	encoded, _ := encodeJSON(map[string]any{"type": "content_block_delta", "index": json.Number(idx), "delta": delta})
	return []byte("event: content_block_delta\ndata: " + string(encoded) + "\n\n")
}

func (s *streamUnmasker) flushAnthropic(idx string) [][]byte {
	var out [][]byte
	if text, ok := s.drain("x" + idx); ok {
		out = append(out, s.syntheticAnthropic(idx, map[string]any{"type": "text_delta", "text": text}))
	}
	if text, ok := s.drain("j" + idx); ok {
		out = append(out, s.syntheticAnthropic(idx, map[string]any{"type": "input_json_delta", "partial_json": text}))
	}
	return out
}

// flushAll sends whatever is still held back (stream ended without a closing event).
func (s *streamUnmasker) flushAll() error {
	for _, key := range append([]string(nil), s.order...) {
		if _, ok := s.pending[key]; !ok {
			continue
		}
		var chunks [][]byte
		switch s.proto {
		case protoOpenAIChat:
			idx := strings.SplitN(key[1:], ".", 2)[0]
			chunks = s.flushChoice(idx)
		case protoAnthropic:
			chunks = s.flushAnthropic(key[1:])
		}
		for _, b := range chunks {
			if err := s.write(b); err != nil {
				return err
			}
		}
	}
	s.order = nil
	return nil
}
