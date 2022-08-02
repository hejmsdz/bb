package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"
)

type BitbucketClient struct {
	Config     AccountConfig
	apiUrl     string
	httpClient *http.Client
}

type BbPullRequestsResponse struct {
	Values []BbPullRequest `json:"values"`
}

type BbUser struct {
	DisplayName string `json:"display_name"`
	AccountId   string `json:"account_id"`
}

type BbLink struct {
	Href string `json:"href"`
}

type BbLinks struct {
	Html BbLink `json:"html"`
}

type BbCommit struct {
	Hash string `json:"hash"`
}

type BbSource struct {
	Commit BbCommit `json:"commit"`
}

type BbParticipant struct {
	User  BbUser `json:"user"`
	Role  string `json:"role"`
	State string `json:"state"`
}

type BbPullRequest struct {
	Id           int             `json:"id"`
	Title        string          `json:"title"`
	UpdatedOn    string          `json:"updated_on"`
	CommentCount int             `json:"comment_count"`
	Author       BbUser          `json:"author"`
	Source       BbSource        `json:"source"`
	Links        BbLinks         `json:"links"`
	Participants []BbParticipant `json:"participants"`
}

func CreateBitbucketClient(config AccountConfig) BitbucketClient {
	return BitbucketClient{
		config,
		"https://api.bitbucket.org/2.0/",
		&http.Client{},
	}
}

func (c BitbucketClient) get(path string) (*http.Response, error) {
	req, err := http.NewRequest("GET", c.apiUrl+path, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(c.Config.Username, c.Config.Password)
	return c.httpClient.Do(req)
}

func processReviewers(participants []BbParticipant, pr *PullRequest, myUserId string) {
	for _, part := range participants {
		if part.Role != "REVIEWER" {
			continue
		}

		pr.ReviewersCount++

		if part.User.AccountId == myUserId {
			pr.AmIParticipating = true
			if part.State == "approved" {
				pr.MyReview = Approved
			} else if part.State == "changes_requested" {
				pr.MyReview = RequestedChanges
			}
		}

		if part.State == "approved" {
			pr.ApprovedCount++
		} else if part.State == "changes_requested" {
			pr.RequestedChangesCount++
		}
	}

}

var prFieldsStr = strings.Join([]string{
	"values.id",
	"values.title",
	"values.updated_on",
	"values.comment_count",
	"values.author.display_name",
	"values.author.account_id",
	"values.source.commit.hash",
	"values.links.html.href",
	"values.participants.role",
	"values.participants.state",
	"values.participants.user.account_id",
}, ",")

func (c BitbucketClient) GetPullRequests(repo string) []PullRequest {
	resp, _ := c.get(fmt.Sprintf("repositories/%s/pullrequests?state=OPEN&pagelen=50&fields=%s", repo, prFieldsStr))
	var bbPrs *BbPullRequestsResponse
	json.NewDecoder(resp.Body).Decode(&bbPrs)
	prs := make([]PullRequest, 0)

	for _, bbPr := range bbPrs.Values {
		pr := PullRequest{
			Id:            fmt.Sprintf("%d", bbPr.Id),
			Repo:          repo,
			Title:         bbPr.Title,
			Author:        bbPr.Author.DisplayName,
			LastCommit:    bbPr.Source.Commit.Hash,
			CommentsCount: bbPr.CommentCount,
			Url:           bbPr.Links.Html.Href,
			IsMine:        bbPr.Author.AccountId == c.Config.UserId,
		}

		pr.UpdatedOn, _ = time.Parse("2006-01-02T15:04:05.000000-07:00", bbPr.UpdatedOn)
		processReviewers(bbPr.Participants, &pr, c.Config.UserId)

		if pr.IsMine || pr.AmIParticipating {
			prs = append(prs, pr)
		}
	}
	return prs
}

func (c BitbucketClient) GetAllPullRequests() []PullRequest {
	allPrs := make([]PullRequest, 0)
	for _, repo := range c.Config.Repositories {
		repoPrs := c.GetPullRequests(repo)
		allPrs = append(allPrs, repoPrs...)
	}
	sort.Slice(allPrs, func(i, j int) bool {
		return allPrs[i].UpdatedOn.After(allPrs[j].UpdatedOn)
	})
	return allPrs
}
