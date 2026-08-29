package markdown

import "testing"

func TestParseTableAndBlocks(t *testing.T) {
	d := Parse("# 見出し\n\n- item\n\n| 名前 | 数 |\n|:---|---:|\n| 太郎 | 12 |")
	if len(d.Nodes) != 3 {
		t.Fatalf("got %d nodes", len(d.Nodes))
	}
	if d.Nodes[0].Kind != Heading || d.Nodes[0].Level != 1 {
		t.Fatal("heading not parsed")
	}
	if d.Nodes[2].Kind != Table || d.Nodes[2].Align[1] != AlignRight {
		t.Fatal("table not parsed")
	}
}

func TestEscapedPipe(t *testing.T) {
	d := Parse("| A | B |\n|---|---|\n| x\\|y | z |")
	if got := d.Nodes[0].Rows[1][0]; got != "x|y" {
		t.Fatalf("got %q", got)
	}
}
