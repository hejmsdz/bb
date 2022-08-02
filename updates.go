package main

import (
	"fmt"
	"strings"
)

func FindUpdates(oldPr PullRequest, newPr PullRequest) []string {
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

func FindUpdatesStr(oldPr PullRequest, newPr PullRequest) string {
	return fmt.Sprint(
		"[",
		strings.Join(FindUpdates(oldPr, newPr), ","),
		"]",
	)
}
