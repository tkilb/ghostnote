package engine

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/tylerkilburn/ghostnote/internal/metadata"
)

// CheckFFmpeg verifies that ffmpeg is installed and available in PATH.
func CheckFFmpeg() error {
	_, err := exec.LookPath("ffmpeg")
	if err != nil {
		return errors.New("ffmpeg is not installed or not in PATH")
	}
	return nil
}

// SliceOptions defines slicing and encoding options for an audio segment.
type SliceOptions struct {
	SourceFile  string
	OutputFile  string
	StartTime   float64
	EndTime     float64
	Format      string // mp3, flac, opus, copy
	Bitrate     string // e.g. 320k
	Tags        metadata.TrackTags
	CoverPath   string
}

// SliceAndTag invokes ffmpeg to cut the chapter audio, transcode, and apply tags/cover.
func SliceAndTag(ctx context.Context, opts SliceOptions) error {
	if err := CheckFFmpeg(); err != nil {
		return err
	}

	args := []string{"-y"} // overwrite output if exists

	// Time boundaries
	if opts.StartTime > 0 {
		args = append(args, "-ss", fmt.Sprintf("%.3f", opts.StartTime))
	}
	if opts.EndTime > opts.StartTime {
		args = append(args, "-to", fmt.Sprintf("%.3f", opts.EndTime))
	}

	// Audio input
	args = append(args, "-i", opts.SourceFile)

	// Cover art input if provided
	hasCover := opts.CoverPath != ""
	if hasCover {
		args = append(args, "-i", opts.CoverPath)
	}

	// Codec configuration
	switch strings.ToLower(opts.Format) {
	case "flac":
		args = append(args, "-c:a", "flac")
	case "opus":
		args = append(args, "-c:a", "libopus")
		if opts.Bitrate != "" {
			args = append(args, "-b:a", opts.Bitrate)
		}
	case "copy":
		args = append(args, "-c:a", "copy")
	case "mp3":
		fallthrough
	default:
		args = append(args, "-c:a", "libmp3lame")
		if opts.Bitrate != "" {
			args = append(args, "-b:a", opts.Bitrate)
		} else {
			args = append(args, "-b:a", "320k")
		}
	}

	// Tagging arguments
	if opts.Tags.Title != "" {
		args = append(args, "-metadata", fmt.Sprintf("title=%s", opts.Tags.Title))
	}
	if opts.Tags.Artist != "" {
		args = append(args, "-metadata", fmt.Sprintf("artist=%s", opts.Tags.Artist))
	}
	if opts.Tags.Album != "" {
		args = append(args, "-metadata", fmt.Sprintf("album=%s", opts.Tags.Album))
	}
	if opts.Tags.TrackNumber > 0 {
		if opts.Tags.TotalTracks > 0 {
			args = append(args, "-metadata", fmt.Sprintf("track=%d/%d", opts.Tags.TrackNumber, opts.Tags.TotalTracks))
		} else {
			args = append(args, "-metadata", fmt.Sprintf("track=%d", opts.Tags.TrackNumber))
		}
	}
	if opts.Tags.Year != "" {
		args = append(args, "-metadata", fmt.Sprintf("date=%s", opts.Tags.Year))
	}

	// Embed thumbnail mapping if cover art exists
	if hasCover {
		args = append(args, "-map", "0:a", "-map", "1:0")
		if strings.ToLower(opts.Format) == "mp3" {
			args = append(args, "-c:v", "mjpeg", "-id3v2_version", "3", "-metadata:s:v", "title=Album cover", "-metadata:s:v", "comment=Cover (front)")
		} else if strings.ToLower(opts.Format) == "flac" {
			args = append(args, "-c:v", "copy")
		}
	}

	args = append(args, opts.OutputFile)

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg failed: %s (%w)", string(out), err)
	}

	return nil
}

// FormatBitrate ensures bitrate argument format is valid (e.g. 320k).
func FormatBitrate(bitrate string) string {
	b := strings.TrimSpace(strings.ToLower(bitrate))
	if b == "" {
		return "320k"
	}
	if !strings.HasSuffix(b, "k") {
		if _, err := strconv.Atoi(b); err == nil {
			return b + "k"
		}
	}
	return b
}
