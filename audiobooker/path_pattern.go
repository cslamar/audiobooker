package audiobooker

import (
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/vjeantet/grok"
	"path/filepath"
	"strconv"
	"strings"
)

// path templates
const (
	AudioFile   = "%f"
	Author      = "%a"
	Genre       = "%g"
	Narrator    = "%n"
	ReleaseDate = "%y"
	Series      = "%s"
	SeriesPart  = "%p"
	Title       = "%t"
)

// pattern Groks
const (
	AuthorGrok      = "%{GREEDYDATA:author}"
	AudioFileGrok   = "%{AUDIO_FILE:audio_file}"
	GenreGrok       = "%{GREEDYDATA:genre}"
	NarratorGrok    = "%{GREEDYDATA:narrator}"
	ReleaseDateGrok = "%{NUMBER:release_date}"
	SeriesGrok      = "%{GREEDYDATA:series}"
	SeriesPartGrok  = "%{NUMBER:series_part}"
	TitleGrok       = "%{GREEDYDATA:title}"
)

// ParsePathTags takes in file path and pattern string returning a map of the pattern matched values
func ParsePathTags(path, pathPattern string) (map[string]string, error) {
	// sanitize path pattern to remove trailing /
	pathPattern = strings.TrimSuffix(pathPattern, "/")
	// split pattern by directory slashes
	parserPatterns := strings.Split(pathPattern, "/")
	// TODO if this fails try `\` for Windows?
	log.Debugln("parser pattern tags:", parserPatterns)

	// create list for patterns
	parserPatternsList := make([]any, len(parserPatterns))
	pathTemplate := ""

	// loop through each parser pattern and associate the correct Grok
	for i := 0; i < len(parserPatterns); i++ {
		switch parserPatterns[i] {
		case AudioFile:
			parserPatternsList[i] = AudioFileGrok
		case Author:
			parserPatternsList[i] = AuthorGrok
		case Genre:
			parserPatternsList[i] = GenreGrok
		case Narrator:
			parserPatternsList[i] = NarratorGrok
		case ReleaseDate:
			parserPatternsList[i] = ReleaseDateGrok
		case Series:
			parserPatternsList[i] = SeriesGrok
		case SeriesPart:
			parserPatternsList[i] = SeriesPartGrok
		case Title:
			parserPatternsList[i] = TitleGrok
		default:
			log.Debugf("couldn't match %s, adding it as a litteral", parserPatterns[i])
			parserPatternsList[i] = parserPatterns[i]
		}
		pathTemplate += "%s/"
	}
	pathTemplate = strings.TrimSuffix(pathTemplate, "/")

	// combine template with list of Groks
	pattern := fmt.Sprintf(pathTemplate, parserPatternsList...)
	log.Debugln("parser pattern:", pattern)

	patterns := make(map[string]string)
	patterns["AUDIO_FILE"] = `%{GREEDYDATA}\.(:?flac|mp3|m4a|ogg|opus)`
	patterns["NUMBER"] = `\d+`

	// create grok from defined patterns only returning named captures
	g, err := grok.NewWithConfig(&grok.Config{Patterns: patterns, NamedCapturesOnly: true})
	if err != nil {
		return nil, err
	}

	// parse out path values
	values, err := g.Parse(pattern, path)
	if err != nil {
		return nil, err
	}

	// if no tags were parsed, error out since something wasn't right
	if len(values) == 0 {
		errStr := fmt.Sprintf(`no tags were parsed from the path template
Parsed patterns were: %s
Parsed path would have been: %s
Input path was: %s
`, parserPatterns, filepath.Join(parserPatterns...), path)
		return nil, errors.New(errStr)
	}

	return values, nil
}

// OutputPathPattern renders directory structure based on path pattern and Book data
func OutputPathPattern(book Book, pathPattern string) (string, error) {
	// sanitize path pattern to remove trailing /
	pathPattern = strings.TrimSuffix(pathPattern, "/")
	// split pattern by directory slashes
	parserPatterns := strings.Split(pathPattern, "/")
	// TODO if this fails try `\` for Windows?
	log.Debugln("parser pattern tags:", parserPatterns)

	// TODO rethink this if default dir can be used?
	if len(parserPatterns) == 1 {
		return "", errors.New("path parser is invalid")
	}

	// create list for patterns
	outputAttributes := make([]string, len(parserPatterns))

	for i := 0; i < len(outputAttributes); i++ {
		switch parserPatterns[i] {
		case Author:
			outputAttributes[i] = book.Author
		case Genre:
			outputAttributes[i] = *book.Genre
		case Narrator:
			outputAttributes[i] = *book.Narrator
		case ReleaseDate:
			outputAttributes[i] = *book.Date
		case Series:
			outputAttributes[i] = *book.seriesName
		case SeriesPart:
			outputAttributes[i] = strconv.Itoa(*book.seriesPart)
		case Title:
			outputAttributes[i] = book.Title
		default:
			log.Debugf("couldn't match %s, adding it as a litteral", parserPatterns[i])
			outputAttributes[i] = parserPatterns[i]
		}
	}

	// TODO maybe use filepath.join?
	outputPath := strings.Join(outputAttributes, "/")

	return outputPath, nil
}

// OutputFilePattern renders filename based on path pattern and Book data
func OutputFilePattern(book Book, pathPattern string) string {
	if pathPattern == "" {
		return fmt.Sprintf("%s - %s.m4b", book.Author, book.Title)
	}

	pathPattern = strings.ReplaceAll(pathPattern, Author, book.Author)
	if book.Genre != nil {
		pathPattern = strings.ReplaceAll(pathPattern, Genre, *book.Genre)
	}
	if book.Narrator != nil {
		pathPattern = strings.ReplaceAll(pathPattern, Narrator, *book.Narrator)
	}
	if book.Date != nil {
		pathPattern = strings.ReplaceAll(pathPattern, ReleaseDate, *book.Date)
	}
	if book.seriesName != nil {
		pathPattern = strings.ReplaceAll(pathPattern, Series, *book.seriesName)
	}
	if book.seriesPart != nil {
		pathPattern = strings.ReplaceAll(pathPattern, SeriesPart, strconv.Itoa(*book.seriesPart))
	}
	pathPattern = strings.ReplaceAll(pathPattern, Title, book.Title)

	return fmt.Sprintf("%s.m4b", pathPattern)
}
