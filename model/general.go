package model

import tea "github.com/charmbracelet/bubbletea"

type MsgUpdateListView struct{}

func UpdateListView() tea.Msg {
	return MsgUpdateListView{}
}
