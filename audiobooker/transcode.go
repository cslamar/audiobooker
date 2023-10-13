package audiobooker

import (
	"context"
	"errors"
	"fmt"
	"github.com/cslamar/mp4tag"
	log "github.com/sirupsen/logrus"
	ffmpeg_go "github.com/u2takey/ffmpeg-go"
	"gopkg.in/vansante/go-ffprobe.v2"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Combine transcode and combines source files into m4a file
func Combine(config Config) error {
	combineCmd := ffmpeg_go.Input(config.TracksFile.Name(), ffmpeg_go.KwArgs{"f": "concat", "safe": 0}).
		//Output(config.preOutputFilePath, ffmpeg_go.KwArgs{"b:a": "64k", "acodec": "aac", "ac": 2, "vn": ""}).
		Output(config.preOutputFilePath, ffmpeg_go.KwArgs{"codec": "copy", "vn": "", "f": "mp4"}).
		OverWriteOutput()
	// check if verbose output should be shown
	if config.VerboseTranscode {
		combineCmd = combineCmd.ErrorToStdOut()
	}
	err := combineCmd.Run()
	if err != nil {
		return err
	}

	return nil
}

// SplitSingleFile splits single file into chunks for later transcoding
func SplitSingleFile(config *Config) error {
	if len(config.sourceFiles) > 1 {
		return errors.New("may only have one source file for pre-splitting for now")
	}

	srcFile := config.sourceFiles[0]

	// get file extension for later use
	fileExt := filepath.Ext(srcFile)
	if fileExt == "" {
		return errors.New("source file must be an audio file")
	}

	// create temporary directory for splitting single files
	splitDir := filepath.Join(config.scratchDir, "split")
	if err := os.Mkdir(splitDir, 0755); err != nil {
		return err
	}

	// calculate total time
	f, err := os.Open(srcFile)
	if err != nil {
		return err
	}
	defer f.Close()

	fileData, err := ffprobe.ProbeURL(context.Background(), f.Name())
	if err != nil {
		return err
	}

	// determine split length depending on file duration
	splitLength := 0
	if fileData.Format.Duration() >= (2 * time.Hour) {
		// if the file duration is over 2 hours, split every 120 minutes
		splitLength = 60 * 60 * 2
	} else if fileData.Format.Duration() < (10 * time.Minute) {
		// if the file duration is under 10 minutes, split every 5 minute
		splitLength = 5 * 60
	} else {
		// default to splitting every 10 minutes
		splitLength = 10 * 60
	}

	outFile := filepath.Join(splitDir, fmt.Sprintf("tc-part-%%03d%s", fileExt))

	splitCmd := ffmpeg_go.Input(srcFile).
		Output(outFile, ffmpeg_go.KwArgs{
			"f":                "segment",
			"segment_time":     splitLength,
			"c":                "copy",
			"reset_timestamps": 1,
		}).
		OverWriteOutput()
	if config.VerboseTranscode {
		splitCmd = splitCmd.ErrorToStdOut()
	}

	log.Infoln("Performing file split, this may take a minute depending on the size of the source file.")
	if err := splitCmd.Run(); err != nil {
		log.Errorln(err)
		return err
	}

	partFiles := make([]string, 0)
	err = filepath.WalkDir(splitDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			pFile, err := os.Open(path)
			if err != nil {
				return err
			}
			defer pFile.Close()
			pData, err := ffprobe.ProbeURL(context.Background(), pFile.Name())
			if err != nil {
				return err
			}
			if pData.Format.Duration() > (1 * time.Second) {
				partFiles = append(partFiles, path)
			}
		}
		return nil
	})

	if err != nil {
		return err
	}

	config.sourceFiles = partFiles

	return nil
}

