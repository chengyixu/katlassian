package output

import (
	"encoding/json"
	"fmt"
	"os"
)

func JSON(data interface{}) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(data); err != nil {
		fmt.Fprintf(os.Stderr, "error encoding JSON: %v\n", err)
		os.Exit(1)
	}
}

func RawJSON(data json.RawMessage) {
	var parsed interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		os.Stdout.Write(data)
		return
	}
	JSON(parsed)
}

func Error(msg string) {
	result := map[string]interface{}{
		"ok":    false,
		"error": msg,
	}
	enc := json.NewEncoder(os.Stderr)
	enc.SetIndent("", "  ")
	enc.Encode(result)
}
