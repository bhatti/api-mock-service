package main

import (
	"embed"
	"github.com/bhatti/api-mock-service/cmd"
)

var (
	version = "xdev"
	commit  = "none"
	date    = "unknown"
)

// swaggerContent holds our swagger-ui content.
//
//go:embed swagger-ui/*
var swaggerContent embed.FS

func main() {
	cmd.Execute(version, commit, date, swaggerContent)
}
