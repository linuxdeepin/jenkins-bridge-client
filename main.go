package main

import (
	"flag"
	"fmt"
	"log"

	"jenkins-bridge-client/client"
	"jenkins-bridge-client/cmd"
)

func main() {

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// for old command compatibility, once old style flag used, do not use subcommand
	var processed = false

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

	cl := client.NewClient()
	cl.SetJonName(jobName)
	if len(host) > 0 {
		cl.SetHost(host)
	} else {
		cl.SetHost("https://jenkins-bridge-deepin-pre.uniontech.com")
	}

	if len(token) > 0 {
		cl.SetToken(token)
	}

	// cl.SetupCloseHandler()

	if triggerAbicheck {
		processed = true
		cl.PostApiJobAbicheck()
		fmt.Println(cl.GetID())
	}

	if triggerArchlinux {
		processed = true
		cl.PostApiJobArchlinux()
		fmt.Println(cl.GetID())
	}

	if triggerSync {
		processed = true
		cl.PostApiJobSync()
		fmt.Println(cl.GetID())
	}

	if triggerBuild {
		processed = true
		if runid != 0 {
			fmt.Println("参数中检测到 runid , 跳过构建")
		} else {
			cl.PostApiJobBuild()
			// 将 runid 打印出来以便在action steps间传递
			fmt.Println(cl.GetID())
		}
	}

	if runid != 0 {
		cl.SetID(runid)
	}

	if printlog {
		processed = true
		cl.PrintLog()
	}

	if downloadArtifacts {
		processed = true
		cl.DownloadArtifacts()
	}

	if cancelBuild {
		processed = true
		cl.GetApiJobCancel()
	}
	if !processed {
		cmd.Execute()
	}
}
