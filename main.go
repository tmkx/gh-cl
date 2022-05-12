package main

import (
	"flag"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"log"
	"os"
	"os/exec"
)

func main() {
	flag.Parse()
	if flag.NArg() == 0 {
		log.Fatalln("Please enter the package's name")
	}
	pkgName := flag.Args()[0]
	isList := contains(flag.Args(), "--list")

	repo := pkgName

	if isNpm(pkgName) {
		p := tea.NewProgram(initialSpinnerModel(fmt.Sprintf("Querying %v's repo...", pkgName)))
		go func() {
			repo = getNpmRepo(pkgName)
			p.Send(endMsg{text: fmt.Sprintf("✨ Got %v\n", repo)})
		}()
		if err := p.Start(); err != nil {
			fmt.Println("err")
			os.Exit(1)
		}
	}

	tag := ""

	if isList {
		//
	}

	cmd := exec.Command("gh", "release", "view", "-R", repo, tag)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	_ = cmd.Run()
}

// For more examples of using go-gh, see:
// https://github.com/cli/go-gh/blob/trunk/example_gh_test.go
