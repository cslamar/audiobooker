package audiobooker

import (
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"io"
	"os"
	"path/filepath"
)

type TranscodeTestSuite struct {
	suite.Suite
	ScratchPath string
}

func (suite *TranscodeTestSuite) SetupSuite() {
	var err error
	// setup test file paths

	// Create scratch path for files
	suite.ScratchPath, err = os.MkdirTemp(UtScratchDirectory, "temp-transcoding-")
	if err != nil {
		log.Errorln(err)
		return
	}

}

func (suite *TranscodeTestSuite) TearDownSuite() {
	if err := os.RemoveAll(suite.ScratchPath); err != nil {
		log.Errorln(err)
		return
	}
}

func (suite *TranscodeTestSuite) TestTranscodeSourceFiles() {
	var err error

	// successful transcode
	c1 := Config{
		VerboseTranscode: true,
	}
	c1.ScratchFilesPath = suite.ScratchPath
	c1.SourceFilesPath = TestFilesMp3
	c1.Jobs = 3
	err = c1.New()
	assert.Nil(suite.T(), err)
	err = TranscodeSourceFiles(&c1)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), 2, len(c1.transcodeFiles))
	f1, err := os.Stat(c1.transcodeFiles[0])
	assert.Nil(suite.T(), err)
	assert.Greater(suite.T(), f1.Size(), int64(0))
	f2, err := os.Stat(c1.transcodeFiles[1])
	assert.Nil(suite.T(), err)
	assert.Greater(suite.T(), f2.Size(), int64(0))

	// successful combine
	err = Combine(c1)
	assert.Nil(suite.T(), err)
	f3, err := os.Stat(c1.preOutputFilePath)
	assert.Greater(suite.T(), f3.Size(), int64(0))
}

func (suite *TranscodeTestSuite) TestBind() {
	// Copy dummy audio file for testing
	srcAudioFile, err := os.Open(filepath.Join(TestDataRoot, "misc/60-min.m4a"))
	if err != nil {
		log.Errorln(err)
		panic(err)
	}
	testAudioFile, err := os.Create(filepath.Join(suite.ScratchPath, "pre-audio-file.m4a"))
	if err != nil {
		log.Errorln(err)
		return
	}
	_, err = io.Copy(testAudioFile, srcAudioFile)
	if err != nil {
		log.Errorln(err)
		panic(err)
	}

	coverImgPath := filepath.Join(TestDataRoot, "misc/cover.jpg")
	chapterFile, err := os.Open(filepath.Join(TestDataRoot, "misc/chapters.ini"))
	if err != nil {
		log.Errorln(err)
		panic(err)
	}

	slug := "series name - part 1 - title"
	book := Book{
		SortSlug: &slug,
	}

	c1 := Config{
		ChaptersFile:      chapterFile,
		coverImage:        &coverImgPath,
		OutputFile:        filepath.Join(suite.ScratchPath, "bound-book.m4b"),
		preOutputFile:     filepath.Join(suite.ScratchPath, "output.m4b"),
		preOutputFilePath: testAudioFile.Name(),
		VerboseTranscode:  true,
	}

	err = Bind(c1, book)
	assert.Nil(suite.T(), err)

	c2 := Config{
		ChaptersFile:      chapterFile,
		OutputFile:        filepath.Join(suite.ScratchPath, "bound-book-2.m4b"),
		preOutputFile:     filepath.Join(suite.ScratchPath, "output-2.m4b"),
		preOutputFilePath: testAudioFile.Name(),
	}
	err = Bind(c2, book)
	assert.Nil(suite.T(), err)
}

func (suite *TranscodeTestSuite) TestSplitSingleFile() {
	var err error

	cleanSplitDir := func() {
		// clean up
		err = os.RemoveAll(filepath.Join(suite.ScratchPath, "split"))
		if err != nil {
			log.Errorln(err)
		}
	}

	// successful split on exact hour file
	c1 := Config{
		scratchDir:  suite.ScratchPath,
		sourceFiles: []string{filepath.Join(TestDataRoot, "misc/60-min.m4a")},
	}

	err = SplitSingleFile(&c1)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), 6, len(c1.sourceFiles))

	cleanSplitDir()

	// successful on four minute file
	c2 := Config{
		scratchDir:  suite.ScratchPath,
		sourceFiles: []string{filepath.Join(TestDataRoot, "misc/8-min.m4a")},
	}

	err = SplitSingleFile(&c2)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), 2, len(c2.sourceFiles))

	// clean up
	cleanSplitDir()

	// failure on extra files
	c3 := Config{
		scratchDir:  suite.ScratchPath,
		sourceFiles: []string{filepath.Join(TestDataRoot, "misc/8-min.m4a"), "no-file.mp3"},
	}

	err = SplitSingleFile(&c3)
	assert.Error(suite.T(), err)

	// success on splitting multi-hour-file into even parts
	c4 := Config{
		scratchDir:       suite.ScratchPath,
		sourceFiles:      []string{filepath.Join(TestDataRoot, "misc/2-hour.m4a")},
		VerboseTranscode: true,
	}

	err = SplitSingleFile(&c4)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), 1, len(c4.sourceFiles))

	// clean up
	cleanSplitDir()

	// success on splitting multi-hour-file into odd parts
	c5 := Config{
		scratchDir:  suite.ScratchPath,
		sourceFiles: []string{filepath.Join(TestDataRoot, "misc/3-hour.m4a")},
	}

	err = SplitSingleFile(&c5)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), 2, len(c5.sourceFiles))

	// clean up
	cleanSplitDir()
}
