/*
Copyright © 2022 Chris Slamar chris@slamar.com
*/
package main

import (
	"fmt"
	"github.com/cslamar/audiobooker/cmd"
	log "github.com/sirupsen/logrus"
	go_log "log"
	"os"
	"path"
	"runtime"
)

// ReportCallerFlag Use build flags to toggle report caller logging
var ReportCallerFlag = ""

func init() {
	log.SetLevel(log.InfoLevel)
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:          true,
		DisableTimestamp:       true,
		DisableLevelTruncation: true,
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			filename := path.Base(f.File)
			return fmt.Sprintf("%s()", f.Function), fmt.Sprintf("%s:%d", filename, f.Line)
		},
	})

	// set Go's builtin logging to output no where (for now) due to a behavior of the ffmpeg package
	devNull, err := os.Open("/dev/null")
	if err != nil {
		panic(err)
	}
	go_log.SetOutput(devNull)

	// Use build flags to toggle report caller logging
	if ReportCallerFlag != "" {
		log.SetReportCaller(true)
	}
}

func main() {
	cmd.Execute()
}
