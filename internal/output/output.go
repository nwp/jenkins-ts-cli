package output

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Print writes value to stdout in the appropriate format.
// If jsonMode is true, it always outputs indented JSON.
// Otherwise: strings are printed as-is, slices one item per line,
// anything else is marshalled to indented JSON.
func Print(v any, jsonMode bool) error {
	if jsonMode {
		b, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			return fmt.Errorf("json marshal: %w", err)
		}
		fmt.Println(string(b))
		return nil
	}

	switch val := v.(type) {
	case string:
		fmt.Print(val)
		if len(val) > 0 && val[len(val)-1] != '\n' {
			fmt.Println()
		}
	case []string:
		fmt.Println(strings.Join(val, "\n"))
	default:
		b, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			return fmt.Errorf("json marshal: %w", err)
		}
		fmt.Println(string(b))
	}
	return nil
}
