package audiobooker

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"github.com/dhowden/tag"
	log "github.com/sirupsen/logrus"
	ffmpeg_go "github.com/u2takey/ffmpeg-go"
	"gopkg.in/vansante/go-ffprobe.v2"
	"html/template"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

//go:embed metadata.ini.tmpl
var metadataTemplate embed.FS

// Book top level construct of book
type Book struct {
	Author   string
	Chapters []*Chapter
	Date     *string
	Genre    *string
	Narrator *string
	SortSlug *string
	Title    string

	seriesName *string
	seriesPart *int
}

// GenerateMetaTemplate writes out the compiled metadata template for use when compiling to m4b
func (b *Book) GenerateMetaTemplate(config Config) error {
	// if series name and part are present, generate Sort property
	if b.seriesName != nil && b.seriesPart != nil {
		b.SortSlug = new(string)
		*b.SortSlug = fmt.Sprintf("%s - %d - %s", *b.seriesName, *b.seriesPart, b.Title)
	}

	tmpl, err := template.ParseFS(metadataTemplate, "metadata.ini.tmpl")
	if err != nil {
		return err
	}

	if err := tmpl.Execute(config.ChaptersFile, b); err != nil {
		return err
	}

	return nil
}

// CalcChapterTimes calculates the duration of the chapter
func (b *Book) CalcChapterTimes() {
	startTime := int64(0)

	for _, chapter := range b.Chapters {
		startTime = chapter.StampTimes(startTime)
		startTime++
	}
}

// ParseFromPattern parses map of tags generated from path into attributes
func (b *Book) ParseFromPattern(tags map[string]string) {
	for k, v := range tags {
		switch k {
		case "author":
			b.Author = v
		case "genre":
			genre := v
			b.Genre = &genre
		case "narrator":
			narrator := v
			b.Narrator = &narrator
		case "release_date":
			date := v
			b.Date = &date
		case "series":
			series := v
			b.seriesName = &series
		case "series_part":
			seriesPart, _ := strconv.Atoi(v)
			b.seriesPart = &seriesPart
		case "title":
			b.Title = v
		default:
			log.Debugf("no attribute matches %s, ignoring.", k)
		}
	}
}

// GenerateStaticChapters creates Chapter objects based on specified length
func (b *Book) GenerateStaticChapters(config Config, chapterLengthMin int) error {
	totalMs := int64(0)
	for _, filename := range config.transcodeFiles {
		f, err := os.Open(filename)
		if err != nil {
			return err
		}
		fileData, err := ffprobe.ProbeURL(context.Background(), f.Name())
		totalMs += fileData.Format.Duration().Milliseconds()
		f.Close()
	}

	extraChapterLen := int64(0)
	numChapters := totalMs / int64(chapterLengthMin*60*1000)

	if numChapters == 0 {
		log.Debugf("book %s - %s has less than one chapter length's worth of audio!  Creating no chapter.", b.Author, b.Title)
	} else if numChapters == 1 {
		// if there's one chapter, check if there's additional trailing audio
		log.Debugln("checking for extra audio after first track")
		extraChapterLen = totalMs - int64(chapterLengthMin*60*1000)
		log.Debugf("found %d remaining after first track", extraChapterLen)
	} else if (totalMs % numChapters) != 0 {
		log.Debugln("found extra audio after last chapter")
		extraChapterLen = totalMs % numChapters
		log.Debugf("found %d remaining after last track", extraChapterLen)
	}

	b.Chapters = make([]*Chapter, numChapters)
	tracker := int64(0)
	for i := 0; i < len(b.Chapters); i++ {
		b.Chapters[i] = new(Chapter)
		b.Chapters[i].Title = fmt.Sprintf("Chapter %d", i+1)
		b.Chapters[i].StartMs = tracker
		tracker += int64(chapterLengthMin * 1000 * 60)
		b.Chapters[i].EndMs = tracker
		b.Chapters[i].Number = i
		b.Chapters[i].LengthMs = int64(chapterLengthMin * 1000 * 60)
	}

	if extraChapterLen > 0 {
		log.Debugf("There is an extra chapter required! %dms more needs adding", extraChapterLen)
		c := new(Chapter)
		c.Title = fmt.Sprintf("Chapter %d", len(b.Chapters)+1)
		stamp := int64(chapterLengthMin * 1000 * 60 * len(b.Chapters))
		c.StartMs = stamp
		c.EndMs = stamp + extraChapterLen
		c.Number = len(b.Chapters) + 1
		c.LengthMs = extraChapterLen
		b.Chapters = append(b.Chapters, c)
	}

	return nil
}

