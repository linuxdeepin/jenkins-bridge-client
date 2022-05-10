package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/dlclark/regexp2"
	"github.com/go-resty/resty/v2"
	"github.com/google/go-github/v44/github"
	"github.com/myml/ghtoken"
)

// Client 客户端
type Client struct {
	job_name string
	host     string
	token    string
	id       int

	gh *github.Client
}

func NewClient() *Client {
	var client Client
	tr := ghtoken.NewGitHubToken(http.DefaultTransport)
	client.gh = github.NewClient(&http.Client{Transport: tr})
	return &client
}

// GetApiJobCancel 取消任务
func (cl *Client) GetApiJobCancel() {

	client := resty.New()
	client.SetRetryCount(3).SetRetryWaitTime(5 * time.Second).SetRetryMaxWaitTime(20 * time.Second)
	resp, err := client.R().
		SetQueryParams(map[string]string{
			"id": strconv.Itoa(cl.id),
		}).
		SetHeader("Accept", "application/json").
		SetHeader("X-token", cl.token).
		Get(cl.host + "/api/job/cancel")

	if err != nil {
		log.Fatal(err)
	}

	if resp.StatusCode() != 200 {
		log.Fatal("Cancel build fail, StatusCode not 200")
	} else {
		log.Println("Cancel build success")
	}
}

// JobLog 构建日志
type JobLog struct {
	Content string `json:"Content"`
	Offset  int    `json:"Offset"`
}

// GetLog 获取构建日志内容和偏移量
func (cl *Client) GetApiJobLog(offset int) (string, int) {
	client := resty.New()
	client.SetRetryCount(3).SetRetryWaitTime(5 * time.Second).SetRetryMaxWaitTime(20 * time.Second)
	resp, err := client.R().
		SetQueryParams(map[string]string{
			"id":     strconv.Itoa(cl.id),
			"offset": strconv.Itoa(offset),
		}).
		SetHeader("Accept", "application/json").
		SetHeader("X-token", cl.token).
		Get(cl.host + "/api/job/log")

	if err != nil {
		log.Fatal(err)
	}

	var joblog JobLog
	json.Unmarshal([]byte(resp.Body()), &joblog)

	return joblog.Content, joblog.Offset
}

//JobInfo 构建状态
type JobInfo struct {
	Stages []struct {
		Name   string `json:"Name"`
		Status string `json:"Status"`
	} `json:"Stages"`
	Status string `json:"Status"`
}

// GetApiJobInfo 获取构建任务状态
func (cl *Client) GetJobStatus() string {

	client := resty.New()
	client.SetRetryCount(3).SetRetryWaitTime(5 * time.Second).SetRetryMaxWaitTime(20 * time.Second)
	resp, err := client.R().
		SetQueryParams(map[string]string{
			"id": strconv.Itoa(cl.id),
		}).
		SetHeader("Accept", "application/json").
		SetHeader("X-token", cl.token).
		Get(cl.host + "/api/job/info")

	if resp.StatusCode() != 200 {
		log.Fatal("get job status fail, StatusCode not 200")
	}

	if err != nil {
		log.Fatal(err)
	}

	var jobstatus JobInfo
	json.Unmarshal([]byte(resp.Body()), &jobstatus)

	return jobstatus.Status
}

// Artifacts 构建产物
type Artifact struct {
	Name string `json:"Name"`
	URL  string `json:"URL"`
}

type Artifacts struct {
	// Files 构建产物
	Files []Artifact `json:"Files"`
}

// GetApiJobArtifacts 获取构建产物清单
func (cl *Client) GetApiJobArtifacts() Artifacts {

	client := resty.New()
	client.SetRetryCount(3).SetRetryWaitTime(5 * time.Second).SetRetryMaxWaitTime(20 * time.Second)
	resp, err := client.R().
		SetQueryParams(map[string]string{
			"id": strconv.Itoa(cl.id),
		}).
		SetHeader("Accept", "application/json").
		SetHeader("X-token", cl.token).
		Get(cl.host + "/api/job/artifacts")

	if err != nil {
		log.Fatal(err)
	}

	var artifacts Artifacts

	json.Unmarshal([]byte(resp.Body()), &artifacts)

	return artifacts
}

