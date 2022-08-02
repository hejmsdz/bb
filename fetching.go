package main

import (
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/hejmsdz/bb/prs"
)

type PrsLoading struct{}

type PrsLoaded struct {
	prs       []prs.PullRequest
	updatedOn time.Time
}

func StartLoadingPrs(m model) tea.Cmd {
	return func() tea.Msg {
		return PrsLoading{}
	}
}

func LoadPrs(m model) tea.Cmd {
	return func() tea.Msg {
		prs := m.client.GetAllPullRequests()
		m.ticker.Reset(m.updateIntvl)
		return PrsLoaded{prs, time.Now()}
	}
}

func UpdateVisiblePrsList(m *model) {
	prItems := make([]list.Item, 0)
	for _, pr := range m.prs {
		uid := pr.Uid()
		ignoredUntil, ok := m.ignored[uid]
		isIgnored := ok && !pr.UpdatedOn.After(ignoredUntil)
		if !m.showIgnored && isIgnored {
			continue
		}
		var prevPrPtr *prs.PullRequest = nil
		if prevPr, ok := m.prevPrs[uid]; ok {
			prevPrPtr = &prevPr
		}
		prItems = append(prItems, PullRequestItem{pr, prevPrPtr, isIgnored})
	}
	m.list.SetItems(prItems)
	if m.list.Cursor() >= len(prItems) {
		m.list.Select(len(prItems) - 1)
	}
}

func WaitForAutoUpdate(m model) tea.Cmd {
	return func() tea.Msg {
		<-m.ticker.C
		return StartLoadingPrs(m)()
	}
}
