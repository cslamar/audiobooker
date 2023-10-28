/*
Copyright © 2023 Chris Slamar chris@slamar.com
*/
package cmd

import (
	"fmt"
	"github.com/cslamar/audiobooker/audiobooker"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// bindTagCmd represents the files command
var bindTagCmd = &cobra.Command{
	Use:   "tag",
	Short: `Write tags to target audiobooks based on directory structures and path-pattern`,
	Long:  `Write tags to target audiobooks based on directory structures and path-pattern.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("Starting bind tagging\n\n")
		processStart := time.Now()
		var err error

		// create config struct and parse ENV variables for configs
		config := audiobooker.Config{}
		defer config.Cleanup()
		if err := config.Parse(); err != nil {
			return err
		}

		// watch for early terminations
		go watchForTermSignals(&config)

		// generate and validate configs
		if err := generateBindOpts(&config, cmd.Flags()); err != nil {
			return err
		}
		// populate Config
		if err := config.New(); err != nil {
			return err
		}

		// source file validation
		info, err := os.Stat(config.SourceFilesPath)
		if err != nil {
			return errors.Errorf("source file is not valid: %v\n", err)
		}
		if info.IsDir() {
			return errors.New("source path must be a file, not a directory")
		}
		if !strings.HasSuffix(info.Name(), ".m4b") {
			return errors.New("source path must be a '.m4b' file")
		}

		// parse source based on pattern
		pathTags, err := audiobooker.ParsePathTags(config.SourceFilesPath, config.PathPattern)
		if err != nil {
			return err
		}

		// create book instance and generate metadata from path
		book := audiobooker.Book{}
		book.ParseFromPattern(pathTags)

		log.Debugln(book)

		fmt.Println("Parsed Tags")
		for k, v := range pathTags {
			fmt.Printf("%+15s: %s\n", k, v)
		}
		fmt.Println("output file:", config.SourceFilesPath)

		// if dry-run flag is given, output metadata for validation but don't convert
		if dryRun {
			fmt.Printf("dry-run flag was set, skipping action\n\n")
			return nil
		}

		// write supplied tags to file
		if err := book.WriteTags(config.SourceFilesPath); err != nil {
			return err
		}

		cmdNotify(fmt.Sprintf("Finished tagging %s - %s", book.Author, book.Title), "Finished")
		fmt.Println("Tagging took:", time.Now().Sub(processStart))
		fmt.Println("fin.")

		return nil
	},
}

func init() {
	bindCmd.AddCommand(bindTagCmd)
}
