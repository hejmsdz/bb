package model

import (
	"github.com/charmbracelet/bubbles/key"
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

	d.FullHelpFunc = func() [][]key.Binding {
		return [][]key.Binding{{
			key.NewBinding(
				key.WithKeys("enter"),
				key.WithHelp("enter", "open in web browser"),
			),
			key.NewBinding(
				key.WithKeys("d"),
				key.WithHelp("d", "dismiss bell"),
			),
			key.NewBinding(
				key.WithKeys("i"),
				key.WithHelp("i", "ignore until next update"),
			),
			key.NewBinding(
				key.WithKeys("u"),
				key.WithHelp("u", "copy url"),
			),
			key.NewBinding(
				key.WithKeys("b"),
				key.WithHelp("b", "copy branch"),
			),
			key.NewBinding(
				key.WithKeys("c"),
				key.WithHelp("c", "checkout locally"),
			),
			key.NewBinding(
				key.WithKeys("P"),
				key.WithHelp("P", "pull target branch"),
			),
		}, {
			key.NewBinding(
				key.WithKeys("r"),
				key.WithHelp("r", "refresh"),
			),
			key.NewBinding(
				key.WithKeys("."),
				key.WithHelp(".", "show ignored"),
			),
			key.NewBinding(
				key.WithKeys("m"),
				key.WithHelp("m", "show mine only"),
			),
		}}
	}

	return d
}
