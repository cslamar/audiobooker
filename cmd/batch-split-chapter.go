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

// batchSplitCmd represents the files command
var batchSplitCmd = &cobra.Command{
	Use:   "split-chapters",
	Short: `Splits a single audio file into chapters using a fixed length`,
	Long:  `TODO`, // TODO
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("Starting batch splitting chapters\n\n")
		processStart := time.Now()
		var err error

		// Parse scoped options
		chapterLength, err := cmd.Flags().GetInt("chapter-length")
		if err != nil {
			return err
		}

		sourceFilesRoot, err := cmd.Flags().GetString("source-files-root")
		if err != nil {
			return err
		}

		pathPattern, err := cmd.Flags().GetString("path-pattern")
		if err != nil {
			return err
		}

		useEmbedded, err := cmd.Flags().GetBool("use-embedded")
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

		for _, dir := range bookDirs {
			// create config struct and parse ENV variables for configs
			config := audiobooker.Config{}
			defer config.Cleanup()
			if err := config.Parse(); err != nil {
				return err
			}

			config.ExternalChapters = useEmbedded

			// watch for early terminations
			go watchForTermSignals(&config)

			// generate and validate flags
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
				fmt.Printf("%+15s: %s\n", k, v)
			}
			fmt.Printf("output filepath: %s\n\n", filepath.Join(config.OutputPath, config.OutputFile))

			// if dry-run flag is given, output metadata for validation but don't convert
			if dryRun {
				fmt.Printf("dry-run flag was set, skipping conversion\n\n")
				continue
			}
			log.Debugln("Beginning conversion")

			// extract embedded chapters if instructed
			if config.ExternalChapters {
				fmt.Println("extracting exiting chapters metadata instead of generating static chapters")
				if err := book.ExtractChapters(config); err != nil {
					return err
				}
			}

			// make output directory paths
			if err := os.MkdirAll(config.OutputPath, 0755); err != nil {
				return err
			}
			if err := audiobooker.SplitSingleFile(&config); err != nil {
				return err
			}

			if err := audiobooker.TranscodeSourceFiles(&config); err != nil {
				return err
			}

			log.Debugln(book)

			// Generate static chapters metadata for book
			if !config.ExternalChapters {
				fmt.Println("generating static chapters based on specified chapter length")
				if err := book.GenerateStaticChapters(config, chapterLength); err != nil {
					return err
				}
			}

			// generate chapters metadata
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

		}

		fmt.Println("Entire process took:", time.Now().Sub(processStart))
		fmt.Println("fin.")
		return nil
	},
}

func init() {
	batchCmd.AddCommand(batchSplitCmd)

	batchSplitCmd.Flags().IntP("chapter-length", "c", 5, "chapter length in minutes")
	batchSplitCmd.Flags().Bool("use-embedded", false, "use existing embedded chapters")
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// filesCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// filesCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}