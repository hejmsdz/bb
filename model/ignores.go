package model

import (
	"encoding/json"
	"os"
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

func (m *IgnoresModel) Init() tea.Cmd {
	m.load()
	return nil
}

func (m IgnoresModel) dump() {
	file, _ := json.MarshalIndent(m, "", " ")
	os.WriteFile("./ignores.json", file, 0644)
}

func (m *IgnoresModel) load() bool {
	data, err := os.ReadFile("./ignores.json")
	if err != nil {
		return false
	}
	err = json.Unmarshal(data, &m)
	return err != nil
}

func (m IgnoresModel) IsIgnored(pr prs.PullRequest) bool {
	ignoredUntil, isIgnored := m.IgnoredPrs[pr.Uid()]
	return isIgnored && !pr.UpdatedOn.After(ignoredUntil)
}

func (m IgnoresModel) IsHidden(pr prs.PullRequest) bool {
	return m.IsIgnored(pr) && !m.ShowIgnored
}

func (m IgnoresModel) Persist() tea.Msg {
	go m.dump()
	return nil
}

func (m IgnoresModel) ToggleIgnore(pr prs.PullRequest) tea.Cmd {
	uid := pr.Uid()
	if _, isIgnored := m.IgnoredPrs[uid]; isIgnored {
		delete(m.IgnoredPrs, uid)
	} else {
		m.IgnoredPrs[pr.Uid()] = pr.UpdatedOn
	}
	return tea.Batch(UpdateListView, m.Persist)
}

func (m *IgnoresModel) ToggleShowIgnored() tea.Cmd {
	m.ShowIgnored = !m.ShowIgnored
	return tea.Batch(UpdateListView, m.Persist)
}
