package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigDefaultAndValidation(t *testing.T) {
	cfg := Default()
	if cfg.Audio.Format != "mp3" {
		t.Fatalf("expected default format mp3, got %s", cfg.Audio.Format)
	}
	if cfg.Audio.Bitrate != "320k" {
		t.Fatalf("expected default bitrate 320k, got %s", cfg.Audio.Bitrate)
	}
	if cfg.Queue.MaxConcurrent != 3 {
		t.Fatalf("expected max concurrent 3, got %d", cfg.Queue.MaxConcurrent)
	}

	cfg.Audio.Format = "UNKNOWN"
	cfg.Audio.Bitrate = ""
	cfg.Queue.MaxConcurrent = -1
	cfg.Validate()

	if cfg.Audio.Format != "mp3" {
		t.Errorf("expected format reset to mp3, got %s", cfg.Audio.Format)
	}
	if cfg.Audio.Bitrate != "320k" {
		t.Errorf("expected bitrate reset to 320k, got %s", cfg.Audio.Bitrate)
	}
	if cfg.Queue.MaxConcurrent != 3 {
		t.Errorf("expected max concurrent reset to 3, got %d", cfg.Queue.MaxConcurrent)
	}
}

func TestConfigSaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	cfg := Default()
	cfg.Audio.Format = "flac"
	cfg.Audio.Bitrate = "original"
	cfg.Naming.StripNoise = false

	if err := cfg.Save(configPath); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	loaded, err := Load(configPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if loaded.Audio.Format != "flac" {
		t.Errorf("expected loaded format flac, got %s", loaded.Audio.Format)
	}
	if loaded.Naming.StripNoise != false {
		t.Errorf("expected strip_noise false, got true")
	}

	// Verify non-existent file creates default
	nonExistentPath := filepath.Join(tmpDir, "sub", "config.yaml")
	newLoaded, err := Load(nonExistentPath)
	if err != nil {
		t.Fatalf("expected default config creation, got error: %v", err)
	}
	if _, err := os.Stat(nonExistentPath); err != nil {
		t.Errorf("expected file to be created at %s", nonExistentPath)
	}
	if newLoaded.Audio.Format != "mp3" {
		t.Errorf("expected default format mp3, got %s", newLoaded.Audio.Format)
	}
}
