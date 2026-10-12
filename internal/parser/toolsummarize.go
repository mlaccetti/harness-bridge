package parser

import (
	"encoding/json"
	"strings"
)

func SummarizeClaudeTool(name string, input map[string]interface{}) string {
	nameLower := strings.ToLower(name)
	switch nameLower {
	case "bash", "shell":
		cmd := strings.TrimSpace(stringVal(input, "command"))
		if cmd == "" {
			cmd = strings.TrimSpace(stringVal(input, "cmd"))
		}
		return "ran `" + truncate(cmd, 160) + "`"
	case "edit", "write":
		path := stringVal(input, "file_path")
		if path == "" {
			path = stringVal(input, "path")
		}
		if path == "" {
			path = "?"
		}
		return nameLower + " " + path
	case "read":
		return "read " + stringVal(input, "file_path", "?")
	case "grep":
		return "grep `" + truncate(stringVal(input, "pattern"), 80) + "`"
	case "glob":
		return "glob `" + truncate(stringVal(input, "pattern"), 80) + "`"
	case "webfetch", "websearch":
		u := stringVal(input, "url")
		if u == "" {
			u = stringVal(input, "query", "?")
		}
		return nameLower + " " + u
	default:
		argsPreview, _ := json.Marshal(input)
		return name + " " + truncate(string(argsPreview), 120)
	}
}

func SummarizeCodexFunction(name string, args map[string]interface{}) string {
	switch name {
	case "shell":
		cmdList, ok := args["command"]
		var cmd string
		if ok {
			switch v := cmdList.(type) {
			case []interface{}:
				parts := make([]string, 0, len(v))
				for _, x := range v {
					parts = append(parts, fmtAny(x))
				}
				cmd = strings.Join(parts, " ")
			default:
				cmd = fmtAny(cmdList)
			}
		}
		return "ran `" + truncate(cmd, 160) + "`"
	case "apply_patch":
		return "applied patch"
	default:
		raw, _ := json.Marshal(args)
		return name + " " + truncate(string(raw), 120)
	}
}

func stringVal(m map[string]interface{}, key string, def ...string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	if len(def) > 0 {
		return def[0]
	}
	return ""
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

func fmtAny(v interface{}) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(jsonMarshal(v), "\"", ""), "\n", " "))
}

func jsonMarshal(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(b)
}
