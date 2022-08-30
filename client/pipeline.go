package client

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-resty/resty/v2"
)

// topic 规范：pr branch 以TOPIC_PREFIX 开始时会被加入到topic仓库中
const (
	TOPIC_PREFIX = "topic-"
	TOPIC_SIZE   = len(TOPIC_PREFIX)
)

// 创建打包构建任务,返回值为 id
//  /api/job/sync
//  /api/job/build
//  /api/job/sync
//  /api/job/abicheck
//  /api/job/archlinux

type JobTriggerJenkins struct {
	ID int `json:"ID"`
}

// build deb params
type Build struct {
	Branch        string `json:"branch"`
	CommentAuthor string `json:"comment_author"`
	GroupName     string `json:"group_name"`
	Project       string `json:"project"`
	RequestEvent  string `json:"request_event"`
	RequestId     int    `json:"request_id"`
	Sha           string `json:"sha"`
	IsPush        bool   `json:"is_push"`
	AuthorEmail   string `json:"author_email"`
	Topic         string `json:"topic"`
	ReversionID   string `json:"reversionID"`
}

// triggerSync
func (cl *Client) PostApiJobSync() {
	client := resty.New()
	client.SetRetryCount(3).SetRetryWaitTime(5 * time.Second).SetRetryMaxWaitTime(20 * time.Second)
	resp, err := client.R().
		SetBody(Build{
			Project:   GetProject(),
			GroupName: GetOwner(),
		}).
		SetHeader("Accept", "application/json").
		SetHeader("X-token", cl.token).
		Post(cl.host + "/api/job/sync")

	if err != nil {
		log.Fatal(err)
	}

	if resp.StatusCode() != 200 {
		log.Fatal("trigger build fail, StatusCode not 200")
	}
	var jobSync JobTriggerJenkins
	err = json.Unmarshal([]byte(resp.Body()), &jobSync)
	if err != nil {
		log.Fatal(err)
	}
	cl.id = jobSync.ID
}

// triggerAbicheck
func (cl *Client) PostApiJobAbicheck() {
	client := resty.New()
	client.SetRetryCount(3).SetRetryWaitTime(5 * time.Second).SetRetryMaxWaitTime(20 * time.Second)

	author, email, _, err := cl.GetPRAuthorAndRef(GetOwner(), GetProject(), GetReqId())
	if err != nil {
		// Ignore failure
		log.Println("get pr author fail: ", err)
	}
	resp, err := client.R().
		SetBody(Build{
			Branch:        GetBranch(),
			GroupName:     GetOwner(),
			Project:       GetProject(),
			RequestEvent:  GetEvent(),
			RequestId:     GetReqId(),
			CommentAuthor: author,
			AuthorEmail:   email,
		}).
		SetHeader("Accept", "application/json").
		SetHeader("X-token", cl.token).
		Post(cl.host + "/api/job/abicheck")

	if err != nil {
		log.Fatal(err)
	}

	if resp.StatusCode() != 200 {
		log.Fatal("trigger build fail, StatusCode not 200")
	}

	var jobAbicheck JobTriggerJenkins

	json.Unmarshal([]byte(resp.Body()), &jobAbicheck)

	cl.id = jobAbicheck.ID
}

// triggerArchlinux
func (cl *Client) PostApiJobArchlinux() {
	client := resty.New()
	client.SetRetryCount(3).SetRetryWaitTime(5 * time.Second).SetRetryMaxWaitTime(20 * time.Second)
	resp, err := client.R().
		SetBody(Build{
			Project: GetProject(),
			Sha:     os.Getenv("GITHUB_SHA"),
			IsPush:  os.Getenv("GITHUB_EVENT_NAME") == "push",
		}).
		SetHeader("Accept", "application/json").
		SetHeader("X-token", cl.token).
		Post(cl.host + "/api/job/archlinux")

	if err != nil {
		log.Fatal(err)
	}

	if resp.StatusCode() != 200 {
		log.Fatal("trigger build fail, StatusCode not 200")
	}

	var jobArchlinux JobTriggerJenkins

	json.Unmarshal([]byte(resp.Body()), &jobArchlinux)

	cl.id = jobArchlinux.ID
}

// triggerBuild build deb
func (cl *Client) PostApiJobBuild() {
	client := resty.New()
	client.SetRetryCount(3).SetRetryWaitTime(5 * time.Second).SetRetryMaxWaitTime(20 * time.Second)

	author, email, ref, err := cl.GetPRAuthorAndRef(GetOwner(), GetProject(), GetReqId())
	if err != nil {
		// Ignore failure
		log.Println("get pr author fail: ", err)
	}
	commitSHA := cl.getCommitSHA(GetOwner(), GetProject(), GetReqId())

	// branch must start with topic- will added into topic repo, or will not imported to any topic repo
	if strings.HasPrefix(ref, TOPIC_PREFIX) {
		ref = ref[TOPIC_SIZE:]
	} else {
		ref = ""
	}

	githubEvent := cl.getGithubEvent()

	resp, err := client.R().
		//// debug pr https://github.com/linuxdeepin/dde-dock/pull/364
		//SetBody(Build{
		//	Branch:        "master",
		//	CommentAuthor: "golf",
		//	GroupName:     "linuxdeepin",
		//	Project:       "dde-dock",
		//	RequestEvent:  "pull_request",
		//	RequestId:     364,
		//}).
		SetBody(Build{
			Branch:        GetBranch(),
			GroupName:     GetOwner(),
			Project:       GetProject(),
			RequestEvent:  githubEvent,
			RequestId:     GetReqId(),
			CommentAuthor: author,
			AuthorEmail:   email,
			Topic:         ref,
			ReversionID:   commitSHA,
		}).
		SetHeader("Accept", "application/json").
		SetHeader("X-token", cl.token).
		Post(cl.host + "/api/job/build")

	if err != nil {
		log.Fatal(err)
	}

	if resp.StatusCode() != 200 {
		log.Fatal("trigger build fail, StatusCode not 200")
	}

	var jobBuild JobTriggerJenkins

	err = json.Unmarshal([]byte(resp.Body()), &jobBuild)
	if err != nil {
		log.Fatal("trigger failed, response can't deserialize to jobBuild")
	}

	cl.id = jobBuild.ID
}

// SetupCloseHandler creates a 'listener' on a new goroutine which will notify the
// program if it receives an interrupt from the OS. We then handle this by calling
// our clean up procedure and exiting the program.
func (cl *Client) SetupCloseHandler() {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\r- Ctrl+C pressed in Terminal")
		cl.GetApiJobCancel()
		os.Exit(0)
	}()
}
