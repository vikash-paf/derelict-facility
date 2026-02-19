package terminal

import (
	"fmt"
	"os"

	"golang.org/x/term"
)

const hideCursor = "\033[?25l"
const showCursor = "\033[?25h"

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

func NewTerminal() *Terminal {
	return &Terminal{}
}
