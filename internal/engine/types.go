package engine

import (
	"fmt"
	"time"

	"github.com/tylerkilburn/ghostnote/internal/metadata"
)

// JobStatus represents the lifecycle state of a track or download item.
type JobStatus int

const (
	StatusQueued JobStatus = iota
	StatusProbing
	StatusDownloading
	StatusSlicing
	StatusTagging
	StatusCompleted
	StatusFailed
	StatusCanceled
)

func (s JobStatus) String() string {
	switch s {
	case StatusQueued:
		return "Queued"
	case StatusProbing:
		return "Probing"
	case StatusDownloading:
		return "Downloading"
	case StatusSlicing:
		return "Slicing"
	case StatusTagging:
		return "Tagging"
	case StatusCompleted:
		return "Completed"
	case StatusFailed:
		return "Failed"
	case StatusCanceled:
		return "Canceled"
	default:
		return "Unknown"
	}
}

// Chapter represents an individual segment/timestamp in a stitched video.
type Chapter struct {
	Index     int     `json:"index"`
	Title     string  `json:"title"`
	StartTime float64 `json:"start_time"`
	EndTime   float64 `json:"end_time"`
}

// FormatTime formats seconds into MM:SS or HH:MM:SS string.
func FormatTime(secs float64) string {
	d := time.Duration(secs * float64(time.Second))
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	if h > 0 {
		return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%02d:%02d", m, s)
}

// Job represents an item in the download/processing queue.
type Job struct {
	ID          string
	URL         string
	Title       string
	RawTitle    string
	Uploader    string
	Duration    float64
	Thumbnail   string
	Chapters    []Chapter
	IsParent    bool     // True if this is a parent job that gets expanded into chapters
	AutoSplit   bool     // True if user requested automatic splitting of continuous audio
	ParentID    string   // ID of parent job if this is a chapter child
	ChapterIdx  int      // 1-based index if chapter
	TotalTracks int      // Total chapters/tracks in album

	// Output & Metadata
	Tags       metadata.TrackTags
	Format     string // mp3, flac, opus, copy
	Bitrate    string // 320k
	DestPath   string // Destination file path on disk
	CachedFile string // Cache audio path for slice source

	// Live status
	Status      JobStatus
	Progress    float64 // 0.0 - 100.0
	Speed       string  // e.g. "3.5MiB/s"
	ETA         string  // e.g. "00:15"
	ErrorMsg    string
	StartTime   time.Time
	EndTime     time.Time
}
