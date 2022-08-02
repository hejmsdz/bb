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
	return m.WaitForAutoUpdate
}

func (m AutoUpdateModel) WaitForAutoUpdate() tea.Msg {
	<-m.ticker.C
	return MsgPrsLoading{}
}

func (m AutoUpdateModel) ScheduleAutoUpdate() tea.Msg {
	m.ticker.Reset(m.interval)
	return nil
}

func (m AutoUpdateModel) Update(msg tea.Msg) (AutoUpdateModel, tea.Cmd) {
	switch msg.(type) {
	case MsgPrsLoaded:
		return m, tea.Batch(m.ScheduleAutoUpdate, m.WaitForAutoUpdate)
	}
	return m, nil
}
