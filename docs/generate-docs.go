package main

import (
	"github.com/cslamar/audiobooker/cmd"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra/doc"
)

func main() {
	c := cmd.RootCmd
	err := doc.GenMarkdownTree(c, "./docs")
	if err != nil {
		log.Fatalln(err)
	}
}
