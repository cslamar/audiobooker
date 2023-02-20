/*
Copyright © 2022 Chris Slamar chris@slamar.com
*/
package cmd

import (
	"fmt"
	"github.com/cslamar/audiobooker/audiobooker"
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
)

// splitChaptersCmd represents the split-chapters command
var splitChaptersCmd = &cobra.Command{
	Use:   "split-chapters",
	Short: "Splits a single audio file into chapters using a fixed length",
	Long:  `A longer description of the bind process.`, // TODO
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("Starting split chatper bind\n\n")
		processStart := time.Now()
		var err error

		chapterLength, err := cmd.Flags().GetInt("chapter-length")
		if err != nil {
			return err
		}
		useEmbedded, err := cmd.Flags().GetBool("use-embedded")
		if err != nil {
			return err
		}

		// create config struct and parse ENV variables for configs
		config := audiobooker.Config{}
		defer config.Cleanup()
		if err := config.Parse(); err != nil {
			return err
		}

		config.ExternalChapters = useEmbedded

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

		// output parsed metadata
		for k, v := range pathTags {
			fmt.Printf("%+15s: %s\n", k, v)
		}
		fmt.Printf("output filepath: %s\n\n", filepath.Join(config.OutputPath, config.OutputFile))

		if dryRun {
			fmt.Println("dry-run flag was set, skipping conversion")
			return nil
		}
		log.Debugln(book)

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

		if !config.ExternalChapters {
			fmt.Println("generating static chapters based on specified chapter length")
			if err := book.GenerateStaticChapters(config, chapterLength); err != nil {
				return err
			}
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
	bindCmd.AddCommand(splitChaptersCmd)
	// define flags for this command
	splitChaptersCmd.Flags().IntP("chapter-length", "c", 5, "chapter length in minutes")
	splitChaptersCmd.Flags().Bool("use-embedded", false, "use existing embedded chapters")
	// Here you will define your flags and configuration settings.
	//splitChaptersCmd.MarkFlagRequired("source-files-path")
	splitChaptersCmd.MarkPersistentFlagRequired("chapter-length")

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// splitChaptersCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// splitChaptersCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
