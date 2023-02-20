package audiobooker

import (
	"github.com/stretchr/testify/suite"
	"os"
	"path/filepath"
	"testing"
)

const (
// TestPath1 = "1980/Robert Ludlum/The Bourne Identity"
// TestDataRoot    = "../test-data"
// TestFilesAac    = "../test-data/files/transcoding/aac"
// TestFilesMp3    = "../test-data/files/transcoding/mp3"
// TestCoverImage  = "cover.jpg"
// TestFolderImage = "folder.jpg"
)

var TestChapterFiles = [7]string{
	"Chapter 1.aac",
	"Chapter 2.flac",
	"Chapter 3.mp4",
	"Chapter 4.m4a",
	"Chapter 5.m4b",
	"Chapter 6.mp3",
	"Chapter 7.ogg",
}

// Test Macros
var (
	TestDataRoot string
	TestFilesMp3 string
	TestFilesAac string
	TestPath1    string
)
var TestCoverFiles = [4]string{"cover.jpg", "cover.png", "folder.jpg", "folder.png"}
var UtScratchDirectory string

func setTestMacros() {
	TestDataRoot = os.Getenv("TEST_DATA_SRC")
	TestFilesMp3 = filepath.Join(TestDataRoot, "transcoding/mp3")
	TestFilesAac = filepath.Join(TestDataRoot, "transcoding/aac")
	TestPath1 = filepath.Join(TestDataRoot, "")
}

func TestSuite(t *testing.T) {
	var err error
	// check for alternative path for scratch files
	scratchPath := os.Getenv("TEST_SCRATCH_PATH")
	if scratchPath == "" {
		scratchPath = "."
	}
	UtScratchDirectory, err = os.MkdirTemp("", "ut-run-")
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := os.RemoveAll(UtScratchDirectory); err != nil {
			panic(err)
		}
	}()

	setTestMacros()

	suite.Run(t, new(BookTestSuite))
	suite.Run(t, new(ChapterSuite))
	suite.Run(t, new(ConfigTestSuite))
	suite.Run(t, new(PathPatternTestSuite))
	suite.Run(t, new(TrackTestSuite))
	suite.Run(t, new(TranscodeTestSuite))
}