// Bind apply metadata and output m4b file
func Bind(config Config, book Book) error {
	var err error
	var tempOutFile *os.File

	// Create a temporary output file for manipulation
	tempOutFile, err = os.CreateTemp(config.scratchDir, "tempOutFile")
	if err != nil {
		return err
	}

	if len(config.sourceFiles) == 1 {
		//config.preOutputFilePath = config.sourceFiles[0]
		// check if the source file is a directory or regular file, return single element if it's a directory
		config.preOutputFilePath, err = config.CheckForSourceFile(config.SourceFilesPath)
		if err != nil {
			return err
		}

		// Create full output path
		if err := os.MkdirAll(config.OutputPath, 0755); err != nil {
			return err
		}
	}

	// run general bind operation
	bindCmd := ffmpeg_go.Input(config.ChaptersFile.Name(), ffmpeg_go.KwArgs{"i": config.preOutputFilePath}).
		Output(tempOutFile.Name(), ffmpeg_go.KwArgs{"map_metadata": 1, "codec": "copy", "f": "mp4"}).
		OverWriteOutput()
	// check if verbose output should be shown
	if config.VerboseTranscode {
		bindCmd = bindCmd.ErrorToStdOut()
	}
	err = bindCmd.Run()
	if err != nil {
		log.Errorln("errored binding files:", err)
		return err
	}

	// if configured to use external chapters, do so now
	if config.ExternalChapters {
		// create new temp output file
		temp2, err := os.CreateTemp(config.scratchDir, "tempOutFile")
		if err != nil {
			return err
		}
		// map the chapters from the external chapters file
		embedCmd := ffmpeg_go.Input(filepath.Join(config.scratchDir, "extracted-chapters.ini"), ffmpeg_go.KwArgs{"i": tempOutFile.Name()}).
			Output(temp2.Name(), ffmpeg_go.KwArgs{"map_chapters": 1, "f": "mp4", "codec": "copy"}).
			OverWriteOutput()
		if config.VerboseTranscode {
			embedCmd = embedCmd.ErrorToStdOut()
		}
		err = embedCmd.Run()
		if err != nil {
			log.Errorln("error embedding external chapters")
			return err
		}
		// move the new temp file to overwrite the old one for continued processing
		if err := os.Rename(temp2.Name(), tempOutFile.Name()); err != nil {
			return err
		}
	}

	// if a cover image was found, add it here
	if config.coverImage != nil {
		// create new temp output file
		temp2, err := os.CreateTemp(config.scratchDir, "tempOutFile")
		if err != nil {
			return err
		}
		s1 := ffmpeg_go.Input(tempOutFile.Name())
		s2 := ffmpeg_go.Input(*config.coverImage)
		out := ffmpeg_go.Output([]*ffmpeg_go.Stream{s1, s2}, temp2.Name(), ffmpeg_go.KwArgs{"c": "copy", "disposition:v:0": "attached_pic", "f": "mp4"})
		log.Debugln(out.GetArgs())
		coverCmd := out.OverWriteOutput()
		// check if verbose output should be shown
		if config.VerboseTranscode {
			coverCmd = coverCmd.ErrorToStdOut()
		}
		err = coverCmd.Run()
		if err != nil {
			log.Errorln("error adding cover")
			return err
		}

		if err := os.Rename(temp2.Name(), tempOutFile.Name()); err != nil {
			return err
		}
	}

	// copy the bound book to the output directory
	if err := os.Rename(tempOutFile.Name(), filepath.Join(config.OutputPath, config.OutputFile)); err != nil {
		return err
	}

	// add sort tag if needed
	if book.SortSlug != nil {
		outputFile, err := mp4tag.Open(filepath.Join(config.OutputPath, config.OutputFile))
		if err != nil {
			log.Errorln("error opening file for tagging!")
			return err
		}
		defer outputFile.Close()
		sortTags := mp4tag.Tags{
			AlbumSort: *book.SortSlug,
			TitleSort: *book.SortSlug,
		}
		if err := outputFile.Write(&sortTags); err != nil {
			log.Errorln("error adding custom tags")
			return err
		}
	}

	return nil
}

