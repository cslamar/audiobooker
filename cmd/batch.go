/*
Copyright © 2023 Chris Slamar chris@slamar.com
*/
package cmd

import (
	"errors"
	"fmt"
	"github.com/cslamar/audiobooker/audiobooker"
	"github.com/spf13/pflag"

	"github.com/spf13/cobra"
)

// batchCmd represents the batch command
var batchCmd = &cobra.Command{
	Use:   "batch",
	Short: "Perform batched operations on a pattern of directories for multiple audiobook binding",
	Long:  `The batch command, and its sub-commands, will perform actions to create multiple audiobooks based on a collection of structured directories.  The way that the audiobooks are created is based on the pattern of the directory structure and sub-commands selected.  This is used when you want to convert a larger collection of books.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("batch called")
		return nil
	},
}

func init() {
	RootCmd.AddCommand(batchCmd)

	batchCmd.PersistentFlags().StringP("file-pattern", "f", "", "The output filename, can be a combination of literal values and patterns")
	batchCmd.PersistentFlags().IntP("jobs", "j", 1, "The number of concurrent transcoding process to run for conversion (don't exceed your cpu count)")
	batchCmd.PersistentFlags().StringP("output-directory", "o", "", "The output directory for the final directory, can be combination of absolute values and path patterns")
	batchCmd.PersistentFlags().StringP("path-pattern", "p", "", "The pattern for metadata picked up via paths (starts from base of source-files-root)")
	batchCmd.PersistentFlags().String("scratch-files-path", "", "The location to generate the scratch directory")
	batchCmd.PersistentFlags().StringP("source-files-root", "s", "", "The path to directory of source files (must match path-pattern for metadata to work)")
	batchCmd.PersistentFlags().Bool("verbose-transcode", false, "Enable output of all ffmpeg commands/operations")

	batchCmd.MarkPersistentFlagRequired("source-files-root")
}

// generateBatchOpts configures and validates bind flags
func generateBatchOpts(config *audiobooker.Config, flags *pflag.FlagSet) error {
	// Apply configs in the order of cli overrides env which overrides config

	// check for verbose logging output
	verboseTranscode, err := flags.GetBool("verbose-transcode")
	if err != nil {
		return err
	}
	if verboseTranscode {
		config.VerboseTranscode = true
	}

	// get path pattern
	pathPattern, err := flags.GetString("path-pattern")
	if err != nil {
		return err
	} else if pathPattern != "" {
		config.PathPattern = pathPattern
	}

	// get file pattern
	filePatten, err := flags.GetString("file-pattern")
	if err != nil {
		return err
	} else if filePatten != "" {
		config.OutputFilePattern = filePatten
	}

	// get jobs count
	jobs, err := flags.GetInt("jobs")
	if err != nil {
		return err
	} else if jobs > config.Jobs {
		config.Jobs = jobs
	}

	// get output directory
	outputDir, err := flags.GetString("output-directory")
	if err != nil {
		return err
	} else if outputDir != "" {
		config.OutputFileDest = outputDir
		config.OutputPathPattern = outputDir
	}

	// get scratch directory
	scratchFilesPath, err := flags.GetString("scratch-files-path")
	if err != nil {
		return err
	} else if scratchFilesPath != "" {
		config.ScratchFilesPath = scratchFilesPath
	}

	// get source files path
	//sourceFilesPath, err := flags.GetString("source-files-root")
	//if err != nil {
	//	return err
	//} else {
	//	config.SourceFilesPath = sourceFilesPath
	//}

	// validate selected options

	// validate that some path pattern variable is defined
	if config.PathPattern == "" && pathPattern == "" {
		return errors.New("path pattern must be defined")
	}
	// validate jobs
	if config.Jobs <= 0 {
		return errors.New("jobs must be greater than 0")
	}
	// validate output destination in config struct
	if config.OutputFileDest == "" {
		return errors.New("output-directory must not be empty")
	}

	return nil
}
