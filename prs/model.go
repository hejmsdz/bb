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
	Branch                string
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

type Uid = string

func (pr PullRequest) Uid() Uid {
	return fmt.Sprint(pr.Repo, "/", pr.Id)
}