// TranscodeSourceFiles runs concurrent transcode of source media into mp4 audio files for combination later
func TranscodeSourceFiles(config *Config) error {
	// define conversion holder
	type conversion struct {
		srcFile  string
		destFile string
	}

	// create the temporary output direction
	tmpDir, err := filepath.Abs(path.Join(config.scratchDir, "out"))
	if err != nil {
		return err
	}
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		return err
	}

	// Create a temporary slice for the conversions
	conversionFiles := make([]conversion, len(config.sourceFiles))
	// Create a slice of the output files
	config.transcodeFiles = make([]string, len(config.sourceFiles))

	// loop through the source files
	for idx := 0; idx < len(config.sourceFiles); idx++ {
		// massage names to new output type
		newFile := strings.TrimSuffix(config.sourceFiles[idx], path.Ext(config.sourceFiles[idx]))
		newFile += ".m4a"
		// create conversion entry
		conversionFiles[idx] = conversion{srcFile: config.sourceFiles[idx], destFile: path.Join(tmpDir, path.Base(newFile))}
		// write transcode file to tracks list
		if err := config.addToFileList(path.Join(tmpDir, path.Base(newFile))); err != nil {
			return err
		}
		// add new path to the config struct
		config.transcodeFiles[idx] = path.Join(tmpDir, path.Base(newFile))
	}

	var wg sync.WaitGroup
	// track completed transcode operations
	completedTranscode := 0

	// Create worker queue based on the number of jobs specified
	var queue = make(chan conversion, config.Jobs-1)
	log.Debugln("chan len:", len(queue))
	log.Debugln("chan cap:", cap(queue))

	// Loop through the provisioning of the workers
	wg.Add(config.Jobs)
	for idx := 0; idx < config.Jobs; idx++ {
		go func() {
			for {
				inputFile, ok := <-queue
				if !ok {
					wg.Done()
					log.Debugln("transcoding routine completed")
					return
				}
				log.Debugln("transcoding:", inputFile.srcFile)
				// Transcode file
				transcodeCmd := ffmpeg_go.Input(inputFile.srcFile).
					Output(inputFile.destFile, ffmpeg_go.KwArgs{"c:a": "aac", "vn": "", "f": "mp4"}).
					OverWriteOutput()
				// check if verbose output should be shown
				if config.VerboseTranscode {
					transcodeCmd = transcodeCmd.ErrorToStdOut()
				}
				err := transcodeCmd.Run()
				if err != nil {
					log.Errorln("failed to convert:", inputFile.srcFile)
					log.Errorln("Maybe I should bail?")
				}

				log.Debugln("converted to:", inputFile.destFile)
				completedTranscode++
				fmt.Printf("...%d%%", int((float64(completedTranscode) / float64(len(conversionFiles)) * 100)))
			}
		}()
	}

	// start time of transcoding operations
	start := time.Now()
	fmt.Printf("Starting the transcoding of %d files!\n", len(conversionFiles))

	// loop through the files list to convert
	for _, f := range conversionFiles {
		log.Debugln("sending:", f)
		queue <- f
		time.Sleep(10 * time.Millisecond)
	}

	// close channel and wait for completion
	close(queue)
	wg.Wait()

	fmt.Printf("\nFinished the transcode of all files\n")
	log.Debugln("transcoding took:", time.Now().Sub(start))

	return nil
}

// shouldTranscode returns a bool if transcoding is needed
func shouldTranscode(stream *ffprobe.Stream) bool {
	if stream.CodecName == "aac" {
		return false
	}
	return true
}

// TranscodeWithMarkers generates individual output files from a single file listed in MarkerPoint slice
func TranscodeWithMarkers(config *Config, points []MarkerPoint) error {
	log.Debugln("transcoding using marker points")
	if len(config.sourceFiles) > 1 {
		return errors.New("may only have one source file for pre-splitting for now")
	}

	// parse the source file
	srcFile := config.sourceFiles[0]
	sourceFileMeta, err := ffprobe.ProbeURL(context.Background(), srcFile)
	if err != nil {
		return err
	}

	// determine if codec transcoding is required
	var audioCodec string
	if shouldTranscode(sourceFileMeta.FirstAudioStream()) {
		// if it does, use aac
		audioCodec = "aac"
	} else {
		// if it's already aac, then just use copy
		audioCodec = "copy"
	}

	// create temporary directory for splitting single files
	splitDir := filepath.Join(config.scratchDir, "split")
	if err := os.Mkdir(splitDir, 0755); err != nil {
		return err
	}

	// start transcode from MarkerPoints
	startingPoint := float64(0)
	for idx := 0; idx < len(points); idx++ {
		endMark := fmt.Sprintf("%f", points[idx].ParseEnd())
		outFile := filepath.Join(splitDir, fmt.Sprintf("Track-%03d.aac", idx+1))
		cmd := ffmpeg_go.Input(srcFile).
			Output(outFile, ffmpeg_go.KwArgs{
				"ss":  fmt.Sprintf("%f", startingPoint),
				"c:a": audioCodec,
				"to":  endMark,
			}).OverWriteOutput().ErrorToStdOut()
		if err := cmd.Run(); err != nil {
			return err
		}
		// if the command ran successfully, update the starting point to the end of the previous
		startingPoint = points[idx].ParseEnd()
	}

	return nil
}
