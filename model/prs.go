package model

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hejmsdz/bb/prs"
)

type PrsModel struct {
	client    prs.Client
	Prs       []prs.PullRequest
	updatedOn time.Time
}

func NewPrsModel(client prs.Client) PrsModel {
	return PrsModel{
		client: client,
		Prs:    make([]prs.PullRequest, 0),
	}
}

func (m PrsModel) Init() tea.Cmd {
	return m.StartLoadingPrs
}

type MsgPrsLoading struct{}

type MsgPrsLoaded struct {
	prs       []prs.PullRequest
	updatedOn time.Time
}

func (m PrsModel) StartLoadingPrs() tea.Msg {
	return MsgPrsLoading{}
}

func (m PrsModel) loadPrs() tea.Msg {
	prs := m.client.GetAllPullRequests()
	return MsgPrsLoaded{prs, time.Now()}
}

func (m PrsModel) Update(msg tea.Msg) (PrsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case MsgPrsLoading:
		return m, m.loadPrs

	case MsgPrsLoaded:
		m.Prs = msg.prs
		m.updatedOn = msg.updatedOn
		return m, UpdateListView
	}

	return m, nil
}
