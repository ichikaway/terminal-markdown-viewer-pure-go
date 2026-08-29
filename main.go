package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"md-viewer/internal/markdown"
	"md-viewer/internal/pager"
	"md-viewer/internal/render"
	"md-viewer/internal/terminal"
	"md-viewer/internal/watch"
)

func main() {
	var noColor, ascii, noPager, watchMode bool
	var width int
	flag.BoolVar(&noColor, "no-color", false, "disable ANSI colors")
	flag.BoolVar(&ascii, "ascii", false, "use ASCII table borders")
	flag.BoolVar(&noPager, "no-pager", false, "print all output without paging")
	flag.BoolVar(&watchMode, "watch", false, "watch file and refresh when it changes")
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
	if width <= 0 {
		width = envInt("COLUMNS", 80)
	}
	color := !noColor && os.Getenv("NO_COLOR") == "" && os.Getenv("TERM") != "dumb" && isTerminal(os.Stdout)
	if watchMode {
		if len(flag.Args()) != 1 || flag.Arg(0) == "-" {
			exitError("-watch requires a file path")
		}
		if !isTerminal(os.Stdin) || !isTerminal(os.Stdout) {
			exitError("-watch requires terminal input and output")
		}
		if err := runWatch(flag.Arg(0), width, color, ascii); err != nil {
			exitError(err.Error())
		}
		return
	}

	data, err := readInput(flag.Args())
	if err != nil {
		fmt.Fprintln(os.Stderr, "md-viewer:", err)
		os.Exit(1)
	}
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

func runWatch(path string, width int, color, ascii bool) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	terminalRows := envInt("LINES", 24)
	if _, rows, err := terminal.Size(os.Stdout); err == nil && rows > 0 {
		terminalRows = rows
	}
	// Reserve one status line and two blank lines below the document.
	height := max(1, terminalRows-3)
	restore, err := terminal.MakeRaw(os.Stdin)
	if err != nil {
		return fmt.Errorf("enable interactive input: %w", err)
	}
	defer restore()
	fmt.Print("\x1b[?25l")
	defer fmt.Print("\x1b[?25h\x1b[2J\x1b[H")
	updates := make(chan []byte, 1)
	errors := make(chan error, 1)
	watchDone := make(chan error, 1)
	go func() {
		watchDone <- watch.Run(ctx, path, 250*time.Millisecond, func(data []byte) error {
			copyOfData := append([]byte(nil), data...)
			select {
			case updates <- copyOfData:
			default:
				<-updates
				updates <- copyOfData
			}
			return nil
		}, func(err error) {
			select {
			case errors <- err:
			default:
			}
		})
	}()
	keys := make(chan byte)
	go readKeys(os.Stdin, keys)
	var lines []string
	offset, escape := 0, 0
	status := watchStatus(path, color)
	redraw := func() {
		offset = max(0, min(offset, max(0, len(lines)-height)))
		end := min(offset+height, len(lines))
		fmt.Print("\x1b[2J\x1b[H", strings.Join(lines[offset:end], "\n"))
		position := fmt.Sprintf(" — lines %d-%d/%d", min(offset+1, len(lines)), end, len(lines))
		fmt.Printf("\x1b[%d;1H\x1b[2K%s%s", height+1, status, position)
	}
	for {
		select {
		case <-ctx.Done():
			return nil
		case err := <-watchDone:
			return err
		case err := <-errors:
			status = "watch: " + err.Error()
			redraw()
		case data := <-updates:
			lines = render.New(render.Options{Width: width, Color: color, ASCII: ascii}).Render(markdown.Parse(string(data)))
			status = watchStatus(path, color)
			redraw()
		case key := <-keys:
			if escape == 1 {
				if key == '[' {
					escape = 2
				} else {
					escape = 0
				}
				continue
			}
			if escape == 2 {
				escape = 0
				if key == 'A' {
					offset--
				}
				if key == 'B' {
					offset++
				}
				redraw()
				continue
			}
			switch key {
			case 0x1b:
				escape = 1
				continue
			case 'q':
				return nil
			case 'j':
				offset++
			case 'k':
				offset--
			case ' ':
				offset += height
			case 'b':
				offset -= height
			case 'g':
				offset = 0
			case 'G':
				offset = max(0, len(lines)-height)
			default:
				continue
			}
			redraw()
		}
	}
}

func readKeys(in *os.File, keys chan<- byte) {
	var b [1]byte
	for {
		if _, err := in.Read(b[:]); err != nil {
			return
		}
		keys <- b[0]
	}
}

func watchStatus(path string, color bool) string {
	s := "watching " + path + " — j/k scroll, Space/b page, g/G ends, q quit"
	if color {
		return "\x1b[7m" + s + "\x1b[0m"
	}
	return s
}

func exitError(message string) {
	fmt.Fprintln(os.Stderr, "md-viewer:", message)
	os.Exit(1)
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
