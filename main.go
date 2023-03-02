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
	"strings"
)

func init() {
	log.SetLevel(log.InfoLevel)
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:          true,
		DisableTimestamp:       true,
		DisableLevelTruncation: true,
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			filename := path.Base(f.File)
			callerFuncParts := strings.Split(f.Function, "/")
			callerFunc := callerFuncParts[len(callerFuncParts)-1]
			return fmt.Sprintf("%s()", callerFunc), fmt.Sprintf(" %s:%d", filename, f.Line)
			//return fmt.Sprintf("%s()", f.Function), fmt.Sprintf(" %s:%d", filename, f.Line)
		},
	})

	// set Go's builtin logging to output no where (for now) due to a behavior of the ffmpeg package
	devNull, err := os.Open("/dev/null")
	if err != nil {
		panic(err)
	}
	go_log.SetOutput(devNull)

	// default report caller to be off
	log.SetReportCaller(false)
}

func main() {
	cmd.Execute()
}
