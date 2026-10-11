package metadata

import (
	"testing"
)

func TestCleanTitle(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "Chrono Trigger - Corridors of Time (Official Music Video)",
			expected: "Chrono Trigger - Corridors of Time",
		},
		{
			input:    "Schala's Theme [HD] (Lyric Video)",
			expected: "Schala's Theme",
		},
		{
			input:    "Wind Scene [4K] (Audio)",
			expected: "Wind Scene",
		},
		{
			input:    "Yasunori Mitsuda - Frog's Theme (Remastered) [1080p]",
			expected: "Yasunori Mitsuda - Frog's Theme",
		},
		{
			input:    "Silent Light (Visualizer)",
			expected: "Silent Light",
		},
		{
			input:    "Robo's Theme (Full OST)",
			expected: "Robo's Theme",
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := CleanTitle(tt.input)
			if got != tt.expected {
				t.Errorf("CleanTitle(%q) = %q; want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestParseArtistAndTitle(t *testing.T) {
	tests := []struct {
		inputRaw      string
		defaultArtist string
		wantArtist    string
		wantTitle     string
	}{
		{
			inputRaw:      "Yasunori Mitsuda - Corridors of Time (Official Video)",
			defaultArtist: "Square Enix",
			wantArtist:    "Yasunori Mitsuda",
			wantTitle:     "Corridors of Time",
		},
		{
			inputRaw:      "Schala's Theme [HD]",
			defaultArtist: "Yasunori Mitsuda",
			wantArtist:    "Yasunori Mitsuda",
			wantTitle:     "Schala's Theme",
		},
		{
			inputRaw:      "Nobuo Uematsu — Tina's Theme",
			defaultArtist: "",
			wantArtist:    "Nobuo Uematsu",
			wantTitle:     "Tina's Theme",
		},
		{
			inputRaw:      "Mystery Track",
			defaultArtist: "",
			wantArtist:    "Unknown Artist",
			wantTitle:     "Mystery Track",
		},
	}

	for _, tt := range tests {
		t.Run(tt.inputRaw, func(t *testing.T) {
			res := ParseArtistAndTitle(tt.inputRaw, tt.defaultArtist)
			if res.Artist != tt.wantArtist {
				t.Errorf("expected artist %q, got %q", tt.wantArtist, res.Artist)
			}
			if res.Title != tt.wantTitle {
				t.Errorf("expected title %q, got %q", tt.wantTitle, res.Title)
			}
		})
	}
}

func TestSanitizeFilenameAndGenerateFilename(t *testing.T) {
	tag := TrackTags{
		Title:       "What/If: Battle *Theme*",
		Artist:      "AC/DC",
		Album:       "Rock? Roll!",
		TrackNumber: 3,
		Year:        "1980",
	}

	filename := tag.GenerateFilename("{track:02d} - {artist} - {title}", "mp3")
	expected := "03 - AC_DC - What_If_ Battle _Theme_.mp3"
	if filename != expected {
		t.Errorf("GenerateFilename got %q, want %q", filename, expected)
	}
}
