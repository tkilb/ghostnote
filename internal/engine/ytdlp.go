package engine

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
)

// YtDlpRawMetadata represents the structure output by `yt-dlp -J --flat-playlist`.
type YtDlpRawMetadata struct {
	ID          string            `json:"id"`
	Title       string            `json:"title"`
	Uploader    string            `json:"uploader"`
	Duration    float64           `json:"duration"`
	Thumbnail   string            `json:"thumbnail"`
	Chapters    []YtDlpRawChapter `json:"chapters"`
	Entries     []YtDlpRawMetadata `json:"entries"`
	Type        string            `json:"_type"`
}

type YtDlpRawChapter struct {
	Title     string  `json:"title"`
	StartTime float64 `json:"start_time"`
	EndTime   float64 `json:"end_time"`
}

// DownloadProgress reports real-time download status.
type DownloadProgress struct {
	Percent float64
	Speed   string
	ETA     string
}

// CheckYtDlp verifies if yt-dlp is installed and available in PATH.
func CheckYtDlp() error {
	_, err := exec.LookPath("yt-dlp")
	if err != nil {
		return errors.New("yt-dlp is not installed or not in PATH")
	}
	return nil
}

// FetchMetadata invokes `yt-dlp --dump-single-json --flat-playlist <url>` and parses output.
func FetchMetadata(ctx context.Context, url string) (*YtDlpRawMetadata, error) {
	if err := CheckYtDlp(); err != nil {
		return nil, err
	}

	cmd := exec.CommandContext(ctx, "yt-dlp",
		"--dump-single-json",
		"--flat-playlist",
		"--no-warnings",
		url,
	)

	out, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return nil, fmt.Errorf("yt-dlp failed: %s (%w)", string(exitErr.Stderr), err)
		}
		return nil, fmt.Errorf("running yt-dlp: %w", err)
	}

	var meta YtDlpRawMetadata
	if err := json.Unmarshal(out, &meta); err != nil {
		return nil, fmt.Errorf("parsing yt-dlp metadata JSON: %w", err)
	}

	return &meta, nil
}

// Regex to capture yt-dlp progress lines:
// e.g. "[download]  45.2% of ~  4.12MiB at    1.25MiB/s ETA 00:02"
var ytDlpProgressRegex = regexp.MustCompile(`\[download\]\s+([\d\.]+)%\s+of\s+.*?at\s+([^\s]+)\s+ETA\s+([^\s]+)`)

// DownloadAudio downloads the best quality audio stream to targetFile, streaming progress to onProgress callback.
func DownloadAudio(ctx context.Context, url, targetFile string, onProgress func(DownloadProgress)) error {
	if err := CheckYtDlp(); err != nil {
		return err
	}

	cmd := exec.CommandContext(ctx, "yt-dlp",
		"-x", // Extract audio
		"--audio-format", "best",
		"--no-playlist",
		"--newline",
		"-o", targetFile,
		url,
	)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("capturing stdout: %w", err)
	}
	cmd.Stderr = cmd.Stdout // merge stderr for inspection

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("starting yt-dlp: %w", err)
	}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		matches := ytDlpProgressRegex.FindStringSubmatch(line)
		if len(matches) == 4 {
			pct, _ := strconv.ParseFloat(matches[1], 64)
			if onProgress != nil {
				onProgress(DownloadProgress{
					Percent: pct,
					Speed:   matches[2],
					ETA:     matches[3],
				})
			}
		}
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("yt-dlp download failed: %w", err)
	}

	return nil
}
