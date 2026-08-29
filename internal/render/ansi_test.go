package render

import (
	"md-viewer/internal/markdown"
	"strings"
	"testing"
)

func TestJapaneseWidth(t *testing.T) {
	if got := displayWidth("A日本"); got != 5 {
		t.Fatalf("got %d", got)
	}
}

func TestTableRendering(t *testing.T) {
	d := markdown.Parse("| Name | Count |\n|:---|---:|\n| 日本 | 12 |")
	out := strings.Join(New(Options{Width: 40}).Render(d), "\n")
	if !strings.Contains(out, "┌") || !strings.Contains(out, "日本") {
		t.Fatalf("unexpected output:\n%s", out)
	}
}

func TestNoColor(t *testing.T) {
	d := markdown.Parse("# Title")
	out := strings.Join(New(Options{Width: 40}).Render(d), "\n")
	if strings.Contains(out, "\x1b[") {
		t.Fatal("unexpected ANSI escape")
	}
}

func TestTableMeasuresRenderedInlineText(t *testing.T) {
	d := markdown.Parse("| 項目 | 表示例 | 備考 |\n|---|---|---|\n| 太字 | **重要** | ANSI太字 |\n| コード | `go test` | 黄色 |")
	out := strings.Join(New(Options{Width: 80}).Render(d), "\n")
	lines := strings.Split(out, "\n")
	var boldRow, codeRow string
	for _, line := range lines {
		if strings.Contains(line, "太字") && strings.Contains(line, "重要") {
			boldRow = line
		}
		if strings.Contains(line, "コード") && strings.Contains(line, "go test") {
			codeRow = line
		}
	}
	if displayWidth(boldRow) != displayWidth(codeRow) {
		t.Fatalf("rows have different widths:\n%s\n%s", boldRow, codeRow)
	}
	if strings.Contains(boldRow, "**") {
		t.Fatalf("markdown marker leaked: %s", boldRow)
	}
}
