package metadata

import (
	"fmt"
	"strings"
)

// TrackTags represents audio ID3 / Vorbis tag metadata for export.
type TrackTags struct {
	Title       string
	Artist      string
	Album       string
	TrackNumber int
	TotalTracks int
	Year        string
	Comment     string
	CoverPath   string // Local path to thumbnail image if present
}

// GenerateFilename builds a file name based on naming pattern template.
func (t TrackTags) GenerateFilename(pattern, extension string) string {
	ext := strings.TrimPrefix(extension, ".")
	if ext == "" {
		ext = "mp3"
	}

	result := pattern
	result = strings.ReplaceAll(result, "{artist}", t.Artist)
	result = strings.ReplaceAll(result, "{title}", t.Title)
	result = strings.ReplaceAll(result, "{album}", t.Album)

	trackStr := fmt.Sprintf("%d", t.TrackNumber)
	trackPadded := fmt.Sprintf("%02d", t.TrackNumber)
	result = strings.ReplaceAll(result, "{track}", trackStr)
	result = strings.ReplaceAll(result, "{track:02d}", trackPadded)
	result = strings.ReplaceAll(result, "{year}", t.Year)

	cleaned := SanitizeFilename(result)
	return fmt.Sprintf("%s.%s", cleaned, ext)
}
