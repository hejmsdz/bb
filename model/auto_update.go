package model

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type AutoUpdateModel struct {
	ticker   *time.Ticker
	interval time.Duration
}

func NewAutoUpdateModel(interval time.Duration) AutoUpdateModel {
	return AutoUpdateModel{
		ticker:   time.NewTicker(interval),
		interval: interval,
	}
}

func (m AutoUpdateModel) Init() tea.Cmd {
	return m.waitForAutoUpdate
}

func (m AutoUpdateModel) waitForAutoUpdate() tea.Msg {
	<-m.ticker.C
	return MsgPrsLoading{}
}

func (m AutoUpdateModel) scheduleAutoUpdate() tea.Msg {
	m.ticker.Reset(m.interval)
	return nil
}

func (m AutoUpdateModel) Update(msg tea.Msg) (AutoUpdateModel, tea.Cmd) {
	switch msg.(type) {
	case MsgPrsLoaded:
		return m, tea.Batch(m.scheduleAutoUpdate, m.waitForAutoUpdate)
	}
	return m, nil
}
