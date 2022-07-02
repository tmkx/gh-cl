package ui

import (
	"fmt"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/tmkx/gh-cl/util"
	"strings"
)

const (
	queryingPackage = iota
	fetchingReleases
	choosingRelease
	fetchingChangelog
	showChangelog
)

var (
	docStyle = lipgloss.NewStyle().Margin(1, 2)

	titleStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Right = "├"
		return lipgloss.NewStyle().BorderStyle(b).Padding(0, 1)
	}()

	infoStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Left = "┤"
		return titleStyle.Copy().BorderStyle(b)
	}()
)

type model struct {
	state     int
	pkgName   string
	repo      string
	tag       string
	releases  []list.Item
	changelog *util.ReleaseDetail
	spinner   spinner.Model
	list      list.Model
	viewport  viewport.Model
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

type errorMsg error
type repoMsg string
type releasesMsg []list.Item
type chooseReleaseMsg string
type showChangelogMsg *util.ReleaseDetail

func InitModel(pkgName string) model {
	s := spinner.New()
	s.Spinner = spinner.MiniDot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	l := list.New(make([]list.Item, 0), list.NewDefaultDelegate(), 0, 0)
	l.Title = "Please select a release"

	v := viewport.New(0, 0)

	return model{spinner: s, list: l, viewport: v, pkgName: pkgName, state: queryingPackage}
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
			if m.state == showChangelog && msg.String() != "ctrl+c" {
				m.state = choosingRelease
				return m, nil
			}
			m.quitting = true
			return m, tea.Quit
		case "enter":
			if m.state == choosingRelease {
				i, ok := m.list.SelectedItem().(item)
				if ok {
					m.tag = i.title
					cmds = append(cmds, func() tea.Msg {
						return chooseReleaseMsg(i.title)
					})
				}
			}
		}
	case tea.WindowSizeMsg:
		// list
		w, h := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-w, msg.Height-h)
		// viewport
		headerHeight := lipgloss.Height(m.headerView())
		footerHeight := lipgloss.Height(m.footerView())
		verticalMarginHeight := headerHeight + footerHeight
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - verticalMarginHeight
		m.viewport.YPosition = headerHeight
	case errorMsg:
		m.err = msg
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
		m.changelog = msg
		in := fmt.Sprintf("%s\n\n**Open in browser**: %s", m.changelog.Description, m.changelog.Url)
		r, _ := glamour.NewTermRenderer(
			glamour.WithStylePath("dracula"),
			glamour.WithWordWrap(m.viewport.Width),
		)
		out, _ := r.Render(in)
		m.viewport.SetContent(out)
		cmds = append(cmds, tea.ClearScrollArea)
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
	case showChangelog:
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
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
	case showChangelog:
		return fmt.Sprintf("%s\n%s\n%s", m.headerView(), m.viewport.View(), m.footerView())
	}
	return ""
}

func (m model) headerView() string {
	title := titleStyle.Render(fmt.Sprintf("%s(%s) %s", m.pkgName, m.repo, m.tag))
	// https://github.com/charmbracelet/lipgloss/issues/40
	line := strings.Repeat("─", util.Max(0, m.viewport.Width-lipgloss.Width(title)))
	return lipgloss.JoinHorizontal(lipgloss.Center, title, line)
}

func (m model) footerView() string {
	info := infoStyle.Render(fmt.Sprintf("%3.f%%", m.viewport.ScrollPercent()*100))
	line := strings.Repeat("─", util.Max(0, m.viewport.Width-lipgloss.Width(info)))
	return lipgloss.JoinHorizontal(lipgloss.Center, line, info)
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
		releases, err := util.GetReleases(repo)
		if err != nil {
			return errorMsg(err)
		}
		var items []list.Item
		for i := 0; i < len(releases); i++ {
			release := releases[i]
			var extra []string
			if release.IsLatest {
				extra = append(extra, "latest")
			}
			if release.IsPrerelease {
				extra = append(extra, "prerelease")
			}
			items = append(items, item{title: release.TagName, desc: fmt.Sprintf("%v %v", release.PublishedAt, strings.Join(extra, ", "))})
		}
		return releasesMsg(items)
	}
}

func getChangelogCmd(repo string, tag string) tea.Cmd {
	return func() tea.Msg {
		releaseDetail, err := util.GetReleaseDetail(repo, tag)
		if err != nil {
			return errorMsg(err)
		}
		return showChangelogMsg(releaseDetail)
	}
}
