package audiobooker

import (
	"errors"
	"fmt"
	"github.com/caarlos0/env/v6"
	log "github.com/sirupsen/logrus"
	"io/fs"
	"os"
	"path/filepath"
)

const (
	Aac  = ".aac"
	Flac = ".flac"
	Mp4  = ".mp4"
	M4a  = ".m4a"
	M4b  = ".m4b"
	Mp3  = ".mp3"
	Ogg  = ".ogg"
	Opus = ".opus"
)

var formats = []string{
	Aac,
	Flac,
	M4a,
	M4b,
	Mp4,
	Mp3,
	Ogg,
	Opus,
}

// Config application config data
type Config struct {
	// ChaptersFile file handler for chapters file
	ChaptersFile *os.File
	// DescriptionFilename optional filename for book description data
	DescriptionFilename string
	// ExternalChapters pull chapters from existing file
	ExternalChapters bool
	// Jobs number of concurrent transcode jobs to run
	Jobs int `yaml:"jobs" env:"JOBS"`
	// OutputFileDest path to output file TODO allow for custom file name
	OutputFileDest string `yaml:"output_file_dest" env:"OUTPUT_FILE_DEST"`
	// OutputFilePattern placeholder for output filename template
	OutputFilePattern string `yaml:"output_file_pattern" env:"OUTPUT_FILE_PATTERN"`
	// OutputPathPattern placeholder for output path template
	OutputPathPattern string `yaml:"output_path_pattern" env:"OUTPUT_PATH_PATTERN"`
	// PathPattern placeholder template string
	PathPattern string `yaml:"path_pattern" env:"PATH_PATTERN"`
	// ScratchFilesPath path to put scratch files
	ScratchFilesPath string `yaml:"scratch_files_path" env:"SCRATCH_FILES_PATH"`
	// SourceFilesPath wildcard glob of files to use as input TODO refine this into options
	SourceFilesPath string
	// TracksFile file handler for tracks to transcode/compile file
	TracksFile *os.File
	// VerboseTranscode show verbose output of ffmpeg commands
	VerboseTranscode bool

	// coverImage scraped cover image
	coverImage *string
	// descriptionFile file handler book description file
	descriptionFile *os.File
	// OutputFile filename of final book output file
	OutputFile string
	// OutputPath rendered path directories
	OutputPath string
	// preOutputFile transcoded and combined output file without metadata
	preOutputFile string
	// preOutputFilePath path to scratch output
	preOutputFilePath string
	// scratchDir Go generated directory for conversion files
	scratchDir string
	// sourceFiles list of the paths of the source files
	sourceFiles []string
	// transcodeFiles list of transcoded files
	transcodeFiles []string
}

// Parse sets defaults of parsed env/config file variables
func (c *Config) Parse() error {
	// Parse environment variables
	if err := env.Parse(c); err != nil {
		return err
	}

	return nil
}

// New provisions new Config object
func (c *Config) New() error {
	var err error

	// if no path for scratch files is defined, use current directory
	if c.ScratchFilesPath == "" {
		c.ScratchFilesPath = "."
	}

	// check if ENV jobs is greater than the value passed in

	// create temporary scratch directory
	c.scratchDir, err = os.MkdirTemp(c.ScratchFilesPath, "scratch-dir")
	if err != nil {
		return err
	}

	// create tracks list file
	c.TracksFile, err = os.Create(filepath.Join(c.scratchDir, "tracks.txt"))
	if err != nil {
		return err
	}

	// create chapters list file
	c.ChaptersFile, err = os.Create(filepath.Join(c.scratchDir, "chapters.ini"))
	if err != nil {
		return err
	}

	// setup final book output file
	c.OutputFile = filepath.Join(c.OutputFileDest, "book.m4b")

	// pre output file path
	c.preOutputFilePath = filepath.Join(c.scratchDir, "out.m4b")

	// compiled book location, pre-binding
	c.preOutputFile = filepath.Join(c.scratchDir, "pre-Bind.m4b")

	// build file list based on source path
	if err := c.gatherSourceFilesFromDir(); err != nil {
		return err
	}

	return nil
}

// Cleanup removes temporary scratch files
func (c *Config) Cleanup() error {
	if err := os.RemoveAll(c.scratchDir); err != nil {
		return err
	}
	return nil
}

// addToFileList adds track file to tracks file list for transcoding
func (c *Config) addToFileList(filename string) error {
	out := fmt.Sprintf("file '%s'\n", filename)
	if _, err := c.TracksFile.WriteString(out); err != nil {
		return err
	}

	return nil
}

// ParseConfig returns a config struct from a config file
func ParseConfig(filename string) (Config, error) {

	return Config{}, nil
}

// gatherSourceFilesFromDir scans directory for valid files
func (c *Config) gatherSourceFilesFromDir() error {
	err := filepath.WalkDir(c.SourceFilesPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Errorln("error parsing source file path", path)
			return err
		}
		if !d.IsDir() {
			if checkValidAudioFile(path) {
				log.Debugf("%s is valid, adding to list", path)
				c.sourceFiles = append(c.sourceFiles, path)
			}
			switch filepath.Base(path) {
			case "cover.jpg", "cover.png", "folder.jpg", "folder.png":
				c.coverImage = &path
			case "description.txt", "comment.txt", c.DescriptionFilename:
				log.Debugf("%s description file found!!", path)
				c.descriptionFile, err = os.Open(path)
				if err != nil {
					log.Errorf("could not open description file %s ", path)
					return err
				}
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

// checkValidAudioFile confirms that a file type is supported by looking up its extension
func checkValidAudioFile(filename string) bool {
	contains := func(ext string) bool {
		for _, format := range formats {
			if ext == format {
				return true
			}
		}
		return false
	}

	fileExt := filepath.Ext(filename)
	if contains(fileExt) {
		log.Debugf("%s has an extension of %s, approved!", filename, fileExt)
		return true
	}

	log.Warnf("%s has an extension of %s, which is not supported", filename, fileExt)
	return false
}

// SetOutputFilename sets the output filename based on metadata
func (c *Config) SetOutputFilename(book Book) error {
	var err error
	// TODO consider not requiring author
	if book.Author == "" || book.Title == "" {
		return errors.New("both author and title must be passed into this function")
	}
	// parse path pattern into output path
	c.OutputPath, err = OutputPathPattern(book, c.OutputPathPattern)
	if err != nil {
		return err
	}

	// parse file pattern into output file
	c.OutputFile = OutputFilePattern(book, c.OutputFilePattern)

	return nil
}
