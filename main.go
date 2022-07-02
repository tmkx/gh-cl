package main

import (
	"flag"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/tmkx/gh-cl/ui"
	"log"
	"os"
)

func main() {
	flag.Parse()
	if flag.NArg() == 0 {
		log.Fatalln("Please enter the package's name")
	}
	pkgName := flag.Args()[0]

	p := tea.NewProgram(ui.InitModel(pkgName), tea.WithAltScreen())
	if err := p.Start(); err != nil {
		os.Exit(1)
	}
}

// For more examples of using go-gh, see:
// https://github.com/cli/go-gh/blob/trunk/example_gh_test.go
