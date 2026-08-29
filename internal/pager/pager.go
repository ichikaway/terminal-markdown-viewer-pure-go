package pager

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

func HasLess() bool {
	_, err := exec.LookPath("less")
	return err == nil
}

func RunLess(out, errOut io.Writer, lines []string) error {
	cmd := exec.Command("less", "-R")
	cmd.Stdin = strings.NewReader(strings.Join(lines, "\n") + "\n")
	cmd.Stdout = out
	cmd.Stderr = errOut
	return cmd.Run()
}

func Run(in io.Reader, out io.Writer, lines []string, height int, color bool) error {
	if height < 1 {
		height = 1
	}
	page, pages := 0, max(1, (len(lines)+height-1)/height)
	scan := bufio.NewScanner(in)
	for {
		if color {
			fmt.Fprint(out, "\x1b[2J\x1b[H")
		}
		start := page * height
		end := min(start+height, len(lines))
		if start < len(lines) {
			fmt.Fprintln(out, strings.Join(lines[start:end], "\n"))
		}
		fmt.Fprintf(out, "[%d/%d] Enter/n next, b back, g first, G last, q quit: ", page+1, pages)
		if !scan.Scan() {
			return scan.Err()
		}
		switch scan.Text() {
		case "q":
			return nil
		case "b":
			if page > 0 {
				page--
			}
		case "g":
			page = 0
		case "G":
			page = pages - 1
		case "", "n":
			if page < pages-1 {
				page++
			} else {
				return nil
			}
		case "h":
			fmt.Fprintln(out, "Commands: Enter/n next, b previous, g first, G last, q quit")
		}
	}
}
