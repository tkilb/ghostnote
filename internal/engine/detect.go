package engine

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
)

var silenceEndRegex = regexp.MustCompile(`silence_end:\s*([\d\.]+)`)

// DetectSilenceSplits analyzes an audio file using ffmpeg's silencedetect filter
// and returns timestamps where silence ends (track transition boundaries).
// minDurationSecs prevents splitting into micro-clips (e.g. at least 30s apart).
func DetectSilenceSplits(ctx context.Context, audioFile string, totalDuration float64) ([]float64, error) {
	if err := CheckFFmpeg(); err != nil {
		return nil, err
	}

	// -af silencedetect=noise=-30dB:d=0.7
	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-i", audioFile,
		"-af", "silencedetect=noise=-30dB:d=0.7",
		"-f", "null",
		"-",
	)

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("reading ffmpeg stderr: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("starting silencedetect: %w", err)
	}

	var splitPoints []float64
	minGap := 45.0 // each track must be at least 45 seconds long

	scanner := bufio.NewScanner(stderr)
	for scanner.Scan() {
		line := scanner.Text()
		matches := silenceEndRegex.FindStringSubmatch(line)
		if len(matches) == 2 {
			ts, err := strconv.ParseFloat(matches[1], 64)
			if err == nil && ts > minGap && (totalDuration <= 0 || ts < totalDuration-minGap) {
				if len(splitPoints) == 0 || ts-splitPoints[len(splitPoints)-1] >= minGap {
					splitPoints = append(splitPoints, ts)
				}
			}
		}
	}

	_ = cmd.Wait()
	return splitPoints, nil
}
