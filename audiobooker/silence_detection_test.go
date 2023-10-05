package audiobooker

import (
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"os"
	"path/filepath"
)

type SilenceDetectionTestSuite struct {
	suite.Suite
	ScratchPath string
}

func (suite *SilenceDetectionTestSuite) SetupSuite() {
	var err error
	// setup test file paths

	// Create scratch path for files
	suite.ScratchPath, err = os.MkdirTemp(UtScratchDirectory, "temp-transcoding-")
	if err != nil {
		log.Errorln(err)
		return
	}

}

func (suite *SilenceDetectionTestSuite) TearDownSuite() {
	if err := os.RemoveAll(suite.ScratchPath); err != nil {
		log.Errorln(err)
		return
	}
}

func (suite *SilenceDetectionTestSuite) TestGenerateVolMarkers() {
	var err error
	file := filepath.Join(TestDataRoot, "misc", "3and5-sec-silence.mp3")

	// Find the 3 and 5 second silence in track
	points1, err := GenerateVolMarkers(file, 3, 30)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), 2, len(points1))

	// Find the 5 seconds silence in the track
	points2, err := GenerateVolMarkers(file, 4.5, 30)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), 1, len(points2))

	// Find no silence point in track
	points3, err := GenerateVolMarkers(file, 10, 30)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), 0, len(points3))
}
