package audiobooker

import (
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"os"
	"path/filepath"
	"strings"
)

type ConfigTestSuite struct {
	suite.Suite
	ScratchPath  string
	TestConfig   Config
	TestDataPath string
	TestPath     string
}

func (suite *ConfigTestSuite) SetupSuite() {
	var err error

	// Create temp directory
	suite.ScratchPath, err = os.MkdirTemp(UtScratchDirectory, "temp-test-config-")
	if err != nil {
		log.Errorln(err)
		return
	}

	suite.TestDataPath = filepath.Join(suite.ScratchPath, "1984/Some Author/Some Title")

	// bootstrap some dummy files
	if err := os.MkdirAll(suite.TestDataPath, 0755); err != nil {
		log.Errorln(err)
		return
	}

	// generate dummy chapter files
	for _, filename := range TestChapterFiles {
		if _, err := os.Create(filepath.Join(suite.TestDataPath, filename)); err != nil {
			log.Errorln(err)
		}
	}

	// generate dummy cover image files
	for _, filename := range TestCoverFiles {
		if _, err := os.Create(filepath.Join(suite.TestDataPath, filename)); err != nil {
			log.Errorln(err)
		}
	}

	suite.TestConfig = Config{
		SourceFilesPath: suite.TestDataPath,
	}

	// set env variable to test parse against
	if err := os.Setenv("PATH_PATTERN", "input/%a/%t"); err != nil {
		log.Errorln(err)
	}
}

func (suite *ConfigTestSuite) TearDownSuite() {
	// clean up env variables
	os.Unsetenv("PATH_PATTERN")
	// clean up scratch directory
	if err := os.RemoveAll(suite.ScratchPath); err != nil {
		log.Errorln(err)
		return
	}
}

func (suite *ConfigTestSuite) TestNewAndCleanup() {
	var err error
	config := Config{
		SourceFilesPath: suite.ScratchPath,
	}

	// test config New
	err = config.New()
	assert.Nil(suite.T(), err)

	// test clean up
	err = config.Cleanup()
	assert.Nil(suite.T(), err)
	// if clean up fails, force it
	if err != nil {
		log.Errorln("encountered an error cleaning up! Forcing!", err)
		os.RemoveAll(config.scratchDir)
	}
}

func (suite *ConfigTestSuite) TestParse() {
	config := Config{}
	err := config.Parse()
	assert.Nil(suite.T(), err)

	assert.Equal(suite.T(), "input/%a/%t", config.PathPattern)
}

func (suite *ConfigTestSuite) TestParseConfig() {
	_, err := ParseConfig("config.yaml")
	assert.Nil(suite.T(), err)
}

func (suite *ConfigTestSuite) TestGatherSourceFilesFromDir() {
	// test failure
	config := Config{}
	err := config.gatherSourceFilesFromDir()
	assert.Error(suite.T(), err)
}

func (suite *ConfigTestSuite) TestAddToFileList() {
	config := Config{
		ScratchFilesPath: suite.ScratchPath,
		SourceFilesPath:  suite.TestDataPath,
	}

	err := config.New()
	assert.Nil(suite.T(), err)
	defer config.Cleanup()

	err = config.addToFileList(TestChapterFiles[0])
	assert.Nil(suite.T(), err)

	data, err := os.ReadFile(config.TracksFile.Name())
	assert.Nil(suite.T(), err)
	hasFile := strings.Contains(string(data), TestChapterFiles[0])
	assert.True(suite.T(), hasFile)
}

func (suite *ConfigTestSuite) TestSetOutputFilename() {
	var err error
	releaseDate := "1980"
	b1 := Book{
		Author: "Some Author",
		Title:  "Some Title",
		Date:   &releaseDate,
	}

	// Test default output file pattern
	c1 := Config{
		OutputPathPattern: filepath.Join(suite.ScratchPath, "output"),
	}
	err = c1.SetOutputFilename(b1)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), filepath.Join(suite.ScratchPath, "output/Some Author - Some Title.m4b"), filepath.Join(c1.OutputPath, c1.OutputFile))

	// Test custom formatted output path and filename
	c2 := Config{
		OutputPathPattern: filepath.Join(suite.ScratchPath, "output/%a/%y/"),
		OutputFilePattern: "%y: %a-%t",
	}
	err = c2.SetOutputFilename(b1)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), filepath.Join(suite.ScratchPath, "output/Some Author/1980/1980: Some Author-Some Title.m4b"), filepath.Join(c2.OutputPath, c2.OutputFile))
}

func (suite *ConfigTestSuite) TestCheckForSourceFile() {
	var err error
	testBookDir := filepath.Join(TestDataRoot, "misc")
	testBookFile := filepath.Join(testBookDir, "60-min-book.m4b")

	// bad path test
	badPath := "./asdf"
	c1 := Config{}
	_, err = c1.CheckForSourceFile(badPath)
	assert.Error(suite.T(), err)

	// regular file test
	c2 := Config{}
	regularFileSuccess, err := c2.CheckForSourceFile(testBookFile)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), filepath.Join(testBookFile), regularFileSuccess)

	// directory files successful test
	c3 := Config{
		sourceFiles: []string{"60-min-book.m4b"},
	}
	dirFilesSuccess, err := c3.CheckForSourceFile(testBookDir)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), dirFilesSuccess, "60-min-book.m4b")

	// directory files too many files test
	c4 := Config{
		sourceFiles: []string{"file1.m4b", "file2.m4b"},
	}
	_, err = c4.CheckForSourceFile(testBookDir)
	assert.Error(suite.T(), err)

	// directory files no files test
	c5 := Config{}
	_, err = c5.CheckForSourceFile(testBookDir)
	assert.Error(suite.T(), err)
}
