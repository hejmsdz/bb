package model

import (
	tea "github.com/charmbracelet/bubbletea"
)

type AsyncModel struct {
	channel chan tea.Cmd
}

type MsgCmdReceived struct {
	cmd tea.Cmd
}

func NewAsyncModel() AsyncModel {
	return AsyncModel{
		channel: make(chan tea.Cmd),
	}
}

func (m AsyncModel) Init() tea.Cmd {
	return m.receive
}

func (m AsyncModel) receive() tea.Msg {
	cmd := <-m.channel
	return MsgCmdReceived{cmd}
}

func (m AsyncModel) GetChannel() chan tea.Cmd {
	return m.channel
}

func (m AsyncModel) Update(msg tea.Msg) (AsyncModel, tea.Cmd) {
	switch msg := msg.(type) {
	case MsgCmdReceived:
		return m, tea.Batch(msg.cmd, m.receive)
	}
	return m, nil
}
