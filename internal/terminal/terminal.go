package terminal

import (
	"fmt"
	"os"

	"golang.org/x/term"
)

const hideCursor = "\033[?25l"
const showCursor = "\033[?25h"
const clearScreen = "\033[2J\033[H"

func NewTerminal() *Terminal {
	return &Terminal{}
}

type Terminal struct {
	oldState *term.State
}

// EnableRawMode https://pkg.go.dev/golang.org/x/term#MakeRaw
// Raw Mode: Passes every keypress immediately to the program.
func (t *Terminal) EnableRawMode() error {
	if t == nil {
		return fmt.Errorf("terminal is nil")
	}

	fd := int(os.Stdin.Fd())

	if !term.IsTerminal(fd) {
		return fmt.Errorf("stdin is not a terminal (run this in a real shell, not an IDE output tab)")
	}

	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return err
	}

	t.oldState = oldState

	// To hide the cursor, print "\033[?25l" to the terminal, and "\033[?25h" to show the cursor.
	// https://man7.org/linux/man-pages/man4/console_codes.4.html

	fmt.Print(hideCursor)

	return nil
}

func (t *Terminal) Restore() {
	if t != nil && t.oldState != nil {
		fd := int(os.Stdin.Fd())
		term.Restore(fd, t.oldState)
	}

	fmt.Print(showCursor)
}

func (t *Terminal) Clear() {
	fmt.Print(clearScreen)
}

func (t *Terminal) PollInput() <-chan InputEvent {
	// constantly poll for new input key

	// TODO: implement polling logic

	return nil
}

type InputEvent struct {
	Key rune
	// we only care about alpha-numeric keys, not special keys like arrows, esc etc.
	// They will be added later.
}
