package client

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/dlclark/regexp2"
	"github.com/go-resty/resty/v2"
)

// direct to get jenkins job such as log, artifact and make it cancel

// R method creates a new request instance, its used for Get, Post, Put, Delete, Patch, Head, Options, etc.
func (cl *Client) R() *resty.Request {
	client := resty.New()
	client.SetRetryCount(3).SetRetryWaitTime(5 * time.Second).SetRetryMaxWaitTime(20 * time.Second)
	return client.R().SetHeader("X-token", cl.token)
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
	client.
		SetRetryCount(3).
		SetRetryWaitTime(5 * time.Second).
		SetRetryMaxWaitTime(20 * time.Second).
		AddRetryCondition(func(r *resty.Response, err error) bool {
			return r.StatusCode() != 200
		})
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
	client.SetRetryCount(3).
		SetRetryWaitTime(5 * time.Second).
		SetRetryMaxWaitTime(20 * time.Second).
		AddRetryCondition(func(r *resty.Response, err error) bool {
			return r.StatusCode() != 200
		})
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