// DownloadArtifacts 下载构建产物
func (cl *Client) DownloadArtifacts() {
	// 获取所有产物
	artifacts := cl.GetApiJobArtifacts()
	//log.Println(artifacts)
	// 实际下载清单
	var realArtifacts []Artifact
	// 创建 ../artifacts/ 目录以存放构建产物
	// 匹配: *.deb
	// 不匹配: *-dbgsym_*.deb
	r := regexp2.MustCompile("^(?!.*dbgsym_).*\\.deb", 0)

	for i := 0; i < len(artifacts.Files); i++ {
		if isMatch, _ := r.MatchString(artifacts.Files[i].Name); isMatch {
			realArtifacts = append(realArtifacts, artifacts.Files[i])
			log.Println("Artifacts Matched: " + artifacts.Files[i].Name)
		} else {
			log.Println("Artifacts Skiped: " + artifacts.Files[i].Name)
		}
	}
	// 创建 ../artifacts/ 目录以存放构建产物
	artifactsDir := "./artifacts/"
	err := os.MkdirAll(artifactsDir, 0755)
	if err != nil {
		log.Fatal(err)
	}

	// 下载文件
	for i := 0; i < len(realArtifacts); i++ {
		fileLocation := artifactsDir + realArtifacts[i].Name
		client := resty.New()
		_, err := client.R().
			SetHeader("Accept", "application/json").
			SetHeader("X-token", cl.token).
			SetOutput(fileLocation).
			Get(realArtifacts[i].URL)
		if err != nil {
			log.Fatal(err)
		} else {
			log.Println("Download " + fileLocation + " Success")
		}
	}
}

// 打印日志
func (cl *Client) PrintLog() {
	offset := 0
	for {
		status := cl.GetJobStatus()
		var res string
		res, offset = cl.GetApiJobLog(offset)
		if len(res) > 0 {
			fmt.Print(res)
		}
		switch status {
		case "Success":
			return
		case "Fail":
			os.Exit(1) // Nonzero value: failure
		case "Progress":
		}
		time.Sleep(1 * time.Second)
	}
}

// Wait job end
func (cl *Client) Wait() error {
	for {
		status := cl.GetJobStatus()
		switch status {
		case "Success":
			return nil
		case "Fail":
			return fmt.Errorf("job fail")
		}
		time.Sleep(time.Second)
	}
}

// 创建打包构建任务,返回值为 id
//  /api/job/sync
//  /api/job/build
//  /api/job/sync
//  /api/job/abicheck
//  /api/job/archlinux

type JobTriggerJenkins struct {
	ID int `json:"ID"`
}

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
}

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

// R method creates a new request instance, its used for Get, Post, Put, Delete, Patch, Head, Options, etc.
func (cl *Client) R() *resty.Request {
	client := resty.New()
	client.SetRetryCount(3).SetRetryWaitTime(5 * time.Second).SetRetryMaxWaitTime(20 * time.Second)
	return client.R().SetHeader("X-token", cl.token)
}

func (cl *Client) PostApiJobSync() {
	client := resty.New()
	client.SetRetryCount(3).SetRetryWaitTime(5 * time.Second).SetRetryMaxWaitTime(20 * time.Second)
	resp, err := client.R().
		SetBody(Build{
			Project: GetProject(),
		}).
		SetHeader("Accept", "application/json").
		SetHeader("X-token", cl.token).
		Post(cl.host + "/api/job/sync")

	if resp.StatusCode() != 200 {
		log.Fatal("trigger build fail, StatusCode not 200")
	}
	if err != nil {
		log.Fatal(err)
	}
	var jobSync JobTriggerJenkins
	err = json.Unmarshal([]byte(resp.Body()), &jobSync)
	if err != nil {
		log.Fatal(err)
	}
	cl.id = jobSync.ID
}

