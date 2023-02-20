/*
Copyright © 2023 Chris Slamar chris@slamar.com
*/
package cmd

import (
	"errors"
	"fmt"
	"github.com/cslamar/audiobooker/audiobooker"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// batchFromTagsCmd represents the files command
var batchFromTagsCmd = &cobra.Command{
	Use:   "from-tags",
	Short: `Bind audiobook combining title tag of each file as chapter names`,
	Long:  `TODO`, // TODO
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("starting batch from tags\n\n")
		processStart := time.Now()
		var err error

		// Parse scoped options
		sourceFilesRoot, err := cmd.Flags().GetString("source-files-root")
		if err != nil {
			return err
		}

		pathPattern, err := cmd.Flags().GetString("path-pattern")
		if err != nil {
			return err
		}

		// slice of directories that contain books
		bookDirs := make([]string, 0)
		log.Debugln("src files root:", sourceFilesRoot)
		err = filepath.WalkDir(sourceFilesRoot, func(path string, d fs.DirEntry, err error) error {
			// return if not a directory
			if !d.IsDir() {
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
			// output the last level directories
			log.Debugln("found book directory at:", path)
			bookDirs = append(bookDirs, path)
			return nil
		})

		// if no book directories were found, error out
		if len(bookDirs) == 0 {
			return errors.New("no book directories found in path")
		}

		// loop through found books
		for _, dir := range bookDirs {
			// create config struct and parse ENV variables for configs
			config := audiobooker.Config{}
			defer config.Cleanup()
			if err := config.Parse(); err != nil {
				return err
			}

			// watch for early terminations
			go watchForTermSignals(&config)

			// generate and validate configs
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
			pathTags, err := audiobooker.ParsePathTags(dir, fullPath)
			if err != nil {
				return err
			}

			// get the source files path to current book directory
			config.SourceFilesPath = dir

			// create book instance and generate metadata from path
			book := audiobooker.Book{}
			book.ParseFromPattern(pathTags)

			// initialize config
			if err := config.New(); err != nil {
				return err
			}
			log.Debugln(book)

			// compute output filename from metadata and patterns
			if err := config.SetOutputFilename(book); err != nil {
				return err
			}

			fmt.Println("book found at:", dir)
			for k, v := range pathTags {
				log.Printf("%+15s: %s\n", k, v)
			}
			fmt.Printf("output filepath: %s\n\n", filepath.Join(config.OutputPath, config.OutputFile))

			// if dry-run flag is given, output metadata for validation but don't convert
			if dryRun {
				fmt.Printf("dry-run flag was set, skipping conversion\n\n")
				continue
			}
			log.Debugln("Beginning conversion")
			// make output directory paths
			if err := os.MkdirAll(config.OutputPath, 0755); err != nil {
				return err
			}

			// pre-transcode files
			if err := audiobooker.TranscodeSourceFiles(&config); err != nil {
				return err
			}

			// Parse files to chapters inside Book struct/object
			if err := book.ParseToChapters(config); err != nil {
				return err
			}

			log.Debugln(book)

			// Generate metadata for book
			if err := book.GenerateMetaTemplate(config); err != nil {
				return err
			}

			// combine pre-transcode files
			if err := audiobooker.Combine(config); err != nil {
				return err
			}

			// Apply metadata to output file
			if err := audiobooker.Bind(config, book); err != nil {
				return err
			}
			log.Debugln("that one is done")
		}

		fmt.Println("Entire process took:", time.Now().Sub(processStart))
		fmt.Println("fin.")
		return nil
	},
}

func init() {
	batchCmd.AddCommand(batchFromTagsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// filesCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// filesCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
