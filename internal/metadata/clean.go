package metadata

import (
	"regexp"
	"strings"
)

// Noise tokens commonly found in video titles.
var noisePatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)\s*[\(\[\{]\s*(?:official\s+)?(?:music\s+)?video\s*[\)\]\}]`),
	regexp.MustCompile(`(?i)\s*[\(\[\{]\s*official\s+audio\s*[\)\]\}]`),
	regexp.MustCompile(`(?i)\s*[\(\[\{]\s*audio\s*[\)\]\}]`),
	regexp.MustCompile(`(?i)\s*[\(\[\{]\s*lyric\s+video\s*[\)\]\}]`),
	regexp.MustCompile(`(?i)\s*[\(\[\{]\s*lyrics\s*[\)\]\}]`),
	regexp.MustCompile(`(?i)\s*[\(\[\{]\s*(?:hd|4k|1080p|hq|remastered|remaster)\s*[\)\]\}]`),
	regexp.MustCompile(`(?i)\s*[\(\[\{]\s*visualizer\s*[\)\]\}]`),
	regexp.MustCompile(`(?i)\s*[\(\[\{]\s*(?:full\s+album|full\s+ost|ost)\s*[\)\]\}]`),
	// Bare phrases without brackets
	regexp.MustCompile(`(?i)\b(?:official\s+music\s+video|official\s+video|official\s+audio|lyric\s+video)\b`),
}

var whitespaceRegex = regexp.MustCompile(`\s+`)
var separatorRegex = regexp.MustCompile(`\s*[-–—:|]\s*`)

// CleanedTitle holds parsed metadata components from a raw title string.
type CleanedTitle struct {
	Artist string
	Title  string
	Raw    string
}

// CleanTitle removes YouTube title noise tokens and trims excess whitespace.
func CleanTitle(raw string) string {
	s := raw
	for _, p := range noisePatterns {
		s = p.ReplaceAllString(s, "")
	}
	s = strings.TrimSpace(s)
	// Remove empty brackets or parens left behind
	emptyBracket := regexp.MustCompile(`[\(\[\{]\s*[\)\]\}]`)
	for emptyBracket.MatchString(s) {
		s = emptyBracket.ReplaceAllString(s, "")
	}
	s = whitespaceRegex.ReplaceAllString(s, " ")
	return strings.TrimSpace(s)
}

// ParseArtistAndTitle attempts to split a cleaned title into Artist and Title.
// Common formats: "Artist - Title", "Artist: Title", "Artist | Title".
// If no separator is found or if defaultArtist is provided, fallback logic applies.
func ParseArtistAndTitle(rawTitle, defaultArtist string) CleanedTitle {
	cleaned := CleanTitle(rawTitle)

	// First check if there is an explicit standard separator: " - ", " – ", " — "
	parts := separatorRegex.Split(cleaned, 2)
	if len(parts) == 2 && strings.TrimSpace(parts[0]) != "" && strings.TrimSpace(parts[1]) != "" {
		artist := strings.TrimSpace(parts[0])
		title := strings.TrimSpace(parts[1])
		return CleanedTitle{
			Artist: artist,
			Title:  title,
			Raw:    rawTitle,
		}
	}

	// Fallback when no separator is present:
	artist := strings.TrimSpace(defaultArtist)
	if artist == "" {
		artist = "Unknown Artist"
	}

	return CleanedTitle{
		Artist: artist,
		Title:  cleaned,
		Raw:    rawTitle,
	}
}

// SanitizeFilename strips invalid filesystem characters and preserves safe characters.
func SanitizeFilename(name string) string {
	// Disallow / \ : * ? " < > |
	invalid := regexp.MustCompile(`[\\/:*?"<>|]`)
	clean := invalid.ReplaceAllString(name, "_")
	clean = whitespaceRegex.ReplaceAllString(clean, " ")
	clean = strings.TrimSpace(clean)
	clean = strings.Trim(clean, ".")
	if clean == "" {
		return "unnamed"
	}
	return clean
}
