package markdown

type Kind int

const (
	Paragraph Kind = iota
	Heading
	CodeBlock
	Quote
	ListItem
	Rule
	Table
)

type Align int

const (
	AlignLeft Align = iota
	AlignCenter
	AlignRight
)

type Node struct {
	Kind    Kind
	Level   int
	Text    string
	Ordered bool
	Number  int
	Rows    [][]string
	Align   []Align
}

type Document struct{ Nodes []Node }
