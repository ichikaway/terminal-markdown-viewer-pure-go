package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"md-viewer/internal/markdown"
	"md-viewer/internal/pager"
	"md-viewer/internal/render"
)

func main() {
	var noColor, ascii, noPager bool
	var width int
	flag.BoolVar(&noColor, "no-color", false, "disable ANSI colors")
	flag.BoolVar(&ascii, "ascii", false, "use ASCII table borders")
	flag.BoolVar(&noPager, "no-pager", false, "print all output without paging")
	flag.IntVar(&width, "width", 0, "display width (default: COLUMNS or 80)")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [options] [file]\n\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() > 1 {
		flag.Usage()
		os.Exit(2)
	}

	data, err := readInput(flag.Args())
	if err != nil {
		fmt.Fprintln(os.Stderr, "md-viewer:", err)
		os.Exit(1)
	}
	if width <= 0 {
		width = envInt("COLUMNS", 80)
	}
	color := !noColor && os.Getenv("NO_COLOR") == "" && os.Getenv("TERM") != "dumb" && isTerminal(os.Stdout)
	doc := markdown.Parse(string(data))
	lines := render.New(render.Options{Width: width, Color: color, ASCII: ascii}).Render(doc)

	interactive := !noPager && isTerminal(os.Stdout)
	if !interactive {
		fmt.Println(strings.Join(lines, "\n"))
		return
	}
	if pager.HasLess() {
		if err := pager.RunLess(os.Stdout, os.Stderr, lines); err != nil {
			fmt.Fprintln(os.Stderr, "md-viewer:", err)
			os.Exit(1)
		}
		return
	}
	if !isTerminal(os.Stdin) {
		fmt.Println(strings.Join(lines, "\n"))
		return
	}
	height := envInt("LINES", 24) - 2
	if err := pager.Run(os.Stdin, os.Stdout, lines, height, color); err != nil {
		fmt.Fprintln(os.Stderr, "md-viewer:", err)
		os.Exit(1)
	}
}

func readInput(args []string) ([]byte, error) {
	if len(args) == 0 || args[0] == "-" {
		return io.ReadAll(bufio.NewReader(os.Stdin))
	}
	return os.ReadFile(args[0])
}

func envInt(name string, fallback int) int {
	if n, err := strconv.Atoi(os.Getenv(name)); err == nil && n > 0 {
		return n
	}
	return fallback
}

func isTerminal(f *os.File) bool {
	info, err := f.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}
