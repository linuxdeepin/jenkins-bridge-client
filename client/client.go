package client

import (
	"net/http"

	"github.com/google/go-github/github"
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

func (cl *Client) SetJonName(jobName string) {
	cl.job_name = jobName
}

func (cl *Client) SetHost(host string) {
	cl.host = host
}

func (cl *Client) SetToken(token string) {
	cl.token = token
}

func (cl *Client) SetID(id int) {
	cl.id = id
}

func (cl *Client) GetJobName() string {
	return cl.job_name
}

func (cl *Client) GetHost() string {
	return cl.host
}

func (cl *Client) GetToken() string {
	return cl.token
}

func (cl *Client) GetID() int {
	return cl.id
}

func (cl *Client) GetGitHub() *github.Client {
	return cl.gh
}
