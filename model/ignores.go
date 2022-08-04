package model

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hejmsdz/bb/prs"
)

type IgnoresModel struct {
	ShowIgnored bool
	IgnoredPrs  map[prs.Uid]time.Time
}

func NewIgnoresModel() IgnoresModel {
	return IgnoresModel{
		ShowIgnored: false,
		IgnoredPrs:  make(map[prs.Uid]time.Time),
	}
}

func (m IgnoresModel) IsIgnored(pr prs.PullRequest) bool {
	ignoredUntil, isIgnored := m.IgnoredPrs[pr.Uid()]
	return isIgnored && !pr.UpdatedOn.After(ignoredUntil)
}

func (m IgnoresModel) IsHidden(pr prs.PullRequest) bool {
	return m.IsIgnored(pr) && !m.ShowIgnored
}

func (m IgnoresModel) ToggleIgnore(pr prs.PullRequest) tea.Cmd {
	uid := pr.Uid()
	if _, isIgnored := m.IgnoredPrs[uid]; isIgnored {
		delete(m.IgnoredPrs, uid)
	} else {
		m.IgnoredPrs[pr.Uid()] = pr.UpdatedOn
	}
	return tea.Batch(UpdateListView)
}

func (m *IgnoresModel) ToggleShowIgnored() tea.Cmd {
	m.ShowIgnored = !m.ShowIgnored
	return tea.Batch(UpdateListView)
}
