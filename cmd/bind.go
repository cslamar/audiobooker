/*
Copyright © 2022 Chris Slamar chris@slamar.com
*/
package cmd

import (
	"errors"
	"github.com/cslamar/audiobooker/audiobooker"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// bindCmd represents the bind command
var bindCmd = &cobra.Command{
	Use:   "bind",
	Short: "Combine multiple audio files into an M4B audiobook file",
	Long:  `A longer description of the bind process.`, // TODO
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

func init() {
	RootCmd.AddCommand(bindCmd)
	// define flags for this command
	bindCmd.PersistentFlags().StringP("file-pattern", "f", "", "The output filename, can be a combination of literal values and patterns")
	bindCmd.PersistentFlags().IntP("jobs", "j", 1, "The number of concurrent transcoding process to run for conversion (don't exceed your cpu count)")
	bindCmd.PersistentFlags().StringP("output-directory", "o", "", "The output directory for the final directory, can be combination of absolute values and path patterns")
	bindCmd.PersistentFlags().StringP("path-pattern", "p", "", "The pattern for metadata picked up via paths")
	bindCmd.PersistentFlags().String("scratch-files-path", "", "The location to generate the scratch directory")
	bindCmd.PersistentFlags().StringP("source-files-path", "s", "", "The path to directory of source files (must match path-pattern for metadata to work)")
	bindCmd.PersistentFlags().Bool("verbose-transcode", false, "Enable output of all ffmpeg commands/operations")
	// Here you will define your flags and configuration settings.
	//bindCmd.MarkFlagRequired("source-files-path")
	bindCmd.MarkPersistentFlagRequired("source-files-path")

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// bindCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// bindCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// generateBindOpts configures and validates bind flags
func generateBindOpts(config *audiobooker.Config, flags *pflag.FlagSet) error {
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
	sourceFilesPath, err := flags.GetString("source-files-path")
	if err != nil {
		return err
	} else {
		config.SourceFilesPath = sourceFilesPath
	}

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
