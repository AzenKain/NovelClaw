package jsonx

import (
	"strings"

	"github.com/bytedance/sonic"
)

func Marshal(v any) ([]byte, error) {
	return sonic.Marshal(v)
}

func Unmarshal(data []byte, v any) error {
	return sonic.Unmarshal(data, v)
}

func MarshalString(v any) (string, error) {
	return sonic.MarshalString(v)
}

func UnmarshalString(buf string, val any) error {
	return sonic.UnmarshalString(buf, val)
}

func MarshalIndent(v any, prefix, indent string) ([]byte, error) {
	return sonic.ConfigDefault.MarshalIndent(v, prefix, indent)
}

// ExtractJSONObject extracts the outermost JSON object/array substring from text that may contain
// conversational commentary, markdown code fences, or arbitrary prefixes/suffixes.
func ExtractJSONObject(raw string) string {
	s := strings.TrimSpace(raw)
	// If markdown code fences exist, extract content within the fences
	if startFence := strings.Index(s, "```"); startFence != -1 {
		rest := s[startFence+3:]
		// Skip optional lang tag like "json\n"
		if newlineIdx := strings.Index(rest, "\n"); newlineIdx != -1 {
			rest = rest[newlineIdx+1:]
		}
		if endFence := strings.Index(rest, "```"); endFence != -1 {
			s = rest[:endFence]
		} else {
			s = rest
		}
	}
	s = strings.TrimSpace(s)

	// Find the first '{' or '[' and matching outermost '}' or ']'
	firstObj := strings.Index(s, "{")
	firstArr := strings.Index(s, "[")
	lastObj := strings.LastIndex(s, "}")
	lastArr := strings.LastIndex(s, "]")

	first := -1
	last := -1
	if firstObj != -1 && (firstArr == -1 || firstObj < firstArr) {
		first = firstObj
		last = lastObj
	} else if firstArr != -1 {
		first = firstArr
		last = lastArr
	}

	if first != -1 && last > first {
		return strings.TrimSpace(s[first : last+1])
	}

	return s
}

