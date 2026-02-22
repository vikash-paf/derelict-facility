package terminal

import (
	"fmt"
	"os"

	"github.com/vikash-paf/derelict-facility/internal/core"
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

// GetSize returns the current width (columns) and height (rows) of the terminal.
func (t *Terminal) GetSize() (width, height int, err error) {
	return term.GetSize(int(os.Stdout.Fd()))
}

// MoveCursorTo returns the string code to jump to a specific spot.
// It automatically handles the math to convert your 0-based game grid to the 1-based terminal grid.
func (t *Terminal) MoveCursorTo(x, y int) string {
	return fmt.Sprintf("\033[%d;%dH", y+1, x+1)
}

func (t *Terminal) PollInput() <-chan core.InputEvent {
	events := make(chan core.InputEvent, 10)

	go func() {
		for {
			b := make([]byte, 1)
			n, err := os.Stdin.Read(b)

			// Safety check: if reading fails, don't crash, just wait.
			if err != nil || n == 0 {
				continue
			}

			// Catch Ctrl+C (ASCII value 3) or ESC (ASCII value 27)
			if b[0] == 3 {
				events <- core.InputEvent{Quit: true}
				continue
			}

			events <- core.InputEvent{Key: rune(b[0])}
		}
	}()

	return events
}
