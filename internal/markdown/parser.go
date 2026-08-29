package markdown

import (
	"regexp"
	"strconv"
	"strings"
)

var listRE = regexp.MustCompile(`^\s*([-+*]|[0-9]+[.)])\s+(.+)$`)

func Parse(src string) Document {
	lines := strings.Split(strings.ReplaceAll(src, "\r\n", "\n"), "\n")
	var out []Node
	for i := 0; i < len(lines); {
		line := lines[i]
		if strings.HasPrefix(strings.TrimSpace(line), "```") || strings.HasPrefix(strings.TrimSpace(line), "~~~") {
			fence := strings.TrimSpace(line)[:3]
			i++
			var body []string
			for i < len(lines) && !strings.HasPrefix(strings.TrimSpace(lines[i]), fence) {
				body = append(body, lines[i])
				i++
			}
			if i < len(lines) {
				i++
			}
			out = append(out, Node{Kind: CodeBlock, Text: strings.Join(body, "\n")})
			continue
		}
		trim := strings.TrimSpace(line)
		if trim == "" {
			i++
			continue
		}
		if level, text := heading(trim); level > 0 {
			out = append(out, Node{Kind: Heading, Level: level, Text: text})
			i++
			continue
		}
		if isRule(trim) {
			out = append(out, Node{Kind: Rule})
			i++
			continue
		}
		if i+1 < len(lines) && strings.Contains(line, "|") {
			if aligns, ok := separator(lines[i+1]); ok {
				rows := [][]string{splitRow(line)}
				i += 2
				for i < len(lines) && strings.Contains(lines[i], "|") && strings.TrimSpace(lines[i]) != "" {
					rows = append(rows, splitRow(lines[i]))
					i++
				}
				out = append(out, Node{Kind: Table, Rows: rows, Align: aligns})
				continue
			}
		}
		if strings.HasPrefix(trim, ">") {
			out = append(out, Node{Kind: Quote, Text: strings.TrimSpace(strings.TrimPrefix(trim, ">"))})
			i++
			continue
		}
		if m := listRE.FindStringSubmatch(line); m != nil {
			n, ordered := 0, m[1][0] >= '0' && m[1][0] <= '9'
			if ordered {
				n, _ = strconv.Atoi(strings.TrimRight(m[1], ".)"))
			}
			out = append(out, Node{Kind: ListItem, Text: m[2], Ordered: ordered, Number: n})
			i++
			continue
		}
		var para []string
		for i < len(lines) && strings.TrimSpace(lines[i]) != "" {
			if len(para) > 0 && startsBlock(lines, i) {
				break
			}
			para = append(para, strings.TrimSpace(lines[i]))
			i++
		}
		out = append(out, Node{Kind: Paragraph, Text: strings.Join(para, " ")})
	}
	return Document{Nodes: out}
}

func heading(s string) (int, string) {
	n := 0
	for n < len(s) && n < 6 && s[n] == '#' {
		n++
	}
	if n > 0 && n < len(s) && s[n] == ' ' {
		return n, strings.TrimSpace(s[n:])
	}
	return 0, ""
}

func isRule(s string) bool {
	x := strings.ReplaceAll(strings.ReplaceAll(s, " ", ""), "\t", "")
	return len(x) >= 3 && (strings.Trim(x, "-") == "" || strings.Trim(x, "*") == "" || strings.Trim(x, "_") == "")
}

func startsBlock(lines []string, i int) bool {
	t := strings.TrimSpace(lines[i])
	if l, _ := heading(t); l > 0 {
		return true
	}
	if strings.HasPrefix(t, ">") || strings.HasPrefix(t, "```") || strings.HasPrefix(t, "~~~") || isRule(t) || listRE.MatchString(lines[i]) {
		return true
	}
	if i+1 < len(lines) {
		_, ok := separator(lines[i+1])
		return ok && strings.Contains(lines[i], "|")
	}
	return false
}

func separator(s string) ([]Align, bool) {
	cells := splitRow(s)
	if len(cells) == 0 {
		return nil, false
	}
	a := make([]Align, len(cells))
	for i, c := range cells {
		c = strings.TrimSpace(c)
		left, right := strings.HasPrefix(c, ":"), strings.HasSuffix(c, ":")
		core := strings.Trim(c, ":")
		if len(core) < 3 || strings.Trim(core, "-") != "" {
			return nil, false
		}
		if left && right {
			a[i] = AlignCenter
		} else if right {
			a[i] = AlignRight
		}
	}
	return a, true
}

func splitRow(s string) []string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "|")
	s = strings.TrimSuffix(s, "|")
	var cells []string
	var b strings.Builder
	escaped := false
	for _, r := range s {
		if escaped {
			if r != '|' {
				b.WriteRune('\\')
			}
			b.WriteRune(r)
			escaped = false
			continue
		}
		if r == '\\' {
			escaped = true
			continue
		}
		if r == '|' {
			cells = append(cells, strings.TrimSpace(b.String()))
			b.Reset()
		} else {
			b.WriteRune(r)
		}
	}
	if escaped {
		b.WriteRune('\\')
	}
	return append(cells, strings.TrimSpace(b.String()))
}
