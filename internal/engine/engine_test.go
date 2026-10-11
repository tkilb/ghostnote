package engine

import (
	"testing"
	"time"

	"github.com/tylerkilburn/ghostnote/internal/config"
	"github.com/tylerkilburn/ghostnote/internal/metadata"
)

func TestFormatTime(t *testing.T) {
	if s := FormatTime(45); s != "00:45" {
		t.Errorf("FormatTime(45) = %s, want 00:45", s)
	}
	if s := FormatTime(125); s != "02:05" {
		t.Errorf("FormatTime(125) = %s, want 02:05", s)
	}
	if s := FormatTime(3665); s != "01:01:05" {
		t.Errorf("FormatTime(3665) = %s, want 01:01:05", s)
	}
}

func TestFormatBitrate(t *testing.T) {
	if b := FormatBitrate("320"); b != "320k" {
		t.Errorf("FormatBitrate(320) = %s, want 320k", b)
	}
	if b := FormatBitrate("256k"); b != "256k" {
		t.Errorf("FormatBitrate(256k) = %s, want 256k", b)
	}
	if b := FormatBitrate(""); b != "320k" {
		t.Errorf("FormatBitrate() = %s, want 320k", b)
	}
}

func TestJobStatusStrings(t *testing.T) {
	if StatusQueued.String() != "Queued" {
		t.Errorf("expected Queued, got %s", StatusQueued.String())
	}
	if StatusDownloading.String() != "Downloading" {
		t.Errorf("expected Downloading, got %s", StatusDownloading.String())
	}
	if StatusSlicing.String() != "Slicing" {
		t.Errorf("expected Slicing, got %s", StatusSlicing.String())
	}
}

type testListener struct {
	updated []*Job
}

func (l *testListener) OnJobUpdated(job *Job) {
	l.updated = append(l.updated, job)
}

func TestQueueOperations(t *testing.T) {
	cfg := config.Default()
	cfg.Destination = t.TempDir()

	q := NewQueue(cfg)
	defer q.Close()

	listener := &testListener{}
	q.AddListener(listener)

	// Add manual job directly to verify queue state
	job := &Job{
		ID:        "test-1",
		Title:     "Sample Track",
		Status:    StatusQueued,
		StartTime: time.Now(),
		Tags: metadata.TrackTags{
			Title:  "Sample Track",
			Artist: "Test Artist",
		},
	}

	q.mu.Lock()
	q.jobs = append(q.jobs, job)
	q.mu.Unlock()

	if len(q.Jobs()) != 1 {
		t.Fatalf("expected 1 job, got %d", len(q.Jobs()))
	}

	// Remove job
	q.RemoveJob("test-1")
	if len(q.Jobs()) != 0 {
		t.Fatalf("expected 0 jobs after removal, got %d", len(q.Jobs()))
	}
}

func TestCheckBinaries(t *testing.T) {
	// ffmpeg is installed on this environment
	if err := CheckFFmpeg(); err != nil {
		t.Errorf("expected CheckFFmpeg to succeed: %v", err)
	}
	// yt-dlp is checked
	_ = CheckYtDlp()
}