func (cl *Client) PostApiJobAbicheck() {
	client := resty.New()
	client.SetRetryCount(3).SetRetryWaitTime(5 * time.Second).SetRetryMaxWaitTime(20 * time.Second)

	author, email, err := cl.GetPRAuthor(GetOwner(), GetProject(), GetReqId())
	if err != nil {
		// Ignore failure
		log.Println("get pr author fail: ", err)
	}
	resp, err := client.R().
		SetBody(Build{
			Branch:        os.Getenv("GITHUB_BASE_REF"),
			GroupName:     os.Getenv("GITHUB_REPOSITORY_OWNER"),
			Project:       GetProject(),
			RequestEvent:  os.Getenv("GITHUB_EVENT_NAME"),
			RequestId:     GetReqId(),
			CommentAuthor: author,
			AuthorEmail:   email,
		}).
		SetHeader("Accept", "application/json").
		SetHeader("X-token", cl.token).
		Post(cl.host + "/api/job/abicheck")

	if resp.StatusCode() != 200 {
		log.Fatal("trigger build fail, StatusCode not 200")
	}
	if err != nil {
		log.Fatal(err)
	}

	var jobAbicheck JobTriggerJenkins

	json.Unmarshal([]byte(resp.Body()), &jobAbicheck)

	cl.id = jobAbicheck.ID
}

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

	if resp.StatusCode() != 200 {
		log.Fatal("trigger build fail, StatusCode not 200")
	}
	if err != nil {
		log.Fatal(err)
	}

	var jobArchlinux JobTriggerJenkins

	json.Unmarshal([]byte(resp.Body()), &jobArchlinux)

	cl.id = jobArchlinux.ID
}

// GetPRAuthor get login name and email of the author of the pull request
func (cl *Client) GetPRAuthor(owner, project string, prID int) (author, email string, err error) {
	req, _, err := cl.gh.PullRequests.Get(context.Background(), owner, project, prID)
	if err != nil {
		return "", "", fmt.Errorf("get pull request: %w", err)
	}
	user, _, err := cl.gh.Users.Get(context.Background(), req.GetUser().GetLogin())
	if err != nil {
		return "", "", fmt.Errorf("get user: %w", err)
	}
	return user.GetLogin(), user.GetEmail(), nil
}

func (cl *Client) PostApiJobBuild() {
	client := resty.New()
	client.SetRetryCount(3).SetRetryWaitTime(5 * time.Second).SetRetryMaxWaitTime(20 * time.Second)

	author, email, err := cl.GetPRAuthor(GetOwner(), GetProject(), GetReqId())
	if err != nil {
		// Ignore failure
		log.Println("get pr author fail: ", err)
	}

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
			Branch:        os.Getenv("GITHUB_BASE_REF"),
			GroupName:     os.Getenv("GITHUB_REPOSITORY_OWNER"),
			Project:       GetProject(),
			RequestEvent:  os.Getenv("GITHUB_EVENT_NAME"),
			RequestId:     GetReqId(),
			CommentAuthor: author,
			AuthorEmail:   email,
		}).
		SetHeader("Accept", "application/json").
		SetHeader("X-token", cl.token).
		Post(cl.host + "/api/job/build")

	if resp.StatusCode() != 200 {
		log.Fatal("trigger build fail, StatusCode not 200")
	}
	if err != nil {
		log.Fatal(err)
	}

	var jobBuild JobTriggerJenkins

	json.Unmarshal([]byte(resp.Body()), &jobBuild)

	cl.id = jobBuild.ID
}

