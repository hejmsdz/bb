package model

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type MsgShowToast struct {
	Text  string
	Style lipgloss.Style
}

func NewItemDelegate() list.DefaultDelegate {
	d := list.NewDefaultDelegate()

	d.UpdateFunc = func(msg tea.Msg, m *list.Model) tea.Cmd {
		switch msg := msg.(type) {
		case MsgPrsLoading:
			return m.StartSpinner()

		case MsgPrsLoaded:
			m.StopSpinner()
			return nil

		case MsgShowToast:
			return m.NewStatusMessage(msg.Style.Render(msg.Text))
		}

		return nil
	}

	return d
}
