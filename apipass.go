package main

import (
	"context"
	"flag"
	"log"
	"os"
	"strconv"
)

func apipass() {
	var server, token string
	fg := flag.NewFlagSet("apipass", flag.ExitOnError)
	fg.StringVar(&server, "server", "", "server")
	fg.StringVar(&token, "token", "", "token (env BRIDGE_TOKEN)")
	fg.Parse(os.Args[2:])
	if len(token) == 0 {
		token = os.Getenv("BRIDGE_TOKEN")
	}
	cl := NewClient()
	cl.host = server
	cl.token = token

	owner := GetOwner()
	project := GetProject()

	pr, _, err := cl.gh.PullRequests.Get(context.Background(), owner, project, GetReqId())
	if err != nil {
		log.Fatal(err)
	}
	var body struct {
		Pass bool   `json:"pass"`
		Msg  string `json:"msg"`
	}
	params := map[string]string{
		"group":         owner,
		"project":       project,
		"branch":        getBranch(),
		"request_id":    strconv.Itoa(GetReqId()),
		"request_event": getEvent(),
		"commit_id":     pr.Head.GetSHA(),
	}
	log.Println("parmas", params)
	resp, err := cl.R().
		SetQueryParams(params).
		SetResult(&body).
		Get(cl.host + "/api/apicheck/status")
	if err != nil {
		log.Println(err)
	}
	if resp.StatusCode() != 200 {
		log.Fatal(resp.Status())
	}
	if !body.Pass {
		log.Fatal(body.Msg)
	}
}
