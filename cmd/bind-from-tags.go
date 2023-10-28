/*
Copyright © 2023 Chris Slamar chris@slamar.com
*/
package cmd

import (
	"fmt"
	"github.com/cslamar/audiobooker/audiobooker"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"path/filepath"
	"time"
)

// bindFromTagsCmd represents the files command
var bindFromTagsCmd = &cobra.Command{
	Use:   "from-tags",
	Short: `Bind audiobook combining title tag of each file as chapter names`,
	Long:  `Bind audiobook combining title tag of each file as chapter names in the compiled audiobook`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("Starting bind from tags\n\n")
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

		if dryRun {
			fmt.Println("dry-run flag was set, skipping conversion, but outputting meta")
			return nil
		}
		log.Debugln(book)

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

		notifyFinishedBook(book, processStart)
		fmt.Println("fin.")

		return nil
	},
}

func init() {
	bindCmd.AddCommand(bindFromTagsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// filesCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// filesCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
