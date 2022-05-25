package main

import (
	"flag"
	"io"
	"log"
	"os"
)

// 在控制台打印编译产物，一般用于输出编辑警告，编译日志等
func cat() {
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
	cl := NewClient()
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
}
