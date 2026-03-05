package waybar

import (
	"encoding/json"
	"fmt"
)

// Module represents a waybar custom module JSON output.
type Module struct {
	Text       string `json:"text"`
	Tooltip    string `json:"tooltip,omitempty"`
	Class      string `json:"class,omitempty"`
	Percentage int    `json:"percentage,omitempty"`
}

// PrintJSON emits waybar-compatible JSON to stdout.
func PrintJSON(m Module) {
	data, _ := json.Marshal(m) //nolint:errcheck
	fmt.Println(string(data))
}
