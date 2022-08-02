package prs

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"
)

type BitbucketClient struct {
	config     AccountConfig
	apiUrl     string
	httpClient *http.Client
}

type bbPullRequestsResponse struct {
	Values []bbPullRequest `json:"values"`
}

type bbUser struct {
	DisplayName string `json:"display_name"`
	AccountId   string `json:"account_id"`
}

type bbLink struct {
	Href string `json:"href"`
}

type bbLinks struct {
	Html bbLink `json:"html"`
}

type bbCommit struct {
	Hash string `json:"hash"`
}

type bbSource struct {
	Commit bbCommit `json:"commit"`
}

type bbParticipant struct {
	User  bbUser `json:"user"`
	Role  string `json:"role"`
	State string `json:"state"`
}

type bbPullRequest struct {
	Id           int             `json:"id"`
	Title        string          `json:"title"`
	UpdatedOn    string          `json:"updated_on"`
	CommentCount int             `json:"comment_count"`
	Author       bbUser          `json:"author"`
	Source       bbSource        `json:"source"`
	Links        bbLinks         `json:"links"`
	Participants []bbParticipant `json:"participants"`
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
	req.SetBasicAuth(c.config.Username, c.config.Password)
	return c.httpClient.Do(req)
}

func processReviewers(participants []bbParticipant, pr *PullRequest, myUserId string) {
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

func (c BitbucketClient) getPullRequests(repo string) []PullRequest {
	resp, _ := c.get(fmt.Sprintf("repositories/%s/pullrequests?state=OPEN&pagelen=50&fields=%s", repo, prFieldsStr))
	var bbPrs *bbPullRequestsResponse
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
			IsMine:        bbPr.Author.AccountId == c.config.UserId,
		}

		pr.UpdatedOn, _ = time.Parse("2006-01-02T15:04:05.000000-07:00", bbPr.UpdatedOn)
		processReviewers(bbPr.Participants, &pr, c.config.UserId)

		if pr.IsMine || pr.AmIParticipating {
			prs = append(prs, pr)
		}
	}
	return prs
}

func (c BitbucketClient) GetAllPullRequests() []PullRequest {
	allPrs := make([]PullRequest, 0)
	for _, repo := range c.config.Repositories {
		repoPrs := c.getPullRequests(repo)
		allPrs = append(allPrs, repoPrs...)
	}
	sort.Slice(allPrs, func(i, j int) bool {
		return allPrs[i].UpdatedOn.After(allPrs[j].UpdatedOn)
	})
	return allPrs
}
