package main

import (
	"crit-go/cmd"
	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetReportCaller(true)
}

func main() {
	cmd.Execute()
}
