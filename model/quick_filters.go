package model

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/hejmsdz/bb/prs"
)

type QuickFiltersModel struct {
	ShowMineOnly bool
}

func NewQuickFiltersModel() QuickFiltersModel {
	return QuickFiltersModel{
		ShowMineOnly: false,
	}
}

func (m QuickFiltersModel) IsHidden(pr prs.PullRequest) bool {
	return m.ShowMineOnly && !pr.IsMine
}

func (m *QuickFiltersModel) ToggleShowMineOnly() tea.Cmd {
	m.ShowMineOnly = !m.ShowMineOnly
	return UpdateListView
}
