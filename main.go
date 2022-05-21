package main

import (
	"flag"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"log"
	"os"
	"os/exec"
)

var chRepo string
var chTag string

func main() {
	flag.Parse()
	if flag.NArg() == 0 {
		log.Fatalln("Please enter the package's name")
	}
	pkgName := flag.Args()[0]

	p := tea.NewProgram(initModel(pkgName))
	if err := p.Start(); err != nil {
		os.Exit(1)
	}

	if chRepo == "" || chTag == "" {
		os.Exit(0)
	}

	fmt.Println("Fetching changelog...")
	cmd := exec.Command("gh", "release", "view", "-R", chRepo, chTag)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	_ = cmd.Run()
}

// For more examples of using go-gh, see:
// https://github.com/cli/go-gh/blob/trunk/example_gh_test.go
