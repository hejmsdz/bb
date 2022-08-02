package prs

import (
	"fmt"
	"time"
)

type Review int

const (
	Approved Review = 1
	RequestedChanges
)

type PullRequest struct {
	Id                    string
	Repo                  string
	Title                 string
	Author                string
	LastCommit            string
	IsMine                bool
	AmIParticipating      bool
	UpdatedOn             time.Time
	CommentsCount         int
	ReviewersCount        int
	ApprovedCount         int
	RequestedChangesCount int
	MyReview              Review
	Url                   string
}

func (pr PullRequest) Uid() string {
	return fmt.Sprint(pr.Repo, "/", pr.Id)
}
