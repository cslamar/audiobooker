package audiobooker

import (
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"os"
	"path/filepath"
)

type TrackTestSuite struct {
	suite.Suite
	ScratchPath string
}

func (suite *TrackTestSuite) SetupSuite() {
	var err error
	suite.ScratchPath, err = os.MkdirTemp(UtScratchDirectory, "track-tests-")
	if err != nil {
		log.Errorln(err)
		return
	}
}

func (suite *TrackTestSuite) TearDownSuite() {
	os.RemoveAll(suite.ScratchPath)
}

func (suite *TrackTestSuite) TestParse() {
	var err error
	// failed file action
	t1 := TrackFile{}
	err = t1.Parse("no-file.txt")
	assert.Error(suite.T(), err)

	// failed probe action
	testFile, err := os.Create(filepath.Join(suite.ScratchPath, "test.mp3"))
	if err != nil {
		log.Errorln(err)
		return
	}
	t2 := TrackFile{File: testFile}
	err = t2.Parse(testFile.Name())
	assert.Error(suite.T(), err)

	// successful probe action
	t3 := TrackFile{}
	err = t3.Parse(filepath.Join(TestDataRoot, "misc/60-min.m4a"))
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), int64(3600000), t3.LengthMs)
}
