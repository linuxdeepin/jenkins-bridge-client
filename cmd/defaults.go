package cmd

import (
	"os"
)

//command default value
var (
	defaultToken  = os.Getenv("BRIDGE_TOKEN")
	defaultServer = "https://jenkins-bridge-deepin-pre.uniontech.com"
	defaultRunid  = 0
)

// var to receive from command flag
var (
	token  string
	server string
	runid  int
)
