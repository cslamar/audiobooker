package audiobooker

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gopkg.in/vansante/go-ffprobe.v2"
	"path/filepath"
)

type ChapterSuite struct {
	suite.Suite
	TestChapter Chapter
}

func (suite *ChapterSuite) SetupSuite() {

}

func (suite *ChapterSuite) TearDownSuite() {

}

func (suite *ChapterSuite) TestStampTimes() {
	c := Chapter{}
	c.LengthMs = int64(1000)
	endMs := c.StampTimes(0)
	assert.Equal(suite.T(), int64(1000), endMs)
	assert.Equal(suite.T(), int64(1000), c.EndMs)
}

func (suite *ChapterSuite) TestCompile() {
	c := Chapter{Tracks: []TrackFile{{LengthMs: int64(1000)}}}
	c.Compile()

	assert.Equal(suite.T(), int64(1000), c.LengthMs)
}

func (suite *ChapterSuite) TestParseFromCueTag() {
	var err error
	c1 := Config{
		sourceFiles: []string{filepath.Join(TestDataRoot, "misc/cue.mp3")},
	}

	// parse cue sheet from tag
	cues1, err := parseFromCueTag(c1)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), 50, len(cues1))

	// convert the first entry of cue to Chapter
	chapter1 := cues1[0].toChapter()
	assert.Equal(suite.T(), 1, chapter1.Number)

	// compile all cues as Chapters
	chapters := make([]*Chapter, len(cues1))
	for idx := 0; idx < len(chapters); idx++ {
		c := cues1[idx].toChapter()
		chapters[idx] = &c
	}
	// check for chapters
	assert.Equal(suite.T(), 50, len(chapters))

	// check for chapter 1 length
	assert.Equal(suite.T(), int64(20000), chapters[0].LengthMs)
	assert.Equal(suite.T(), int64(0), chapters[0].StartMs)
	assert.Equal(suite.T(), int64(20000), chapters[0].EndMs)

	// check for chapter 2 length
	assert.Equal(suite.T(), int64(823000), chapters[1].LengthMs)
	assert.Equal(suite.T(), int64(20000), chapters[1].StartMs)
	assert.Equal(suite.T(), int64(843000), chapters[1].EndMs)

	// parse file with no cue tag
	c2 := Config{
		sourceFiles: []string{filepath.Join(TestDataRoot, "misc/tagged.mp3")},
	}
	_, err = parseFromCueTag(c2)
	// should return a tag not found error
	assert.Equal(suite.T(), ffprobe.ErrTagNotFound, err)
}
