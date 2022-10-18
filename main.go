package main

import "github.com/bhatti/api-mock-service/cmd"

var (
	version = "xdev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	cmd.Execute(version, commit, date)
}
