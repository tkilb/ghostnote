package engine

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/tylerkilburn/ghostnote/internal/config"
	"github.com/tylerkilburn/ghostnote/internal/metadata"
)

// QueueListener notifies external consumers (such as TUI models) of state updates.
type QueueListener interface {
	OnJobUpdated(job *Job)
}

// Queue manages background download and slicing worker pool.
type Queue struct {
	mu        sync.RWMutex
	jobs      []*Job
	completed []*Job
	cfg       *config.Config
	workChan  chan *Job
	listeners []QueueListener
	ctx       context.Context
	cancel    context.CancelFunc
	cacheDir  string
}

// NewQueue creates and initializes a Queue instance.
func NewQueue(cfg *config.Config) *Queue {
	ctx, cancel := context.WithCancel(context.Background())
	cacheDir := filepath.Join(os.TempDir(), "ghostnote_cache")
	_ = os.MkdirAll(cacheDir, 0o755)

	q := &Queue{
		cfg:       cfg,
		workChan:  make(chan *Job, 200),
		ctx:       ctx,
		cancel:    cancel,
		cacheDir:  cacheDir,
	}

	// Start worker pool
	workers := cfg.Queue.MaxConcurrent
	if workers <= 0 {
		workers = 3
	}
	for i := 0; i < workers; i++ {
		go q.worker(i)
	}

	return q
}

// Close cancels active work and clears temporary files.
func (q *Queue) Close() {
	q.cancel()
}

// AddListener registers a listener for job events.
func (q *Queue) AddListener(l QueueListener) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.listeners = append(q.listeners, l)
}

func (q *Queue) notify(job *Job) {
	q.mu.RLock()
	listeners := append([]QueueListener(nil), q.listeners...)
	q.mu.RUnlock()

	for _, l := range listeners {
		l.OnJobUpdated(job)
	}
}

// AddURL queues a URL for inspection, expansion, and download.
func (q *Queue) AddURL(rawURL string) error {
	rawURL = filepath.Clean(rawURL)
	if rawURL == "" {
		return fmt.Errorf("empty URL provided")
	}

	job := &Job{
		ID:        fmt.Sprintf("job_%d", time.Now().UnixNano()),
		URL:       rawURL,
		RawTitle:  rawURL,
		Title:     "Inspecting...",
		Status:    StatusQueued,
		StartTime: time.Now(),
	}

	q.mu.Lock()
	q.jobs = append(q.jobs, job)
	q.mu.Unlock()

	q.notify(job)
	q.workChan <- job
	return nil
}

// Jobs returns a snapshot of active and pending jobs.
func (q *Queue) Jobs() []*Job {
	q.mu.RLock()
	defer q.mu.RUnlock()
	res := make([]*Job, len(q.jobs))
	copy(res, q.jobs)
	return res
}

// Completed returns a snapshot of successfully completed jobs.
func (q *Queue) Completed() []*Job {
	q.mu.RLock()
	defer q.mu.RUnlock()
	res := make([]*Job, len(q.completed))
	copy(res, q.completed)
	return res
}

// ClearCompleted clears the list of completed jobs.
func (q *Queue) ClearCompleted() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.completed = nil
}

// RemoveJob cancels or removes a job by ID.
func (q *Queue) RemoveJob(id string) {
	var canceledJob *Job

	q.mu.Lock()
	for i, j := range q.jobs {
		if j.ID == id {
			j.Status = StatusCanceled
			canceledJob = j
			q.jobs = append(q.jobs[:i], q.jobs[i+1:]...)
			break
		}
	}
	q.mu.Unlock()

	if canceledJob != nil {
		q.notify(canceledJob)
	}
}

// RetryJob resets a failed or canceled job and puts it back on workChan.
func (q *Queue) RetryJob(id string) {
	q.mu.Lock()
	var target *Job
	for _, j := range q.jobs {
		if j.ID == id && (j.Status == StatusFailed || j.Status == StatusCanceled) {
			j.Status = StatusQueued
			j.Progress = 0
			j.ErrorMsg = ""
			j.StartTime = time.Now()
			target = j
			break
		}
	}
	q.mu.Unlock()

	if target != nil {
		q.notify(target)
		select {
		case q.workChan <- target:
		default:
		}
	}
}

// worker processes jobs sequentially from workChan.
func (q *Queue) worker(workerID int) {
	for {
		select {
		case <-q.ctx.Done():
			return
		case job, ok := <-q.workChan:
			if !ok {
				return
			}
			q.processJob(job)
		}
	}
}

