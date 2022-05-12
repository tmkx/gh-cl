package main

import (
	"flag"
	"fmt"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/cli/go-gh"
	graphql "github.com/cli/shurcooL-graphql"
	"log"
	"os"
	"os/exec"
	"strings"
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
		done := false
		go func() {
			repo = getNpmRepo(pkgName)
			p.Send(endMsg{text: fmt.Sprintf("✨ Got %v\n", repo)})
			done = true
		}()
		if err := p.Start(); err != nil || !done {
			if err != nil {
				log.Fatal(err)
			}
			os.Exit(1)
		}
	}

	parts := strings.SplitN(repo, "/", 2)
	owner := parts[0]
	name := parts[1]

	tag := ""

	if isList {
		client, err := gh.GQLClient(nil)
		if err != nil {
			log.Fatal(err)
		}
		var query struct {
			Repository struct {
				Releases struct {
					Nodes []struct {
						Name        string
						TagName     string
						Url         string
						IsLatest    bool
						PublishedAt string
					}
				} `graphql:"releases(first: $perPage, orderBy: { field: CREATED_AT, direction: DESC })"`
			} `graphql:"repository(owner: $owner, name: $name)"`
		}
		variables := map[string]interface{}{
			"owner":   graphql.String(owner),
			"name":    graphql.String(name),
			"perPage": graphql.Int(30),
		}
		err = client.Query("RepositoryReleases", &query, variables)
		if err != nil {
			log.Fatal(err)
		}

		var items []list.Item
		for i := 0; i < len(query.Repository.Releases.Nodes); i++ {
			node := query.Repository.Releases.Nodes[i]
			var extra string
			if node.IsLatest {
				extra = "(latest)"
			}
			items = append(items, item{title: node.TagName, desc: fmt.Sprintf("%v %v", node.PublishedAt, extra)})
		}

		m := listModel{list: list.New(items, list.NewDefaultDelegate(), 0, 0)}
		m.choice = make(chan string)
		m.list.Title = "Please select a release"

		p := tea.NewProgram(&m, tea.WithAltScreen())

		go func() {
			if err := p.Start(); err != nil {
				fmt.Println("Error running program:", err)
				os.Exit(1)
			}
		}()

		tag = <-m.choice
	}

	cmd := exec.Command("gh", "release", "view", "-R", repo, tag)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	_ = cmd.Run()
}

// For more examples of using go-gh, see:
// https://github.com/cli/go-gh/blob/trunk/example_gh_test.go
