package main

import (
	"fmt"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"os"
	"os/exec"
)

const (
	QueryingPackage = iota
	FetchingReleases
	ChoosingRelease
	FetchingChangelog
	ShowChangelog
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)

type model struct {
	state     int
	pkgName   string
	repo      string
	releases  []list.Item
	changelog string
	spinner   spinner.Model
	list      list.Model
	quitting  bool
	err       error
}

type item struct {
	title string
	desc  string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

type repoMsg string
type releasesMsg []list.Item
type chooseReleaseMsg string

func initModel(pkgName string) model {
	s := spinner.New()
	s.Spinner = spinner.MiniDot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	l := list.New(make([]list.Item, 0), list.NewDefaultDelegate(), 0, 0)
	l.Title = "Please select a release"

	return model{spinner: s, list: l, pkgName: pkgName, state: QueryingPackage}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(getRepoCmd(m.pkgName), m.spinner.Tick)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "enter":
			if m.state == ChoosingRelease {
				i, ok := m.list.SelectedItem().(item)
				if ok {
					cmds = append(cmds, func() tea.Msg {
						return chooseReleaseMsg(i.title)
					})
				}
			}
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	case repoMsg:
		m.state = FetchingReleases
		m.repo = string(msg)
		cmds = append(cmds, getReleasesCmd(m.repo))
	case releasesMsg:
		m.state = ChoosingRelease
		m.releases = msg
		m.list.SetItems(msg)
	case chooseReleaseMsg:
		m.state = FetchingChangelog
		cmds = append(cmds, showReleaseCmd(m.repo, string(msg)))
	}

	switch m.state {
	case QueryingPackage, FetchingReleases, FetchingChangelog:
		newSpinner, newCmd := m.spinner.Update(msg)
		m.spinner = newSpinner
		cmd = newCmd
	case ChoosingRelease:
		newList, newCmd := m.list.Update(msg)
		m.list = newList
		cmd = newCmd
	}

	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	if m.err != nil {
		return m.err.Error()
	}
	switch m.state {
	case QueryingPackage:
		return fmt.Sprintf("%s Querying %s's repo...", m.spinner.View(), m.pkgName)
	case FetchingReleases:
		return fmt.Sprintf("%s Querying %s's releases...", m.spinner.View(), m.repo)
	case ChoosingRelease:
		return docStyle.Render(m.list.View())
	case FetchingChangelog:
		return fmt.Sprintf("%s Fetching changelog...", m.spinner.View())
	}
	return ""
}

func getRepoCmd(pkgName string) tea.Cmd {
	return func() tea.Msg {
		repo := pkgName
		if isNpm(repo) {
			repo = getNpmRepo(pkgName)
		}
		return repoMsg(repo)
	}
}

func getReleasesCmd(repo string) tea.Cmd {
	return func() tea.Msg {
		releases := getReleases(repo)
		var items []list.Item
		for i := 0; i < len(releases); i++ {
			release := releases[i]
			var extra string
			if release.IsLatest {
				extra = "(latest)"
			}
			items = append(items, item{title: release.TagName, desc: fmt.Sprintf("%v %v", release.PublishedAt, extra)})
		}
		return releasesMsg(items)
	}
}

func showReleaseCmd(repo string, tag string) tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command("gh", "release", "view", "-R", repo, tag)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		_ = cmd.Run()
		return tea.Quit()
	}
}
