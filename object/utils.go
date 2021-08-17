package object

import (
	"bytes"
	"encoding/json"
	"strings"
)

func indentJSON(inp string) string {
	j := []byte(inp)
	var prettyJSON bytes.Buffer
	err := json.Indent(&prettyJSON, j, "", "    ")
	if err != nil {
		return "JSON parse error: " + err.Error()
	}

	return prettyJSON.String()
}

func escapeQuotes(inp string) string {
	return strings.ReplaceAll(inp, `"`, `\"`)
}
