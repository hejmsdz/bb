package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hejmsdz/bb/model"
	"github.com/hejmsdz/bb/prs"
	"github.com/pkg/browser"
)

const listHeight = 14

var (
	docStyle   = lipgloss.NewStyle().Margin(1, 2)
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#25A065")).
			Padding(0, 1)
	paginationStyle       = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle             = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	ignoredStyle          = lipgloss.NewStyle().Faint(true).Strikethrough(true)
	approvesStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	requestedChangesStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	updatesStyle          = lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Bold(true)
	infoToastStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	successToastStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	errorToastStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
)

type rootModel struct {
	Ignores      model.IgnoresModel
	WhatChanged  model.WhatChangedModel
	QuickFilters model.QuickFiltersModel
	prs          model.PrsModel
	list         list.Model
	autoUpdate   model.AutoUpdateModel
	async        model.AsyncModel
	localRepos   map[string]string
	quitting     bool
}

func (m rootModel) Init() tea.Cmd {
	return tea.Batch(
		m.prs.Init(),
		m.autoUpdate.Init(),
		m.async.Init(),
	)
}

func (m rootModel) dump() tea.Msg {
	file, _ := json.MarshalIndent(m, "", " ")
	os.WriteFile(stateFilePath, file, 0600)
	return nil
}

func (m *rootModel) load() bool {
	data, err := os.ReadFile(stateFilePath)
	if err != nil {
		return false
	}
	err = json.Unmarshal(data, &m)
	return err != nil
}

func NewInfoToast(text string) tea.Cmd {
	return func() tea.Msg {
		return model.MsgShowToast{Text: text, Style: infoToastStyle}
	}
}

func NewToast(text string, isOk bool) tea.Cmd {
	return func() tea.Msg {
		style := successToastStyle
		if !isOk {
			style = errorToastStyle
		}
		return model.MsgShowToast{Text: text, Style: style}
	}
}

func NewErrorToast(text string) tea.Cmd {
	return func() tea.Msg {
		return model.MsgShowToast{Text: text, Style: errorToastStyle}
	}
}

func OpenBrowser(url string) tea.Cmd {
	return func() tea.Msg {
		browser.OpenURL(url)
		return nil
	}
}

func CopyToClipboard(str string, m rootModel) tea.Cmd {
	clipboard.WriteAll(str)
	return NewInfoToast("Copied '" + str + "'")
}

func UpdateListView(m *rootModel) {
	prItems := make([]list.Item, 0)
	for _, pr := range m.prs.Prs {
		if m.Ignores.IsHidden(pr) || m.QuickFilters.IsHidden(pr) {
			continue
		}

		prItems = append(prItems, PullRequestItem{
			pr,
			m.WhatChanged.WhatChanged(pr),
			m.Ignores.IsIgnored(pr),
		})
	}
	m.list.SetItems(prItems)
	if m.list.Cursor() >= len(prItems) {
		m.list.Select(len(prItems) - 1)
	}
	if m.list.Cursor() < 0 {
		m.list.Select(0)
	}
}

func (m rootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
		return m, nil

	case model.MsgUpdateListView:
		UpdateListView(&m)
		return m, nil

	case tea.KeyMsg:
		if m.list.FilterState() == list.Filtering {
			break
		}

		switch keypress := msg.String(); keypress {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Batch(m.dump, tea.Quit)

		case "r":
			return m, m.prs.StartLoadingPrs

		case ".":
			cmd := m.Ignores.ToggleShowIgnored()
			return m, cmd

		case "m":
			cmd := m.QuickFilters.ToggleShowMineOnly()
			if m.QuickFilters.ShowMineOnly {
				m.list.Title = "My pull requests"
			} else {
				m.list.Title = "Pull requests"
			}
			return m, cmd
		}

		sel, ok := m.list.SelectedItem().(PullRequestItem)
		if !ok {
			return m, nil
		}

		switch keypress := msg.String(); keypress {
		case "i":
			cmd := m.Ignores.ToggleIgnore(sel.Pr)
			return m, cmd

		case "d":
			cmd := m.WhatChanged.DismissChanges(sel.Pr)
			return m, cmd

		case "u":
			return m, CopyToClipboard(sel.Pr.Url, m)

		case "b":
			return m, CopyToClipboard(sel.Pr.Branch, m)

		case "c":
			return m, Checkout(sel.Pr, m)

		case "P":
			return m, PullOrigin(sel.Pr, m)

		case "enter":
			cmd := m.WhatChanged.DismissChanges(sel.Pr)
			return m, tea.Batch(OpenBrowser(sel.Pr.Url), cmd)
		}
	}

	var listCmd, prsCmd, autoUpdateCmd, whatChangedCmd, toastStreamCmd tea.Cmd
	m.list, listCmd = m.list.Update(msg)
	m.prs, prsCmd = m.prs.Update(msg)
	m.autoUpdate, autoUpdateCmd = m.autoUpdate.Update(msg)
	m.WhatChanged, whatChangedCmd = m.WhatChanged.Update(msg)
	m.async, toastStreamCmd = m.async.Update(msg)

	return m, tea.Batch(listCmd, prsCmd, autoUpdateCmd, whatChangedCmd, toastStreamCmd)
}

func (m rootModel) View() string {
	return m.list.View()
}

func main() {
	config, ok := ReadConfig()
	if !ok {
		CreateSampleConfig()
		fmt.Println("Welcome to " + infoToastStyle.Render("bb") + ", a command-line pull requests dashboard!")
		fmt.Println()
		fmt.Println("To get started, please open the following file:")
		fmt.Println(infoToastStyle.Render(configFilePath))
		fmt.Println("and complete your configuration.")
		os.Exit(1)
	}
	c, ok := prs.CreateBitbucketClient(config.Bitbucket)
	if !ok {
		fmt.Println(errorToastStyle.Render("Could not connect to Bitbucket API."))
		fmt.Println("Make sure that your credentials configured in the file:")
		fmt.Println(infoToastStyle.Render(configFilePath))
		fmt.Println("are valid and have the permissions " + successToastStyle.Render("account") + " and " + successToastStyle.Render("pullrequest") + ".")

		os.Exit(1)
	}
	interval := time.Duration(config.UpdateIntervalMinutes) * time.Minute

	const defaultWidth = 20

	l := list.New(make([]list.Item, 0), model.NewItemDelegate(), defaultWidth, listHeight)
	l.Title = "Pull requests"
	l.SetShowStatusBar(false)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	m := rootModel{
		list:         l,
		prs:          model.NewPrsModel(c),
		Ignores:      model.NewIgnoresModel(),
		autoUpdate:   model.NewAutoUpdateModel(interval),
		WhatChanged:  model.NewWhatChangedModel(),
		QuickFilters: model.NewQuickFiltersModel(),
		async:        model.NewAsyncModel(),
		localRepos:   config.LocalRepositoryPaths,
	}

	m.load()
	defer m.dump()

	if err := tea.NewProgram(m, tea.WithAltScreen()).Start(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}

}
