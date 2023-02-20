package audiobooker

import (
	"context"
	"gopkg.in/vansante/go-ffprobe.v2"
	"os"
)

// TrackFile handles individual data and operations for files
type TrackFile struct {
	LengthMs int64
	File     *os.File
}

// Parse loads the metadata from file including the file name and length
func (t *TrackFile) Parse(filename string) error {
	var err error
	t.File, err = os.Open(filename)
	if err != nil {
		return err
	}

	lenMs, err := ffprobe.ProbeURL(context.Background(), t.File.Name())
	if err != nil {
		return err
	}

	t.LengthMs = lenMs.Format.Duration().Milliseconds()

	return nil
}
