package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/hejmsdz/bb/prs"
)

type PullRequestItem struct {
	Pr          prs.PullRequest
	WhatChanged []string
	IsIgnored   bool
}

func (i PullRequestItem) Title() string {
	if i.IsIgnored {
		return ignoredStyle.Render(i.Pr.Title)
	}
	if len(i.WhatChanged) > 0 {
		return fmt.Sprint("🔔 ", i.Pr.Title, " [", updatesStyle.Render(strings.Join(i.WhatChanged, ", ")), "]")
	}
	return i.Pr.Title
}
func (i PullRequestItem) FilterValue() string { return fmt.Sprint(i.Pr.Title, i.Pr.Author) }
func (i PullRequestItem) Description() string {
	timeAgo := TimeAgo(i.Pr.UpdatedOn)
	var myReviewEmoji = ""
	if i.Pr.MyReview == prs.Approved {
		myReviewEmoji = " / ✅"
	} else if i.Pr.MyReview == prs.RequestedChanges {
		myReviewEmoji = " / 👎"
	}
	reviewSummary := fmt.Sprintf("%s / %s%s",
		approvesStyle.Render(fmt.Sprint(i.Pr.ApprovedCount)),
		requestedChangesStyle.Render(fmt.Sprint(i.Pr.RequestedChangesCount)),
		myReviewEmoji,
	)
	return fmt.Sprintf("%s | %s | %d 💬 | %s", i.Pr.Author, timeAgo, i.Pr.CommentsCount, reviewSummary)
}

func TimeAgo(t time.Time) string {
	d := -time.Until(t)
	if d < time.Minute {
		return "just now"
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	}
	if d < 30*24*time.Hour {
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	}
	if d < 365*24*time.Hour {
		return fmt.Sprintf("%dmo ago", int(d.Hours()/(24*30)))
	}
	return fmt.Sprintf("%dy ago", int(d.Hours()/(24*365)))
}
