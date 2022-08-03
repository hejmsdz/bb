package model

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/hejmsdz/bb/prs"
)

type WhatChangedModel struct {
	prevPrs map[prs.Uid]prs.PullRequest
}

func NewWhatChangedModel() WhatChangedModel {
	return WhatChangedModel{
		prevPrs: make(map[prs.Uid]prs.PullRequest),
	}
}

func (m WhatChangedModel) WhatChanged(pr prs.PullRequest) []string {
	prevPr, exists := m.prevPrs[pr.Uid()]
	if !exists {
		return []string{}
	}
	return findUpdates(prevPr, pr)
}

func (m WhatChangedModel) DismissChanges(pr prs.PullRequest) tea.Cmd {
	uid := pr.Uid()
	m.prevPrs[uid] = pr
	return UpdateListView
}

func (m WhatChangedModel) Update(msg tea.Msg) (WhatChangedModel, tea.Cmd) {
	switch msg := msg.(type) {
	case MsgPrsLoaded:
		for _, oldPr := range msg.prs {
			_, isCached := m.prevPrs[oldPr.Uid()]
			if !isCached {
				m.prevPrs[oldPr.Uid()] = oldPr
			}
		}
	}
	return m, nil
}

func findUpdates(oldPr prs.PullRequest, newPr prs.PullRequest) []string {
	updates := make([]string, 0)

	if newPr.LastCommit != oldPr.LastCommit {
		updates = append(updates, "commited")
	}

	if newPr.CommentsCount != oldPr.CommentsCount {
		updates = append(updates, "commented")
	}

	if newPr.ApprovedCount != oldPr.ApprovedCount {
		updates = append(updates, "approved")
	}

	if newPr.RequestedChangesCount != oldPr.RequestedChangesCount {
		updates = append(updates, "changesRequested")
	}

	if newPr.MyReview != oldPr.MyReview {
		updates = append(updates, "youReviewed")
	}

	return updates
}
