package model

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

func NewItemDelegate() list.DefaultDelegate {
	d := list.NewDefaultDelegate()

	d.UpdateFunc = func(msg tea.Msg, m *list.Model) tea.Cmd {
		switch msg.(type) {
		case MsgPrsLoading:
			return m.StartSpinner()

		case MsgPrsLoaded:
			m.StopSpinner()
			return nil
		}

		return nil
	}

	return d
}
