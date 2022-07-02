package ui

import (
	"fmt"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/tmkx/gh-cl/util"
)

const (
	queryingPackage = iota
	fetchingReleases
	choosingRelease
	fetchingChangelog
	showChangelog
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
type showChangelogMsg string

func InitModel(pkgName string) model {
	s := spinner.New()
	s.Spinner = spinner.MiniDot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	l := list.New(make([]list.Item, 0), list.NewDefaultDelegate(), 0, 0)
	l.Title = "Please select a release"

	return model{spinner: s, list: l, pkgName: pkgName, state: queryingPackage}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(getRepoCmd(m.pkgName), spinner.Tick)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			if m.state == showChangelog {
				m.state = choosingRelease
				return m, nil
			}
			m.quitting = true
			return m, tea.Quit
		case "enter":
			if m.state == choosingRelease {
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
		m.state = fetchingReleases
		m.repo = string(msg)
		m.list.Title = fmt.Sprintf("Releases of %v", m.repo)
		cmds = append(cmds, getReleasesCmd(m.repo))
	case releasesMsg:
		m.state = choosingRelease
		m.releases = msg
		m.list.SetItems(msg)
	case chooseReleaseMsg:
		m.state = fetchingChangelog
		cmds = append(cmds, spinner.Tick, getChangelogCmd(m.repo, string(msg)))
	case showChangelogMsg:
		m.state = showChangelog
		m.changelog = string(msg)
		fmt.Println(m.changelog)
	}

	switch m.state {
	case queryingPackage, fetchingReleases, fetchingChangelog:
		newSpinner, newCmd := m.spinner.Update(msg)
		m.spinner = newSpinner
		cmds = append(cmds, newCmd)
	case choosingRelease:
		newList, newCmd := m.list.Update(msg)
		m.list = newList
		cmds = append(cmds, newCmd)
	}

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	if m.err != nil {
		return m.err.Error()
	}
	switch m.state {
	case queryingPackage:
		return fmt.Sprintf("%s Querying %s's repo...", m.spinner.View(), m.pkgName)
	case fetchingReleases:
		return fmt.Sprintf("%s Querying %s's releases...", m.spinner.View(), m.repo)
	case choosingRelease:
		return docStyle.Render(m.list.View())
	case fetchingChangelog:
		return fmt.Sprintf("%s Fetching changelog...", m.spinner.View())
	}
	return ""
}

func getRepoCmd(pkgName string) tea.Cmd {
	return func() tea.Msg {
		repo := pkgName
		if util.IsNpm(repo) {
			repo = util.GetNpmRepo(pkgName)
		}
		return repoMsg(repo)
	}
}

func getReleasesCmd(repo string) tea.Cmd {
	return func() tea.Msg {
		releases := util.GetReleases(repo)
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

func getChangelogCmd(repo string, tag string) tea.Cmd {
	return func() tea.Msg {
		releaseDetail := util.GetReleaseDetail(repo, tag)
		in := fmt.Sprintf("%s", releaseDetail.Description)
		out, _ := glamour.Render(in, "dark")
		return showChangelogMsg(out)
	}
}
