package sanitize

import (
	"fmt"
	"strings"
)

// Document makes untrusted text safe for terminal output while preserving
// newlines and tabs used by Markdown formatting.
func Document(s string) string {
	return clean(strings.ReplaceAll(s, "\r\n", "\n"), true)
}

// Line makes untrusted text safe for use in a single terminal status line.
func Line(s string) string { return clean(s, false) }

func clean(s string, multiline bool) string {
	var b strings.Builder
	for _, r := range s {
		if multiline && (r == '\n' || r == '\t') {
			b.WriteRune(r)
			continue
		}
		switch {
		case r == 0x7f:
			b.WriteString("^?")
		case r < 0x20:
			b.WriteByte('^')
			b.WriteByte(byte(r) + '@')
		case r >= 0x80 && r <= 0x9f:
			fmt.Fprintf(&b, "<U+%04X>", r)
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}
