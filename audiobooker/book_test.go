package audiobooker

import (
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"io/fs"
	"os"
	"path/filepath"
)

type BookTestSuite struct {
	suite.Suite
	ScratchPath string
	Config      Config
}

func (suite *BookTestSuite) SetupSuite() {
	var err error
	// Create temp directory
	suite.ScratchPath, err = os.MkdirTemp(UtScratchDirectory, "temp-test-book-")
	if err != nil {
		log.Errorln(err)
		return
	}
	scanTestFiles, err := func(dir string) ([]string, error) {
		files := make([]string, 0)
		err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !d.IsDir() {
				if checkValidAudioFile(path) {
					log.Debugf("%s is valid", path)
					files = append(files, path)
				}
			}

			return nil
		})
		if err != nil {
			return nil, err
		}

		return files, nil
	}(filepath.Join(TestDataRoot, "bind/chapter-by-tag/Test Author/Test Book"))
	if err != nil {
		log.Errorln(err)
		return
	}
	suite.Config.transcodeFiles = scanTestFiles
}

func (suite *BookTestSuite) TearDownSuite() {
	if err := os.RemoveAll(suite.ScratchPath); err != nil {
		log.Errorln(err)
	}
}

func (suite *BookTestSuite) TestParseFromPattern() {
	contentPath := "1234/Author Name/Narrator Name/Genre Name/Series Name/1/The Book Title"
	contentPattern := "%y/%a/%n/%g/%s/%p/%t"

	parsedTags, err := ParsePathTags(contentPath, contentPattern)
	assert.Nil(suite.T(), err)

	book := Book{}
	book.ParseFromPattern(parsedTags)

	assert.Equal(suite.T(), "1234", *book.Date)
	assert.Equal(suite.T(), "Author Name", book.Author)
	assert.Equal(suite.T(), "Narrator Name", *book.Narrator)
	assert.Equal(suite.T(), "Genre Name", *book.Genre)
	assert.Equal(suite.T(), "The Book Title", book.Title)
	assert.Equal(suite.T(), "Series Name", *book.seriesName)
	assert.Equal(suite.T(), 1, *book.seriesPart)
}

func (suite *BookTestSuite) TestCalcChapterTimes() {
	chapter := Chapter{
		LengthMs: int64(1000),
	}
	book := Book{
		Chapters: []*Chapter{&chapter},
	}
	book.CalcChapterTimes()
	assert.Equal(suite.T(), int64(1000), chapter.EndMs)
}

func (suite *BookTestSuite) TestGenerateMetaTemplate() {
	var err error
	book := Book{
		Author: "Author Name 1",
		Title:  "Book Title 1",
		Chapters: []*Chapter{
			{
				Title:   "Chapter 1",
				StartMs: int64(0),
				EndMs:   int64(1000),
			},
		},
	}

	// fail execute case
	err = book.GenerateMetaTemplate(Config{})
	assert.Error(suite.T(), err)

	// success case
	config := Config{
		OutputFileDest:   suite.ScratchPath,
		SourceFilesPath:  suite.ScratchPath,
		ScratchFilesPath: suite.ScratchPath,
	}
	err = config.New()
	assert.Nil(suite.T(), err)

	err = book.GenerateMetaTemplate(config)
	assert.Nil(suite.T(), err)

	fileInfo, err := config.ChaptersFile.Stat()
	assert.Nil(suite.T(), err)
	assert.Greater(suite.T(), fileInfo.Size(), int64(0))

	err = config.Cleanup()
	assert.Nil(suite.T(), err)

	// generate with sort slug
	c2 := Config{
		OutputFileDest:   suite.ScratchPath,
		SourceFilesPath:  suite.ScratchPath,
		ScratchFilesPath: suite.ScratchPath,
	}
	err = c2.New()
	assert.Nil(suite.T(), err)

	seriesName := "Series Name"
	seriesPart := 2

	b2 := &Book{
		Author: "Author Name 2",
		Title:  "Book Title 2",
		Chapters: []*Chapter{
			{
				Title:   "Chapter 1",
				StartMs: int64(0),
				EndMs:   int64(1000),
			},
		},
		seriesName: &seriesName,
		seriesPart: &seriesPart,
	}

	err = b2.GenerateMetaTemplate(c2)
	assert.Nil(suite.T(), err)

	fInfo2, err := c2.ChaptersFile.Stat()
	assert.Nil(suite.T(), err)
	assert.Greater(suite.T(), fInfo2.Size(), int64(0))
	assert.Equal(suite.T(), "Series Name - 2 - Book Title 2", *b2.SortSlug)

	// generate with invalid sort value
	c3 := Config{
		OutputFileDest:   suite.ScratchPath,
		SourceFilesPath:  suite.ScratchPath,
		ScratchFilesPath: suite.ScratchPath,
	}
	err = c3.New()
	assert.Nil(suite.T(), err)

	b3 := &Book{
		Author: "Author Name 3",
		Title:  "Book Title 3",
		Chapters: []*Chapter{
			{
				Title:   "Chapter 1",
				StartMs: int64(0),
				EndMs:   int64(1000),
			},
		},
		seriesName: &seriesName,
	}

	err = b3.GenerateMetaTemplate(c3)
	assert.Nil(suite.T(), err)

	fInfo3, err := c3.ChaptersFile.Stat()
	assert.Nil(suite.T(), err)
	assert.Greater(suite.T(), fInfo3.Size(), int64(0))
	assert.Nil(suite.T(), b3.SortSlug)
}

