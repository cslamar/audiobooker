package audiobooker

import (
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"os/exec"
	"strconv"
	"strings"
)

// MarkerPoint tracks the marker points for silence detection
type MarkerPoint struct {
	// Duration of the silence detected
	Duration float64
	// End in seconds of the silence
	End float64
}

// ParseEnd returns the endpoint of the silence trimming off some for better playback experience
func (p MarkerPoint) ParseEnd() float64 {
	return p.End - (p.Duration / 2)
}

// GenerateVolMarkers parses file for silence detection marker points
func GenerateVolMarkers(filename string, duration float64, dbFloor int) ([]MarkerPoint, error) {
	log.Debugln("generating silence detection marker points")
	if dbFloor >= 0 {
		return nil, errors.New("dbFloor MUST be less than 0")
	}

	//cmd := exec.Command("ffmpeg", "-i", filename, "-af", "silencedetect=noise=-30dB:d=3.5", "-f", "null", "-")
	cmd := exec.Command("ffmpeg", "-i", filename, "-af", fmt.Sprintf("silencedetect=noise=%ddB:d=%f", dbFloor, duration), "-f", "null", "-")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	markers := make([]MarkerPoint, 0)
	volLines := strings.Split(string(output), "\n")
	for _, line := range volLines {
		if strings.Contains(line, "silence_end") {
			log.Info(line)
			cols := strings.Split(line, " ")
			endPoint, err := strconv.ParseFloat(cols[4], 32)
			if err != nil {
				return nil, err
			}
			silenceDuration, err := strconv.ParseFloat(cols[7], 32)
			if err != nil {
				return nil, err
			}
			marker := MarkerPoint{
				Duration: silenceDuration,
				End:      endPoint,
			}
			markers = append(markers, marker)
		}
	}

	log.Debugln("marker points found:", len(markers))
	return markers, nil
}
