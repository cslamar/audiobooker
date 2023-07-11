package main

import (
	"github.com/cslamar/audiobooker/cmd"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra/doc"
)

func main() {
	c := cmd.RootCmd
	c.DisableAutoGenTag = true
	err := doc.GenMarkdownTree(c, "./docs/cli-usage")
	if err != nil {
		log.Fatalln(err)
	}
}
