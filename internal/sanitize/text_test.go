package sanitize

import (
	"strings"
	"testing"
)

func TestDocumentNeutralizesTerminalControls(t *testing.T) {
	input := "safe\n\t\x1b]52;c;secret\a\x1b[2J\u009b31m"
	got := Document(input)
	if strings.ContainsAny(got, "\x00\x01\x02\x03\x04\x05\x06\x07\x08\x0b\x0c\x0d\x0e\x0f\x10\x11\x12\x13\x14\x15\x16\x17\x18\x19\x1a\x1b\x1c\x1d\x1e\x1f\x7f") {
		t.Fatalf("unsafe control remains in %q", got)
	}
	if !strings.Contains(got, "^[") || !strings.Contains(got, "<U+009B>") {
		t.Fatalf("controls were not made visible: %q", got)
	}
	if !strings.Contains(got, "\n\t") {
		t.Fatalf("Markdown whitespace was not preserved: %q", got)
	}
}

func TestLineNeutralizesNewlinesAndTabs(t *testing.T) {
	got := Line("name\n\t\x1b[2J")
	if strings.ContainsAny(got, "\n\t\x1b") {
		t.Fatalf("status line contains a control: %q", got)
	}
}

func TestDocumentPreservesWindowsLineEndings(t *testing.T) {
	if got := Document("first\r\nsecond\rthird"); got != "first\nsecond^Mthird" {
		t.Fatalf("unexpected line-ending handling: %q", got)
	}
}
