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
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// splitChaptersCmd represents the split-chapters command
var splitChaptersCmd = &cobra.Command{
	Use:   "split-chapters",
	Short: "Splits a single audio file into chapters using a fixed length",
	Long: `Bind split-chapters will split a single file into a chapter marked audiobook file based on two options.  

First a static number (in minutes) can be passed in to make hard chapter marks at the specified duration.  Each mark will result in chapter metadata being created at those increments with the name "Chapter X" (where X in the index).

The other way that split-chapters can be used is if the existing file already has metadata embedded.  Passing in the '--use-embedded' flag will use that metadata when creating the chapters for the new audiobook file.`,
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
		generateChapters, err := cmd.Flags().GetBool("generate-chapters")
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

		// TODO find a better/cleaner/nicer way of handling extensions
		if strings.HasSuffix(config.OutputFile, ".m4b.m4b") {
			config.OutputFile = strings.TrimSuffix(config.OutputFile, ".m4b")
		}

		// output parsed metadata
		for k, v := range pathTags {
			fmt.Printf("%+15s: %s\n", k, v)
		}
		fmt.Printf("output filepath: %s\n\n", filepath.Join(config.OutputPath, config.OutputFile))

		if dryRun {
			fmt.Println("dry-run flag was set, skipping conversion, but outputting meta")
			return nil
		}
		log.Debugln(book)

		// process the chapter split and generate a chapters metadata file only, no encoding
		if generateChapters {
			log.Infoln("Generating/Embedding static chapters and metadata")
			if err := book.GenerateStaticChapters(config, chapterLength, config.SourceFilesPath); err != nil {
				return err
			}

			// generate chapters metadata
			if err := book.GenerateMetaTemplate(config); err != nil {
				log.Errorln(err)
				return err
			}

			log.Debugln(book.Chapters)

			// embed metadata
			if err := audiobooker.Bind(config, book); err != nil {
				log.Errorln(err)
				return err
			}

			return nil
		}

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
			if err := book.GenerateStaticChapters(config, chapterLength, ""); err != nil {
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

		notifyFinishedBook(book, processStart)
		fmt.Println("fin.")

		return nil
	},
}

func init() {
	bindCmd.AddCommand(splitChaptersCmd)
	// define flags for this command
	splitChaptersCmd.Flags().IntP("chapter-length", "c", 5, "chapter length in minutes")
	splitChaptersCmd.Flags().Bool("use-embedded", false, "use existing embedded chapters")
	splitChaptersCmd.Flags().Bool("generate-chapters", false, "generate chapters and embed them in and existing .m4b audiobook (no transcoding required)")
}
