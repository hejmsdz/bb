package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
)

type PullRequestItem struct {
	Pr        prs.PullRequest
	PrevState *prs.PullRequest
	IsIgnored bool
}

func (i PullRequestItem) Title() string {
	if i.IsIgnored {
		return ignoredStyle.Render(i.Pr.Title)
	}
	if i.PrevState != nil && i.Pr.UpdatedOn.After(i.PrevState.UpdatedOn) {
		return fmt.Sprint("🔔 ", i.Pr.Title, " ", updatesStyle.Render(FindUpdatesStr(*i.PrevState, i.Pr)))
	}
	return i.Pr.Title
}
func (i PullRequestItem) FilterValue() string { return fmt.Sprint(i.Pr.Title, i.Pr.Author) }
func (i PullRequestItem) Description() string {
	timeAgo := TimeAgo(i.Pr.UpdatedOn)
	var myReviewEmoji = ""
	if i.Pr.MyReview == prs.Approved {
		myReviewEmoji = " / ✅"
	} else if i.Pr.MyReview == prs.RequestedChanges {
		myReviewEmoji = " / 👎"
	}
	reviewSummary := fmt.Sprintf("%s / %s%s",
		approvesStyle.Render(fmt.Sprint(i.Pr.ApprovedCount)),
		requestedChangesStyle.Render(fmt.Sprint(i.Pr.RequestedChangesCount)),
		myReviewEmoji,
	)
	return fmt.Sprintf("%s | %s | %d 💬 | %s", i.Pr.Author, timeAgo, i.Pr.CommentsCount, reviewSummary)
}

type ignoredMap map[string]time.Time

type model struct {
	client      prs.Client
	list        list.Model
	prs         []prs.PullRequest
	prevPrs     map[string]prs.PullRequest
	ignored     ignoredMap
	quitting    bool
	updatedOn   time.Time
	showIgnored bool
	ticker      time.Ticker
	updateIntvl time.Duration
}

func OpenBrowser(url string) tea.Cmd {
	return func() tea.Msg {
		browser.OpenURL(url)
		return nil
	}
}

func SaveIgnored(ignored ignoredMap) {
	file, _ := json.MarshalIndent(ignored, "", " ")
	os.WriteFile("./ignored.json", file, 0644)
}

func LoadIgnored() ignoredMap {
	var ignored ignoredMap = make(ignoredMap)
	ignoredJson, err := os.ReadFile("./ignored.json")
	if err != nil {
		return ignored
	}
	json.Unmarshal(ignoredJson, &ignored)
	return ignored
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		StartLoadingPrs(m),
		WaitForAutoUpdate(m),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
		return m, nil

	case PrsLoading:
		cmd := m.list.StartSpinner()
		return m, tea.Batch(cmd, LoadPrs(m))

	case PrsLoaded:
		m.list.StopSpinner()
		for _, oldPr := range m.prs {
			_, isCached := m.prevPrs[oldPr.Uid()]
			if !isCached {
				m.prevPrs[oldPr.Uid()] = oldPr
			}
		}
		m.prs = msg.prs
		m.updatedOn = msg.updatedOn
		UpdateVisiblePrsList(&m)
		return m, WaitForAutoUpdate(m)

	case tea.KeyMsg:
		if m.list.FilterState() == list.Filtering {
			break
		}

		switch keypress := msg.String(); keypress {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "r":
			return m, StartLoadingPrs(m)

		case "i":
			i, ok := m.list.SelectedItem().(PullRequestItem)
			if ok {
				uid := i.Pr.Uid()
				if i.IsIgnored {
					delete(m.ignored, uid)
				} else {
					m.ignored[i.Pr.Uid()] = i.Pr.UpdatedOn
				}
			}
			go SaveIgnored(m.ignored)
			UpdateVisiblePrsList(&m)
			return m, nil

		case "d":
			i, ok := m.list.SelectedItem().(PullRequestItem)
			if ok {
				uid := i.Pr.Uid()
				m.prevPrs[uid] = i.Pr
			}
			UpdateVisiblePrsList(&m)
			return m, nil
		case ".":
			m.showIgnored = !m.showIgnored
			UpdateVisiblePrsList(&m)
			return m, nil

		case "enter":
			i, ok := m.list.SelectedItem().(PullRequestItem)
			if ok {
				return m, OpenBrowser(i.Pr.Url)
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return m.list.View()
}

func main() {
	config := ReadConfig()
	c := prs.CreateBitbucketClient(config.Bitbucket)

	const defaultWidth = 20

	l := list.New(make([]list.Item, 0), list.NewDefaultDelegate(), defaultWidth, listHeight)
	l.Title = "Pull requests"
	l.SetShowStatusBar(false)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	intvl := time.Duration(config.UpdateIntervalMinutes) * time.Minute

	m := model{
		client:      c,
		list:        l,
		ignored:     LoadIgnored(),
		prevPrs:     make(map[string]prs.PullRequest),
		ticker:      *time.NewTicker(intvl),
		updateIntvl: intvl,
	}

	if err := tea.NewProgram(m, tea.WithAltScreen()).Start(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
