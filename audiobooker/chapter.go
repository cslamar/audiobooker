package audiobooker

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"gopkg.in/vansante/go-ffprobe.v2"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Chapter holds data for each chapter construct
type Chapter struct {
	LengthMs int64
	StartMs  int64
	EndMs    int64
	Number   int
	Title    string
	Tracks   []TrackFile
}

// StampTimes compiles the start and end time for each chapter
func (c *Chapter) StampTimes(startTime int64) int64 {
	c.StartMs = startTime
	c.EndMs = startTime + c.LengthMs

	return c.EndMs
}

// Compile compiles Chapter length based off the collection of tracks associated to the chapter
func (c *Chapter) Compile() {
	c.LengthMs = 0
	for _, track := range c.Tracks {
		c.LengthMs += track.LengthMs
	}
}

// cueEntry holds the parsed chapter info based on a CUE SHEET
type cueEntry struct {
	Track              int
	Title              string
	TimeEndMs          int64
	TimeStartStr       string
	TimeStartMs        int64
	LengthMs           int64
	trackStartDuration time.Duration
}

// parseTimeCode parses CUE SHEET time code into a duration
func (c *cueEntry) parseTimeCode(timeCode string) error {
	timeValues := strings.Split(timeCode, ":")
	//toParse := fmt.Sprintf("%sm%ss%sms", timeValues[0], timeValues[1], timeValues[2]) // TODO could work out way to calculate frames if desired...
	toParse := fmt.Sprintf("%sm%ss", timeValues[0], timeValues[1])
	duration, err := time.ParseDuration(toParse)
	if err != nil {
		return err
	}

	c.trackStartDuration = duration
	return nil
}

// toChapter converts a cueEntry to a Chapter
func (c *cueEntry) toChapter() Chapter {
	return Chapter{
		LengthMs: c.TimeEndMs - c.TimeStartMs,
		StartMs:  c.TimeStartMs,
		EndMs:    c.TimeEndMs,
		Number:   c.Track,
		Title:    c.Title,
	}
}

// parseFromCueTag creates Chapter object from embedded CUESHEET tag
func parseFromCueTag(config Config) ([]cueEntry, error) {
	f, err := os.Open(config.sourceFiles[0])
	if err != nil {
		return nil, err
	}
	fileMetadata, err := ffprobe.ProbeURL(context.Background(), f.Name())
	if err != nil {
		return nil, err
	}

	trackSeconds, err := strconv.ParseFloat(fileMetadata.FirstAudioStream().Duration, 6)
	if err != nil {
		return nil, err
	}
	fullTrackMs := int64(trackSeconds * 1000)

	rawCueSheet, err := fileMetadata.Format.TagList.GetString("CUESHEET")
	if err == ffprobe.ErrTagNotFound {
		// return a tag not found error so the process can possibly continue
		return nil, ffprobe.ErrTagNotFound
	} else if err != nil {
		// error was more than not found, bail
		return nil, err
	}

	log.Debugln(rawCueSheet)

	// prepare for cue entry parsing
	var entry *cueEntry
	entries := make([]cueEntry, 0)
	trackCount := 1

	// loop through the raw CUE SHEET split by newlines
	for _, val := range strings.Split(rawCueSheet, "\n") {
		// sanitize the leading and trailing spaces prior to parsing
		clean := strings.TrimSpace(val)
		if strings.HasPrefix(clean, "TRACK ") {
			// if the line starts with TRACK, create a new entry and assign the index value to it
			entry = new(cueEntry)
			// TODO maybe grab track number from regexp?
			entry.Track = trackCount
		} else if strings.HasPrefix(clean, "TITLE") {
			// if the line starts with TITLE, clean it up and store it as the chapter title
			clean = strings.TrimPrefix(clean, "TITLE ")
			clean = strings.ReplaceAll(clean, `"`, "")
			entry.Title = clean
		} else if strings.HasPrefix(clean, "INDEX") {
			// if the line begins with INDEX, sanitize and capture the time code for later usage
			re := regexp.MustCompile(`INDEX \d+ `)
			clean = re.ReplaceAllString(clean, "")
			entry.TimeStartStr = clean
			// parse the string time into a usable time.Duration format for later
			if err := entry.parseTimeCode(clean); err != nil {
				log.Errorln(err)
				continue
			}

			// append the completed entry to the entries slice
			entries = append(entries, *entry)
			trackCount++
		}
	}

	tracker := int64(0)

	for idx := 0; idx < len(entries); idx++ {
		if idx == 0 {
			// if at the first element
			// figure out starting inference
			log.Debugf("at the first element\n")
			entries[idx].TimeStartMs = 0
			tracker += entries[idx+1].trackStartDuration.Milliseconds()
			entries[idx].TimeEndMs = tracker
			log.Debugf("track %d starts at %dms and ends at %dms\n", entries[idx].Track, 0, entries[idx].TimeEndMs)
		} else if idx == (len(entries) - 1) {
			// if at the last element
			log.Debugf("at the last element\n")
			entries[idx].TimeStartMs = tracker
			entries[idx].TimeEndMs = fullTrackMs
			log.Debugf("track %d starts at %d and ends at %d\n", entries[idx].Track, entries[idx].TimeStartMs, entries[idx].TimeEndMs)
		} else {
			// must be somewhere in the middle
			entries[idx].TimeStartMs = tracker
			tracker += entries[idx+1].trackStartDuration.Milliseconds() - tracker
			entries[idx].TimeEndMs = tracker
			log.Debugf("track %d starts at %d element and ends at %d\n", entries[idx].Track, entries[idx].TimeStartMs, entries[idx].TimeEndMs)
		}
	}

	return entries, nil
}
