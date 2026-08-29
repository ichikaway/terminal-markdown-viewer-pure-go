package render

import (
	"strings"
	"unicode"
)

func displayWidth(s string) int {
	w := 0
	for _, r := range stripANSI(s) {
		if unicode.Is(unicode.Mn, r) || unicode.Is(unicode.Me, r) {
			continue
		}
		if isWide(r) {
			w += 2
		} else {
			w++
		}
	}
	return w
}

func isWide(r rune) bool {
	return r >= 0x1100 && (r <= 0x115f || r == 0x2329 || r == 0x232a ||
		(r >= 0x2e80 && r <= 0xa4cf) || (r >= 0xac00 && r <= 0xd7a3) ||
		(r >= 0xf900 && r <= 0xfaff) || (r >= 0xfe10 && r <= 0xfe6f) ||
		(r >= 0xff00 && r <= 0xff60) || (r >= 0xffe0 && r <= 0xffe6) ||
		(r >= 0x1f300 && r <= 0x1faff) || (r >= 0x20000 && r <= 0x3fffd))
}

func stripANSI(s string) string {
	var b strings.Builder
	esc := false
	for _, r := range s {
		if r == 0x1b {
			esc = true
			continue
		}
		if esc {
			if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') {
				esc = false
			}
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

func pad(s string, n int) string { return s + strings.Repeat(" ", max(0, n-displayWidth(s))) }

func wrapPlain(s string, width int) []string {
	if width < 1 {
		width = 1
	}
	var out []string
	var b strings.Builder
	w := 0
	flush := func() { out = append(out, strings.TrimRight(b.String(), " ")); b.Reset(); w = 0 }
	for _, r := range s {
		rw := 1
		if isWide(r) {
			rw = 2
		}
		if w+rw > width && b.Len() > 0 {
			flush()
		}
		b.WriteRune(r)
		w += rw
	}
	if b.Len() > 0 || len(out) == 0 {
		flush()
	}
	return out
}
