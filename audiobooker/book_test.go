package audiobooker

import (
	"github.com/cslamar/mp4tag"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
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
	err := b1.GenerateStaticChapters(c1, 5, "")
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), 12, len(b1.Chapters))

	// four minute track test, should produce no chapters
	c2 := Config{
		transcodeFiles: []string{filepath.Join(TestDataRoot, "misc/4-min.m4a")},
	}
	b2 := Book{}
	err = b2.GenerateStaticChapters(c2, 5, "")
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), 0, len(b2.Chapters))

	// eight minute track test, should produce 2 chapters
	c3 := Config{
		transcodeFiles: []string{filepath.Join(TestDataRoot, "misc/8-min.m4a")},
	}
	b3 := Book{}
	err = b3.GenerateStaticChapters(c3, 5, "")
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), 2, len(b3.Chapters))

	// eight minute track with 3 minute chapter length test, should produce 3 chapters
	c4 := Config{
		transcodeFiles: []string{filepath.Join(TestDataRoot, "misc/8-min.m4a")},
	}
	b4 := Book{}
	err = b4.GenerateStaticChapters(c4, 3, "")
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), 3, len(b4.Chapters))

	// sixty minute single file test
	c5 := Config{
		sourceFiles: []string{filepath.Join(TestDataRoot, "misc", "60-min-book.m4b")},
	}
	b5 := Book{}
	err = b5.GenerateStaticChapters(c5, 5, filepath.Join(TestDataRoot, "misc", "60-min-book.m4b"))
	assert.Nil(suite.T(), err)
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

func (suite *BookTestSuite) TestEmbedDescription() {
	var err error

	// Set Config to test path, create new, and check for errors
	c1 := Config{
		OutputPath:       suite.ScratchPath,
		ScratchFilesPath: suite.ScratchPath,
		SourceFilesPath:  filepath.Join(TestDataRoot, "misc", "Test Author", "Test Book", "Title One"),
		VerboseTranscode: true,
	}
	err = c1.New()
	assert.Nil(suite.T(), err)

	// Create dummy book
	b1 := Book{
		Author: "Test Author",
		Title:  "Title One",
	}

	// generate metadata template
	err = b1.GenerateMetaTemplate(c1)
	// check for error
	assert.Nil(suite.T(), err)
	// check for existence of description
	assert.NotNil(suite.T(), b1.Description)

	// invalid description path
	// Set Config to test path, create new, and check for errors
	c2 := Config{
		OutputPath:       suite.ScratchPath,
		ScratchFilesPath: suite.ScratchPath,
		SourceFilesPath:  filepath.Join(TestDataRoot, "misc", "Test Author", "Test Book", "Title One"),
		VerboseTranscode: true,
	}
	err = c2.New()
	assert.Nil(suite.T(), err)
	// set empty description file and check for file pointer error
	c2.descriptionFile, err = os.Open(filepath.Join(TestDataRoot, "misc", "empty.txt"))
	assert.Nil(suite.T(), err)

	// Create dummy book
	b2 := Book{
		Author: "Test Author",
		Title:  "Title One",
	}

	// attempt to generate a description, it should not error even though it does not exist
	err = b2.GenerateMetaTemplate(c2)
	assert.Nil(suite.T(), err)

}

func (suite *BookTestSuite) TestWriteTags() {
	var err error
	// helper func to compare expected results
	var compareBookToFileTags = func(suite *BookTestSuite, book Book, filename *os.File) {
		// read temp file for newly written tags
		testFile, err := mp4tag.Open(filename.Name())
		if err != nil {
			log.Fatalln(err)
		}
		// get file tags
		tags, err := testFile.Read()
		if err != nil {
			log.Fatalln(err)
		}
		testFile.Close()

		// parse year to convert allow reference
		parsedYear := strconv.Itoa(tags.Year)

		// Check for values
		assert.Equal(suite.T(), book.Author, tags.Artist)
		assert.Equal(suite.T(), book.Date, &parsedYear)
		assert.Equal(suite.T(), book.Genre, &tags.Genre)
		assert.Equal(suite.T(), book.Narrator, &tags.Composer)
		assert.Equal(suite.T(), book.SortSlug, &tags.AlbumSort)
		assert.Equal(suite.T(), book.SortSlug, &tags.TitleSort)
		assert.Equal(suite.T(), book.Title, tags.Album)
	}

	srcTestFile := filepath.Join(TestDataRoot, "misc/tagging.m4b")
	// Generate temp file for read/write operations
	tmpFile, err := generateTestFile(suite.ScratchPath, srcTestFile)
	if err != nil {
		panic(err)
	}
	defer os.Remove(tmpFile.Name())

	// All tags
	author := "Carl von Clausewitz"
	title := "On War - Volume 1"
	seriesName := "On War"
	seriesPart := 1
	date := "1903"
	narrator := "Random Guy"
	genre := "History"

	b1 := Book{
		Author:     author,
		Title:      title,
		seriesPart: &seriesPart,
		seriesName: &seriesName,
		Date:       &date,
		Narrator:   &narrator,
		Genre:      &genre,
	}

	// test with all tags being set
	err = b1.WriteTags(tmpFile.Name())
	assert.Nil(suite.T(), err)
	compareBookToFileTags(suite, b1, tmpFile)

	// copy book 1 data for updated check
	b2 := b1
	title = "On War - Volume 2"
	seriesPart = 2
	b2.Title = title
	b2.seriesPart = &seriesPart

	err = b2.WriteTags(tmpFile.Name())
	assert.Nil(suite.T(), err)
	compareBookToFileTags(suite, b2, tmpFile)

	// Copy book 2 for final tests
	b3 := b2
	b3.seriesPart = nil

	err = b3.WriteTags(tmpFile.Name())
	assert.Nil(suite.T(), err)
	compareBookToFileTags(suite, b3, tmpFile)

	// invalid date case
	badDate := "not-a-date"
	b4 := Book{
		Date: &badDate,
	}
	err = b4.WriteTags(tmpFile.Name())
	assert.Nil(suite.T(), err)

	// file fail case
	b5 := Book{}
	err = b5.WriteTags("not-a-file")
	assert.Error(suite.T(), err)

}

func (suite *BookTestSuite) TestFormatDescription() {
	var err error

	// test file pointers
	invalidFile, _ := os.Open("/dev/asdf")
	emptyFile, _ := os.Open(filepath.Join(TestDataRoot, "misc", "empty.txt"))
	descriptionFile, _ := os.Open(filepath.Join(TestDataRoot, "misc", "description.txt"))

	// fail case invalid file
	c1 := Config{descriptionFile: invalidFile}
	b1 := Book{}
	err = b1.formatDescription(c1)
	assert.Error(suite.T(), err)

	// warning case empty file
	c2 := Config{descriptionFile: emptyFile}
	b2 := Book{}
	err = b2.formatDescription(c2)
	assert.Nil(suite.T(), err)

	// success case
	c3 := Config{descriptionFile: descriptionFile}
	b3 := Book{}
	err = b3.formatDescription(c3)
	assert.Nil(suite.T(), err)
}
