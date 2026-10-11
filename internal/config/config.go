package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// AudioConfig holds target audio conversion settings.
type AudioConfig struct {
	Format  string `yaml:"format"`  // mp3, flac, opus, copy
	Bitrate string `yaml:"bitrate"` // e.g. 320k, 256k, 192k
}

// NamingConfig holds pattern preferences for output filenames.
type NamingConfig struct {
	Pattern    string `yaml:"pattern"`     // e.g. "{artist} - {title}"
	OSTPattern string `yaml:"ost_pattern"` // e.g. "{track:02d} - {title}"
	StripNoise bool   `yaml:"strip_noise"` // Strip YouTube video tags like [HD], (Official Video)
}

// QueueConfig holds queue concurrency and clipboard preferences.
type QueueConfig struct {
	MaxConcurrent      int  `yaml:"max_concurrent"`
	ClipboardQuickAdd bool `yaml:"clipboard_quick_add"`
}

// Config represents the application configuration.
type Config struct {
	Version     int          `yaml:"version"`
	Destination string       `yaml:"destination"`
	Audio       AudioConfig  `yaml:"audio"`
	Naming      NamingConfig `yaml:"naming"`
	Queue       QueueConfig  `yaml:"queue"`
}

// Default returns a Config initialized with sensible defaults.
func Default() *Config {
	home, err := os.UserHomeDir()
	dest := "~/Music"
	if err == nil {
		dest = filepath.Join(home, "Music")
	}

	return &Config{
		Version:     1,
		Destination: dest,
		Audio: AudioConfig{
			Format:  "mp3",
			Bitrate: "320k",
		},
		Naming: NamingConfig{
			Pattern:    "{artist} - {title}",
			OSTPattern: "{track:02d} - {title}",
			StripNoise: true,
		},
		Queue: QueueConfig{
			MaxConcurrent:      3,
			ClipboardQuickAdd: true,
		},
	}
}

// DefaultConfigPath returns the canonical path to ~/.config/ghostnote/config.yaml.
func DefaultConfigPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("unable to resolve user config dir: %w", err)
	}
	return filepath.Join(configDir, "ghostnote", "config.yaml"), nil
}

// Load loads configuration from the specified path, or default path if empty.
// If the file does not exist, it creates and saves the default configuration.
func Load(customPath ...string) (*Config, error) {
	var path string
	if len(customPath) > 0 && customPath[0] != "" {
		path = customPath[0]
	} else {
		var err error
		path, err = DefaultConfigPath()
		if err != nil {
			return nil, err
		}
	}

	path = ExpandHome(path)

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		cfg := Default()
		if err := cfg.Save(path); err != nil {
			// Non-fatal if we cannot write to disk, still return default config
			return cfg, nil
		}
		return cfg, nil
	} else if err != nil {
		return nil, fmt.Errorf("reading config file %q: %w", path, err)
	}

	cfg := Default()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parsing config yaml %q: %w", path, err)
	}

	cfg.Destination = ExpandHome(cfg.Destination)
	cfg.Validate()
	return cfg, nil
}

// Save writes configuration to disk as YAML.
func (c *Config) Save(customPath ...string) error {
	var path string
	if len(customPath) > 0 && customPath[0] != "" {
		path = customPath[0]
	} else {
		var err error
		path, err = DefaultConfigPath()
		if err != nil {
			return err
		}
	}

	path = ExpandHome(path)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("writing config file: %w", err)
	}

	return nil
}

// Validate sanitizes configuration fields to valid presets.
func (c *Config) Validate() {
	if c.Version == 0 {
		c.Version = 1
	}

	c.Audio.Format = strings.ToLower(strings.TrimSpace(c.Audio.Format))
	switch c.Audio.Format {
	case "mp3", "flac", "opus", "copy":
	default:
		c.Audio.Format = "mp3"
	}

	if c.Audio.Bitrate == "" {
		c.Audio.Bitrate = "320k"
	}

	if c.Naming.Pattern == "" {
		c.Naming.Pattern = "{artist} - {title}"
	}
	if c.Naming.OSTPattern == "" {
		c.Naming.OSTPattern = "{track:02d} - {title}"
	}

	if c.Queue.MaxConcurrent <= 0 {
		c.Queue.MaxConcurrent = 3
	}
}

// ExpandHome resolves leading ~ to the user's home directory.
func ExpandHome(path string) string {
	if strings.HasPrefix(path, "~/") || path == "~" {
		home, err := os.UserHomeDir()
		if err == nil {
			if path == "~" {
				return home
			}
			return filepath.Join(home, path[2:])
		}
	}
	return path
}
