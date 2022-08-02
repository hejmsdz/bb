package prs

type Client interface {
	GetAllPullRequests() []PullRequest
}
