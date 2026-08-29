package render

import (
	"regexp"
	"strconv"
	"strings"

	"md-viewer/internal/markdown"
)

type Options struct {
	Width        int
	Color, ASCII bool
}
type Renderer struct{ o Options }

func New(o Options) *Renderer {
	if o.Width < 20 {
		o.Width = 20
	}
	return &Renderer{o: o}
}

const reset = "\x1b[0m"

func (r *Renderer) style(code, s string) string {
	if !r.o.Color {
		return s
	}
	return "\x1b[" + code + "m" + s + reset
}

func (r *Renderer) Render(doc markdown.Document) []string {
	var out []string
	for idx, n := range doc.Nodes {
		if idx > 0 && n.Kind != markdown.ListItem {
			out = append(out, "")
		}
		switch n.Kind {
		case markdown.Heading:
			code := "1;36"
			if n.Level == 2 {
				code = "1;34"
			}
			if n.Level >= 3 {
				code = "1;37"
			}
			text := r.inline(n.Text)
			out = append(out, r.style(code, text))
			if n.Level == 1 {
				out = append(out, r.style("36", strings.Repeat("═", min(r.o.Width, displayWidth(n.Text)))))
			}
		case markdown.Paragraph:
			out = append(out, r.wrapInline(n.Text, r.o.Width)...)
		case markdown.Quote:
			for _, x := range r.wrapInline(n.Text, r.o.Width-2) {
				out = append(out, r.style("90", "│ "+x))
			}
		case markdown.ListItem:
			prefix := "• "
			if n.Ordered {
				prefix = strconv.Itoa(n.Number) + ". "
			}
			for i, x := range r.wrapInline(n.Text, r.o.Width-displayWidth(prefix)) {
				if i == 0 {
					out = append(out, prefix+x)
				} else {
					out = append(out, strings.Repeat(" ", displayWidth(prefix))+x)
				}
			}
		case markdown.CodeBlock:
			for _, x := range strings.Split(n.Text, "\n") {
				out = append(out, r.style("33", "  "+x))
			}
		case markdown.Rule:
			ch := "─"
			if r.o.ASCII {
				ch = "-"
			}
			out = append(out, r.style("90", strings.Repeat(ch, r.o.Width)))
		case markdown.Table:
			out = append(out, r.table(n)...)
		}
	}
	return out
}

var inlineRE = regexp.MustCompile(`(!?\[([^]]+)\]\(([^)]+)\)|` + "`([^`]+)`" + `|\*\*([^*]+)\*\*|__([^_]+)__|~~([^~]+)~~|\*([^*]+)\*|_([^_]+)_)`)

func (r *Renderer) inline(s string) string {
	return inlineRE.ReplaceAllStringFunc(s, func(m string) string {
		g := inlineRE.FindStringSubmatch(m)
		switch {
		case g[2] != "":
			if strings.HasPrefix(m, "!") {
				return "[image: " + g[2] + "]"
			}
			return r.style("4;34", g[2]) + " (" + g[3] + ")"
		case g[4] != "":
			return r.style("33", g[4])
		case g[5] != "":
			return r.style("1", g[5])
		case g[6] != "":
			return r.style("1", g[6])
		case g[7] != "":
			return r.style("9", g[7])
		case g[8] != "":
			return r.style("3", g[8])
		default:
			return r.style("3", g[9])
		}
	})
}

func (r *Renderer) wrapInline(s string, width int) []string {
	// Wrap before styling so escape sequences never affect width calculations.
	plain := wrapPlain(s, width)
	out := make([]string, len(plain))
	for i := range plain {
		out[i] = r.inline(plain[i])
	}
	return out
}

func (r *Renderer) table(n markdown.Node) []string {
	cols := len(n.Align)
	if cols == 0 {
		return nil
	}
	widths := make([]int, cols)
	for _, row := range n.Rows {
		for c := 0; c < cols && c < len(row); c++ {
			// Markdown markers and ANSI escapes do not occupy terminal cells.
			widths[c] = max(widths[c], displayWidth(r.inline(row[c])))
		}
	}
	available := r.o.Width - 3*cols - 1
	for sum(widths) > available {
		c := widest(widths)
		if widths[c] <= 3 {
			break
		}
		widths[c]--
	}
	chars := [6]string{"┌", "┬", "┐", "├", "┼", "┤"}
	bottom := [3]string{"└", "┴", "┘"}
	if r.o.ASCII {
		chars = [6]string{"+", "+", "+", "+", "+", "+"}
		bottom = [3]string{"+", "+", "+"}
	}
	border := func(l, mid, rr string) string {
		var b strings.Builder
		b.WriteString(l)
		for i, w := range widths {
			b.WriteString(strings.Repeat(func() string {
				if r.o.ASCII {
					return "-"
				}
				return "─"
			}(), w+2))
			if i < cols-1 {
				b.WriteString(mid)
			}
		}
		b.WriteString(rr)
		return b.String()
	}
	var out []string
	out = append(out, border(chars[0], chars[1], chars[2]))
	for ri, row := range n.Rows {
		wrapped := make([][]string, cols)
		h := 1
		for c := 0; c < cols; c++ {
			v := ""
			if c < len(row) {
				v = row[c]
			}
			rendered := r.inline(v)
			if displayWidth(rendered) <= widths[c] {
				wrapped[c] = []string{rendered}
			} else {
				// Avoid carrying an ANSI style across a physical row boundary.
				wrapped[c] = wrapPlain(stripANSI(rendered), widths[c])
			}
			h = max(h, len(wrapped[c]))
		}
		for line := 0; line < h; line++ {
			var b strings.Builder
			b.WriteString("│")
			if r.o.ASCII {
				b.Reset()
				b.WriteString("|")
			}
			for c := 0; c < cols; c++ {
				v := ""
				if line < len(wrapped[c]) {
					v = wrapped[c][line]
				}
				left := 0
				gap := widths[c] - displayWidth(v)
				if n.Align[c] == markdown.AlignRight {
					left = gap
				} else if n.Align[c] == markdown.AlignCenter {
					left = gap / 2
				}
				b.WriteString(" " + strings.Repeat(" ", left) + v + strings.Repeat(" ", gap-left) + " ")
				if r.o.ASCII {
					b.WriteString("|")
				} else {
					b.WriteString("│")
				}
			}
			out = append(out, b.String())
		}
		if ri == 0 {
			out = append(out, border(chars[3], chars[4], chars[5]))
		}
	}
	out = append(out, border(bottom[0], bottom[1], bottom[2]))
	return out
}

func sum(xs []int) int {
	n := 0
	for _, x := range xs {
		n += x
	}
	return n
}
func widest(xs []int) int {
	best := 0
	for i := range xs {
		if xs[i] > xs[best] {
			best = i
		}
	}
	return best
}
