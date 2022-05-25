package main

import "os"

func subcommands(args []string) bool {
	if len(os.Args) < 2 {
		return false
	}
	switch os.Args[1] {
	case "wait": // wait job end
		wait()
		return true
	case "cat": // print job artifact
		cat()
		return true
	case "apipass":
		apipass()
		return true
	}
	return false
}