// ParseToChapters creates Chapter objects out of tagged files
func (b *Book) ParseToChapters(config Config) error {
	currentChapter := new(Chapter)
	chapterIndex := 0

	for _, fileName := range config.transcodeFiles {
		// parse track file
		track := TrackFile{}
		if err := track.Parse(fileName); err != nil {
			return err
		}

		// get tags from the file
		// TODO confirm this works with non-mp3s
		trackTag, err := tag.ReadFrom(track.File)
		if err != nil {
			return err
		}

		// if the last chapter name does not match the current track title, process as a new Chapter
		if currentChapter.Title != trackTag.Title() {
			log.Debugf("%s has chapter name: %s", track.File.Name(), trackTag.Title())
			if currentChapter.Title != "" {
				// if this is a new chapter, compile the old one and add it to the listing
				currentChapter.Compile()
				b.Chapters = append(b.Chapters, currentChapter)
				chapterIndex++
			}
			// Create new chapter and fill it with data
			c := Chapter{}
			c.Number = chapterIndex
			c.Title = trackTag.Title()
			c.Tracks = []TrackFile{track}
			// set this new chapter as the current working chapter
			currentChapter = &c
		} else {
			// Not a new chapter, so add it to the chapter tracks list
			log.Debugf("%s has no new chapter, adding to %s", track.File.Name(), currentChapter.Title)
			currentChapter.Tracks = append(currentChapter.Tracks, track)
		}
	}

	// Compile data for final chapter
	currentChapter.Compile()
	b.Chapters = append(b.Chapters, currentChapter)
	b.CalcChapterTimes()

	return nil
}

// ChapterByFile creates Chapter objects from individual files
func (b *Book) ChapterByFile(config Config, useFileNames, useTagTitle bool) error {
	for idx, filename := range config.transcodeFiles {
		chapter := new(Chapter)
		track := TrackFile{}
		if err := track.Parse(filename); err != nil {
			return err
		}

		if useFileNames {
			// use filename as the Chapter title
			// capture and remove extension from name
			name := filepath.Base(track.File.Name())
			fileExt := filepath.Ext(name)
			chapter.Title = strings.TrimSuffix(name, fileExt)
		} else if useTagTitle {
			// use track title tag as the Chapter title
			trackTag, err := tag.ReadFrom(track.File)
			if err != nil {
				return err
			}
			chapter.Title = trackTag.Title()
		} else {
			// Use the index as a Chapter title
			chapter.Title = fmt.Sprintf("Chapter %d", idx+1)
		}
		chapter.Number = idx
		chapter.Tracks = []TrackFile{track}
		chapter.Compile()
		c := &chapter
		b.Chapters = append(b.Chapters, *c)
	}

	b.CalcChapterTimes()

	return nil
}

// ExtractChapters creates file from embedded chapters
func (b *Book) ExtractChapters(config Config) error {
	var err error
	if len(config.sourceFiles) > 1 {
		return errors.New("only one source file is allowed for now")
	}

	err = ffmpeg_go.Input(config.sourceFiles[0]).
		Output(filepath.Join(config.scratchDir, "extracted-chapters.ini"), ffmpeg_go.KwArgs{"f": "ffmetadata"}).
		OverWriteOutput().
		Run()
	if err != nil {
		return err
	}
	return nil
}
