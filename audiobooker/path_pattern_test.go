package audiobooker

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"os"
	"path/filepath"
)

type PathPatternTestSuite struct {
	suite.Suite
	ScratchPath string
}

func (suite *PathPatternTestSuite) SetupSuite() {
	var err error
	// setup test file paths

	// Create scratch path for files
	suite.ScratchPath, err = os.MkdirTemp(UtScratchDirectory, "temp-path-pattern-")
	if err != nil {
		log.Errorln(err)
		return
	}
}

func (suite *PathPatternTestSuite) TearDownSuite() {
	if err := os.RemoveAll(suite.ScratchPath); err != nil {
		log.Errorln(err)
		return
	}
}

func (suite *PathPatternTestSuite) TestParsePathTags() {
	// 3 tag test
	pathPattern1 := fmt.Sprint(TestDataRoot + "/bind/chapter-by-file/%a/%s/%p/%t")
	pathTags1, err := ParsePathTags(filepath.Join(TestDataRoot, "/bind/chapter-by-file/Author Name/Series Name/1/Title Two"), pathPattern1)
	assert.Nil(suite.T(), err)
	assert.Len(suite.T(), pathTags1, 4)

	// malformed pattern test
	malformedPattern1 := "%a/%t/"
	malformedPatternTags1, err := ParsePathTags(filepath.Join(TestDataRoot, "/bind/chapter-by-tag/Test Author/Test Book"), malformedPattern1)
	assert.Nil(suite.T(), err, "should not error as malformed")
	assert.Len(suite.T(), malformedPatternTags1, 2)

	// mismatched path/tags combo
	mismatchPath1 := "tbi1/Some Author/Some Book"
	mismatchPattern1 := "tbi1/%y/%t"
	mismatchTags1, err := ParsePathTags(mismatchPath1, mismatchPattern1)
	assert.Error(suite.T(), err)
	assert.Len(suite.T(), mismatchTags1, 0)

	// 2 tag test
	pathPattern2 := fmt.Sprint(TestDataRoot + "/bind/chapter-by-tag/%a/%t")
	pathTags2, err := ParsePathTags(filepath.Join(TestDataRoot, "/bind/chapter-by-tag/Test Author/Test Book"), pathPattern2)
	assert.Nil(suite.T(), err)
	assert.Len(suite.T(), pathTags2, 2)

	// audio file test
	pathPattern3 := fmt.Sprint(TestDataRoot + "/bind/single-file-split/%a/%s/%p/%t/%f")
	pathTags3, err := ParsePathTags(filepath.Join(TestDataRoot, "/bind/single-file-split/Author Name/Series Name/1/Title Two/hour-test.m4a"), pathPattern3)
	assert.Nil(suite.T(), err)
	assert.Len(suite.T(), pathTags3, 5)

	// all tag test
	pathPattern4 := fmt.Sprint(TestDataRoot + "/bind/all-path-patterns/%g/%a/%s/%p/%y/%n/%t/%f")
	pathTags4, err := ParsePathTags(filepath.Join(TestDataRoot, "/bind/all-path-patterns/Thriller/Author Name/Series Name/1/1998/Scott Narrator/Title One/hour-test.m4a"), pathPattern4)
	assert.Nil(suite.T(), err)
	assert.Len(suite.T(), pathTags4, 8)
}

func (suite *PathPatternTestSuite) TestOutputPathPattern() {
	var err error

	author := "Some Author"
	title := "Some Title"
	genre := "Thriller"
	date := "1980"
	seriesName := "Some Series"
	seriesPart := 3
	narrator := "Joe"

	b1 := Book{
		Author:     author,
		Title:      title,
		Genre:      &genre,
		Date:       &date,
		Narrator:   &narrator,
		seriesName: &seriesName,
		seriesPart: &seriesPart,
	}

	// pure pattern output path
	outPath1, err := OutputPathPattern(b1, "%a/%g/%y/%s/%p/%n/%t")
	assert.Nil(suite.T(), err)
	log.Infoln(outPath1)
	assert.Equal(suite.T(), "Some Author/Thriller/1980/Some Series/3/Joe/Some Title", outPath1)

	// prefixed path
	outPath2, err := OutputPathPattern(b1, "output/%a/%s/%p/%t")
	assert.Nil(suite.T(), err)
	log.Infoln(outPath2)
	assert.Equal(suite.T(), "output/Some Author/Some Series/3/Some Title", outPath2)

	// mixed path
	outPath3, err := OutputPathPattern(b1, "output/%a/books/%s/%p/%t")
	assert.Nil(suite.T(), err)
	log.Infoln(outPath3)
	assert.Equal(suite.T(), "output/Some Author/books/Some Series/3/Some Title", outPath3)

	// static path, no patterns
	outPath4, err := OutputPathPattern(b1, "output/books")
	assert.Nil(suite.T(), err)
	log.Infoln(outPath4)
	assert.Equal(suite.T(), "output/books", outPath4)

	// parsed with relative path
	outPath5, err := OutputPathPattern(b1, "./%a/%g/%y/%s/%p/%n/%t")
	assert.Nil(suite.T(), err)
	log.Infoln(outPath5)
	assert.Equal(suite.T(), "./Some Author/Thriller/1980/Some Series/3/Joe/Some Title", outPath5)

	// invalid path
	_, err = OutputPathPattern(b1, "no path")
	assert.Error(suite.T(), err)
}

func (suite *PathPatternTestSuite) TestOutputFilePattern() {
	author := "Some Author"
	title := "Some Title"
	genre := "Thriller"
	date := "1980"
	seriesName := "Some Series"
	seriesPart := 3
	narrator := "Joe"

	b1 := Book{
		Author:     author,
		Title:      title,
		Genre:      &genre,
		Date:       &date,
		Narrator:   &narrator,
		seriesName: &seriesName,
		seriesPart: &seriesPart,
	}

	// just the title
	pattern1 := "%t"
	out1 := OutputFilePattern(b1, pattern1)
	assert.Equal(suite.T(), "Some Title.m4b", out1)

	// author name, space, title
	pattern2 := "%a %t"
	out2 := OutputFilePattern(b1, pattern2)
	assert.Equal(suite.T(), "Some Author Some Title.m4b", out2)

	// mix up patterns with characters
	pattern3 := "%a - %s %p - %t"
	out3 := OutputFilePattern(b1, pattern3)
	assert.Equal(suite.T(), "Some Author - Some Series 3 - Some Title.m4b", out3)

	// default title for no pattern
	out4 := OutputFilePattern(b1, "")
	assert.Equal(suite.T(), "Some Author - Some Title.m4b", out4)
}
