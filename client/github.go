package client

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

func GetProject() string {
	// GITHUB_REPOSITORY="org/project" => prject
	return strings.Split(os.Getenv("GITHUB_REPOSITORY"), "/")[1]
}

func GetOwner() string {
	return os.Getenv("GITHUB_REPOSITORY_OWNER")
}

func GetReqId() int {
	// When workflows triggered on pull_request, GITHUB_REF_NAME is [pr-number]/merge
	// reqId, _ := strconv.Atoi(strings.Split(os.Getenv("GITHUB_REF_NAME"), "/")[0])

	// When workflows triggered on pull_request_target, GITHUB_REF_NAME is master
	// we set CHANGE_ID: ${{ github.event.pull_request.number }} in workflows env
	reqId, _ := strconv.Atoi(os.Getenv("CHANGE_ID"))

	return reqId
}

func GetEvent() string {
	return os.Getenv("GITHUB_EVENT_NAME")
}

func GetBranch() string {
	if GetEvent() == "push" {
		return os.Getenv("GITHUB_REF_NAME")
	}
	return os.Getenv("GITHUB_BASE_REF")
}

func (cl *Client) getGithubEvent() string {
	// because all called by pull_request_target event,
	// so when pr merged, rename event to pull_request_merged
	pr, _, err := cl.gh.PullRequests.Get(context.Background(), GetOwner(), GetProject(), GetReqId())
	event := GetEvent()
	if err != nil {
		log.Println("get pr merged failed: ", err)
		return event
	}
	if pr.GetMerged() {
		return "pull_request_merged"
	}
	return event
}

func (cl *Client) getCommitSHA(owner, repo string, prID int) string {
	// get commit sha
	// get merged commit sha, while get base sha if not merged
	pr, _, err := cl.gh.PullRequests.Get(context.Background(), owner, repo, prID)
	if err != nil {
		log.Println("get pr reversion failed: ", err)
		return ""
	}
	if pr.GetMerged() {
		return pr.GetMergeCommitSHA()
	}

	return pr.GetBase().GetSHA()
}

// GetPRAuthor get login name and email of the author of the pull request
func (cl *Client) GetPRAuthorAndRef(owner, project string, prID int) (author, email string, ref string, err error) {
	req, _, err := cl.gh.PullRequests.Get(context.Background(), owner, project, prID)
	if err != nil {
		return "", "", "", fmt.Errorf("get pull request: %w", err)
	}
	user, _, err := cl.gh.Users.Get(context.Background(), req.GetUser().GetLogin())
	if err != nil {
		return "", "", "", fmt.Errorf("get user: %w", err)
	}
	return user.GetLogin(), user.GetEmail(), req.Head.GetRef(), nil
}
