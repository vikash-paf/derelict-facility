package main

import (
	"fmt"
	"log"

	"github.com/vikash-paf/derelict-facility/internal/engine"
	"github.com/vikash-paf/derelict-facility/internal/terminal"
)

func main() {
	term, err := terminal.New()
	if err != nil {
		log.Fatalf("failed to initialize terminal: %v", err)
	}
	defer term.Restore()

	fmt.Println("Starting derelict-facility engine...")

	eng := engine.New(term)
	if err := eng.Run(); err != nil {
		log.Fatalf("engine error: %v", err)
	}
}