func (suite *BookTestSuite) TestChapterByFile() {
	var err error

	// use file names as chapter names
	b1 := Book{}
	err = b1.ChapterByFile(suite.Config, true, false)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), 3, len(b1.Chapters))

	// use file tag's title as chapter name
	b2 := Book{}
	err = b2.ChapterByFile(suite.Config, false, true)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), 3, len(b2.Chapters))

	// use file order as chapter name
	b3 := Book{}
	err = b3.ChapterByFile(suite.Config, false, false)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), 3, len(b3.Chapters))

	// error parsing track file
	c4 := Config{transcodeFiles: []string{"no-file.mp3"}}
	b4 := Book{}
	err = b4.ChapterByFile(c4, false, false)
	assert.Error(suite.T(), err)

	// error parsing file tag
	c5 := Config{transcodeFiles: []string{filepath.Join(TestDataRoot, "files/no-tag.mp3")}}
	b5 := Book{}
	err = b5.ChapterByFile(c5, false, true)
	assert.Error(suite.T(), err)
}

func (suite *BookTestSuite) TestParseToChapters() {
	var err error

	// successful operation
	b1 := Book{}
	err = b1.ParseToChapters(suite.Config)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), 3, len(b1.Chapters))

	// error no file
	c2 := Config{
		transcodeFiles: []string{"no-file.mp3"},
	}
	b2 := Book{}
	err = b2.ParseToChapters(c2)
	assert.Error(suite.T(), err)

	// error parsing file tag
	c3 := Config{transcodeFiles: []string{filepath.Join(TestDataRoot, "files/no-tag.mp3")}}
	b3 := Book{}
	err = b3.ParseToChapters(c3)
	assert.Error(suite.T(), err)
}

func (suite *BookTestSuite) TestGenerateStaticChapters() {
	// hour long track test should produce 12 chapters
	c1 := Config{
		transcodeFiles: []string{filepath.Join(TestDataRoot, "misc/60-min.m4a")},
	}
	b1 := Book{}
	err := b1.GenerateStaticChapters(c1, 5)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), 12, len(b1.Chapters))

	// four minute track test, should produce no chapters
	c2 := Config{
		transcodeFiles: []string{filepath.Join(TestDataRoot, "misc/4-min.m4a")},
	}
	b2 := Book{}
	err = b2.GenerateStaticChapters(c2, 5)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), 0, len(b2.Chapters))

	// eight minute track test, should produce 2 chapters
	c3 := Config{
		transcodeFiles: []string{filepath.Join(TestDataRoot, "misc/8-min.m4a")},
	}
	b3 := Book{}
	err = b3.GenerateStaticChapters(c3, 5)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), 2, len(b3.Chapters))

	// eight minute track with 3 minute chapter length test, should produce 3 chapters
	c4 := Config{
		transcodeFiles: []string{filepath.Join(TestDataRoot, "misc/8-min.m4a")},
	}
	b4 := Book{}
	err = b4.GenerateStaticChapters(c4, 3)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), 3, len(b4.Chapters))
}

func (suite *BookTestSuite) TestBindEmbeddedChapters() {
	var err error

	c1 := Config{
		OutputPath:       suite.ScratchPath,
		ScratchFilesPath: suite.ScratchPath,
		SourceFilesPath:  filepath.Join(TestDataRoot, "misc/embedded-chapters.opus"),
		VerboseTranscode: true,
	}
	err = c1.New()
	assert.Nil(suite.T(), err)

	imgFile := filepath.Join(TestDataRoot, "misc/cover.jpg")
	c1.ExternalChapters = true
	c1.preOutputFilePath = filepath.Join(TestDataRoot, "misc/10-min.m4a")
	c1.coverImage = &imgFile

	seriesName := "some series"
	b1 := Book{
		Author:   "Some Author",
		Title:    "Some Title",
		SortSlug: &seriesName,
	}

	// generate book meta
	err = b1.GenerateMetaTemplate(c1)
	assert.Nil(suite.T(), err)

	// pull embed
	err = b1.ExtractChapters(c1)
	assert.Nil(suite.T(), err)

	// bind
	err = Bind(c1, b1)
	assert.Nil(suite.T(), err)
}
