/*
Copyright © 2023 Chris Slamar chris@slamar.com
*/
package cmd

import (
	"errors"
	"fmt"
	"github.com/cslamar/audiobooker/audiobooker"
	log "github.com/sirupsen/logrus"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// batchTagCmd represents the files command
var batchTagCmd = &cobra.Command{
	Use:   "tag",
	Short: `Write tags to target audiobooks based on directory structures and path-pattern`,
	Long:  `Write tags to target audiobooks based on directory structures and path-pattern.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("Starting batch tagging\n\n")
		processStart := time.Now()
		var err error

		sourceFilesRoot, err := cmd.Flags().GetString("source-files-root")
		if err != nil {
			return err
		}

		pathPattern, err := cmd.Flags().GetString("path-pattern")
		if err != nil {
			return err
		}

		// slice of directories that contain books
		//bookDirs := make([]string, 0)
		audiobookFiles := make([]string, 0)
		log.Debugln("src files root:", sourceFilesRoot)
		err = filepath.WalkDir(sourceFilesRoot, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			// if a file is found, and it ends with '.m4b' add it to the audiobooks slice
			if !d.IsDir() {
				if strings.HasSuffix(d.Name(), ".m4b") {
					log.Debugln("found book at:", path)
					audiobookFiles = append(audiobookFiles, path)
				}
				return nil
			}
			// scan files in path
			dirs, err := os.ReadDir(path)
			if err != nil {
				return err
			}
			// if this directory contains directories, skip
			for _, dir := range dirs {
				if dir.Type().IsDir() {
					return nil
				}
			}
			return nil
		})
		if err != nil {
			return err
		}

		log.Debugln(audiobookFiles)

		// if no books were found, error out
		if len(audiobookFiles) == 0 {
			return errors.New("no books found in path")
		}

		for _, audiobook := range audiobookFiles {
			// create config struct and parse ENV variables for configs
			config := audiobooker.Config{}
			defer config.Cleanup()
			if err := config.Parse(); err != nil {
				return err
			}

			// watch for early terminations
			go watchForTermSignals(&config)

			// generate and validate flags, get around validation with hardcoded "none" for output file dest
			if err := generateBatchOpts(&config, cmd.Flags()); err != nil {
				return err
			}

			// validate full path formatting
			var fullPath string
			if strings.HasSuffix(sourceFilesRoot, "/") {
				fullPath = sourceFilesRoot + pathPattern
			} else {
				fullPath = sourceFilesRoot + "/" + pathPattern
			}
			// parse source based on pattern
			pathTags, err := audiobooker.ParsePathTags(audiobook, fullPath)
			if err != nil {
				return err
			}

			// get the source files path to current book directory
			config.SourceFilesPath = audiobook

			// create book instance and generate metadata from path
			book := audiobooker.Book{}
			book.ParseFromPattern(pathTags)

			// initialize config
			if err := config.New(); err != nil {
				return err
			}
			log.Debugln(book)

			fmt.Println("book found at:", audiobook)
			for k, v := range pathTags {
				fmt.Printf("%+15s: %s\n", k, v)
			}
			fmt.Printf("output file: %s\n\n", audiobook)
			// if dry-run flag is given, output metadata for validation but don't convert
			if dryRun {
				fmt.Printf("dry-run flag was set, skipping action\n\n")
				continue
			}
			log.Debugln("Beginning tagging")

			// write supplied tags to file
			if err := book.WriteTags(audiobook); err != nil {
				return err
			}

			log.Debugln(book)
			cmdNotify(fmt.Sprintf("Finished tagging %s - %s", book.Author, book.Title), "Finished")
		}

		fmt.Println("Entire process took:", time.Now().Sub(processStart))
		fmt.Println("fin.")
		return nil
	},
}

func init() {
	batchCmd.AddCommand(batchTagCmd)
}