func (q *Queue) processJob(job *Job) {
	if job.Status == StatusCanceled {
		return
	}

	// Step 1: Probe metadata if it's a fresh URL
	if job.Status == StatusQueued {
		job.Status = StatusProbing
		q.notify(job)

		meta, err := FetchMetadata(q.ctx, job.URL)
		if err != nil {
			job.Status = StatusFailed
			job.ErrorMsg = err.Error()
			q.notify(job)
			return
		}

		job.RawTitle = meta.Title
		job.Duration = meta.Duration
		job.Uploader = meta.Uploader
		job.Thumbnail = meta.Thumbnail

		// Check if chaptered video
		if len(meta.Chapters) > 0 {
			job.IsParent = true
			job.Status = StatusDownloading
			job.Title = fmt.Sprintf("[Album] %s", meta.Title)
			q.notify(job)

			// Download audio once to cache
			cachedAudio := filepath.Join(q.cacheDir, fmt.Sprintf("%s.cache", meta.ID))
			err := DownloadAudio(q.ctx, job.URL, cachedAudio, func(p DownloadProgress) {
				job.Progress = p.Percent
				job.Speed = p.Speed
				job.ETA = p.ETA
				q.notify(job)
			})

			if err != nil {
				job.Status = StatusFailed
				job.ErrorMsg = err.Error()
				q.notify(job)
				return
			}

			// Expand chapters into queue items
			totalTracks := len(meta.Chapters)
			for idx, ch := range meta.Chapters {
				parsed := metadata.ParseArtistAndTitle(ch.Title, meta.Uploader)
				tags := metadata.TrackTags{
					Title:       parsed.Title,
					Artist:      parsed.Artist,
					Album:       meta.Title,
					TrackNumber: idx + 1,
					TotalTracks: totalTracks,
				}

				filename := tags.GenerateFilename(q.cfg.Naming.OSTPattern, q.cfg.Audio.Format)
				destPath := filepath.Join(q.cfg.Destination, filename)

				chJob := &Job{
					ID:          fmt.Sprintf("%s_ch_%d", job.ID, idx+1),
					URL:         job.URL,
					Title:       tags.Title,
					RawTitle:    ch.Title,
					ParentID:    job.ID,
					ChapterIdx:  idx + 1,
					TotalTracks: totalTracks,
					Tags:        tags,
					Format:      q.cfg.Audio.Format,
					Bitrate:     q.cfg.Audio.Bitrate,
					DestPath:    destPath,
					CachedFile:  cachedAudio,
					Status:      StatusSlicing,
					StartTime:   time.Now(),
				}

				q.mu.Lock()
				q.jobs = append(q.jobs, chJob)
				q.mu.Unlock()
				q.notify(chJob)

				// Slice and tag
				sliceErr := SliceAndTag(q.ctx, SliceOptions{
					SourceFile: cachedAudio,
					OutputFile: destPath,
					StartTime:  ch.StartTime,
					EndTime:    ch.EndTime,
					Format:     chJob.Format,
					Bitrate:    chJob.Bitrate,
					Tags:       tags,
				})

				if sliceErr != nil {
					chJob.Status = StatusFailed
					chJob.ErrorMsg = sliceErr.Error()
				} else {
					chJob.Status = StatusCompleted
					chJob.Progress = 100.0
					chJob.EndTime = time.Now()
					q.mu.Lock()
					q.completed = append(q.completed, chJob)
					q.mu.Unlock()
				}
				q.notify(chJob)
			}

			job.Status = StatusCompleted
			job.Progress = 100.0
			job.EndTime = time.Now()
			q.notify(job)
			return
		}

		// Single Track
		parsed := metadata.ParseArtistAndTitle(meta.Title, meta.Uploader)
		job.Title = parsed.Title
		job.Tags = metadata.TrackTags{
			Title:  parsed.Title,
			Artist: parsed.Artist,
		}
		filename := job.Tags.GenerateFilename(q.cfg.Naming.Pattern, q.cfg.Audio.Format)
		job.DestPath = filepath.Join(q.cfg.Destination, filename)
		job.Format = q.cfg.Audio.Format
		job.Bitrate = q.cfg.Audio.Bitrate

		// Duplicate detection: check if file already exists in destination
		if _, statErr := os.Stat(job.DestPath); statErr == nil {
			job.Status = StatusCompleted
			job.Progress = 100.0
			job.EndTime = time.Now()
			job.ErrorMsg = "[Existing file found in destination]"
			q.mu.Lock()
			q.completed = append(q.completed, job)
			q.mu.Unlock()
			q.notify(job)
			return
		}

		job.Status = StatusDownloading
		q.notify(job)

		err = DownloadAudio(q.ctx, job.URL, job.DestPath, func(p DownloadProgress) {
			job.Progress = p.Percent
			job.Speed = p.Speed
			job.ETA = p.ETA
			q.notify(job)
		})

		if err != nil {
			job.Status = StatusFailed
			job.ErrorMsg = err.Error()
			q.notify(job)
			return
		}

		job.Status = StatusCompleted
		job.Progress = 100.0
		job.EndTime = time.Now()
		q.mu.Lock()
		q.completed = append(q.completed, job)
		q.mu.Unlock()
		q.notify(job)
	}
}
