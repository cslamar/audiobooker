/*
Copyright © 2022 Chris Slamar chris@slamar.com
*/
package cmd

import (
	"fmt"
	"github.com/cslamar/audiobooker/audiobooker"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"time"
)

// filesCmd represents the files command
var filesCmd = &cobra.Command{
	Use:   "files",
	Short: "Bind audiobook using each file as a chapter",
	Long:  `Bind audiobook using each file as a chapter using either the source audio filename as the chapter name, or the source audio file's "title" metadata tag as the chapter name.'`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("Starting bind by filename\n\n")
		processStart := time.Now()
		var err error

		// Parse scoped options
		useFileNames, err := cmd.Flags().GetBool("file-name")
		if err != nil {
			return err
		}
		useTitleTag, err := cmd.Flags().GetBool("title-tag")
		if err != nil {
			return err
		}

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

		pathTags, err := audiobooker.ParsePathTags(config.SourceFilesPath, config.PathPattern) // TODO change this to pass in just the config struct
		if err != nil {
			return err
		}

		book := audiobooker.Book{}
		book.ParseFromPattern(pathTags)
		if err := config.SetOutputFilename(book); err != nil {
			return err
		}

		for k, v := range pathTags {
			fmt.Printf("%+15s: %s\n", k, v)
		}
		fmt.Printf("output filepath: %s\n\n", filepath.Join(config.OutputPath, config.OutputFile))

		// if dry-run flag is given, output metadata for validation but don't convert
		if dryRun {
			fmt.Println("dry-run flag was set, skipping conversion, but outputting meta")
			return nil
		}
		log.Debugln(book)

		// make output directory paths
		if err := os.MkdirAll(config.OutputPath, 0755); err != nil {
			return err
		}

		if err := audiobooker.TranscodeSourceFiles(&config); err != nil {
			return err
		}

		if err := book.ChapterByFile(config, useFileNames, useTitleTag); err != nil {
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

		fmt.Println("Entire process took:", time.Now().Sub(processStart))
		fmt.Println("fin.")

		return nil
	},
}

func init() {
	bindCmd.AddCommand(filesCmd)
	filesCmd.Flags().Bool("file-name", false, "Use the name of the file as the chapter name")
	filesCmd.Flags().Bool("title-tag", false, "Use the file's title tag as the chapter name")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// filesCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// filesCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
