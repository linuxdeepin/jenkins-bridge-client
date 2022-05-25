package main

import (
	"flag"
	"log"
	"os"
)

// 等待任务结束
func wait() {
	cl := NewClient()
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
}