type Run struct {
	Status string `json:"Status"`
	RunID  int    `json:"RunID"`
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

func subcommands(args []string) bool {
	if len(os.Args) < 2 {
		return false
	}
	cl := NewClient()
	switch os.Args[1] {
	case "wait": // wait job end
		var runid int
		var server, token string
		waitFlag := flag.NewFlagSet("wait", flag.ExitOnError)
		waitFlag.IntVar(&runid, "runid", 0, "job runid")
		waitFlag.StringVar(&server, "server", "", "server")
		waitFlag.StringVar(&token, "token", "", "token (env BRIDGE_TOKEN)")
		waitFlag.Parse(os.Args[2:])
		if len(token) == 0 {
			token = os.Getenv("BRIDGE_TOKEN")
		}
		if runid == 0 {
			waitFlag.Usage()
			os.Exit(1)
		}
		cl.host = server
		cl.token = token
		cl.id = runid
		err := cl.Wait()
		if err != nil {
			log.Fatal(err)
		}
		return true
	case "cat": // print job artifact
		var runid int
		var server, token string
		var filename string
		catFlag := flag.NewFlagSet("cat", flag.ExitOnError)
		catFlag.IntVar(&runid, "runid", 0, "job runid")
		catFlag.StringVar(&filename, "file", "", "file path")
		catFlag.StringVar(&server, "server", "", "server")
		catFlag.StringVar(&token, "token", "", "token (env BRIDGE_TOKEN)")
		catFlag.Parse(os.Args[2:])
		if len(token) == 0 {
			token = os.Getenv("BRIDGE_TOKEN")
		}
		if runid == 0 || len(filename) == 0 {
			catFlag.Usage()
			os.Exit(1)
		}
		cl.host = server
		cl.token = token
		cl.id = runid
		for _, f := range cl.GetApiJobArtifacts().Files {
			if f.Name == filename {
				resp, err := cl.R().SetDoNotParseResponse(true).Get(f.URL)
				if err != nil {
					log.Fatal(err)
				}
				defer resp.RawBody().Close()
				io.Copy(os.Stdout, resp.RawBody())
			}
		}
		return true
	}
	return false
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	if subcommands(os.Args) {
		return
	}
	var (
		downloadArtifacts bool
		jobName           string
		token             string
		host              string
		cancelBuild       bool
		printlog          bool
		triggerAbicheck   bool
		triggerBuild      bool
		runid             int
		triggerSync       bool
		triggerArchlinux  bool
	)
	flag.BoolVar(&downloadArtifacts, "downloadArtifacts", false, "是否下载产物")
	flag.BoolVar(&printlog, "printlog", false, "是否打印日志")
	flag.BoolVar(&triggerAbicheck, "triggerAbicheck", false, "是否触发Abicheck")
	flag.BoolVar(&triggerArchlinux, "triggerArchlinux", false, "是否触发Archlinux编译")
	flag.BoolVar(&triggerBuild, "triggerBuild", false, "是否触发编译")
	flag.BoolVar(&cancelBuild, "cancelBuild", false, "是否取消编译")
	flag.BoolVar(&triggerSync, "triggerSync", false, "是否触发同步")
	flag.IntVar(&runid, "runid", 0, "job runid")
	flag.StringVar(&jobName, "jobName", "github-pipeline", "要触发的 Jenkins 任务名")
	flag.StringVar(&token, "token", "", "bridge server token")
	flag.StringVar(&host, "host", "", "bridge server address")
	flag.Parse()

	cl := NewClient()
	cl.job_name = jobName
	if len(host) > 0 {
		cl.host = host
	} else {
		cl.host = "https://jenkins-bridge-deepin-pre.uniontech.com"
	}

	if len(token) > 0 {
		cl.token = token
	}

	// cl.SetupCloseHandler()

	if triggerAbicheck {
		cl.PostApiJobAbicheck()
		fmt.Println(cl.id)
	}

	if triggerArchlinux {
		cl.PostApiJobArchlinux()
		fmt.Println(cl.id)
	}

	if triggerSync {
		cl.PostApiJobSync()
		fmt.Println(cl.id)
	}

	if triggerBuild {
		if runid != 0 {
			fmt.Println("参数中检测到 runid , 跳过构建")
		} else {
			cl.PostApiJobBuild()

			// 将 runid 打印出来以便在action steps间传递
			fmt.Println(cl.id)
		}
	}

	if runid != 0 {
		cl.id = runid
	}

	if printlog {
		cl.PrintLog()
	}

	if downloadArtifacts {
		cl.DownloadArtifacts()
	}

	if cancelBuild {
		cl.GetApiJobCancel()
	}
}
