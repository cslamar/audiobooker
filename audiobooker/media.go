package audiobooker

import (
	"context"
	"gopkg.in/vansante/go-ffprobe.v2"
	"os"
)

// getFileMetadata returns extracted metadata from media file
func getFileMetadata(filename string) (*ffprobe.ProbeData, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	fileMetadata, err := ffprobe.ProbeURL(context.Background(), f.Name())
	if err != nil {
		return nil, err
	}
	f.Close()

	return fileMetadata, nil
}
