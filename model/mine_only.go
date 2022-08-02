package model

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/hejmsdz/bb/prs"
)

type MineOnlyModel struct {
	ShowMineOnly bool
}

func NewMineOnlyModel() MineOnlyModel {
	return MineOnlyModel{
		ShowMineOnly: false,
	}
}

func (m MineOnlyModel) IsHidden(pr prs.PullRequest) bool {
	return m.ShowMineOnly && !pr.IsMine
}

func (m *MineOnlyModel) ToggleShowMineOnly() tea.Cmd {
	m.ShowMineOnly = !m.ShowMineOnly
	return UpdateListView
}
