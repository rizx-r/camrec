package recorder

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/jackc/pgx/v5/pgxpool"

	"camrec/internal/db"
	"camrec/internal/storage"
)

type Config struct {
	FfmpegPath           string
	RtspURL              string
	OutputDir            string
	SegmentSeconds       int
	ResetTimestamps      bool
	RtspTransportTCP     bool
	UseStrftimeFilenames bool
}

type Recorder struct {
	cfg   Config
	store *storage.Minio
	pool  *pgxpool.Pool
}

func New(cfg Config, store *storage.Minio, pool *pgxpool.Pool) *Recorder {
	return &Recorder{cfg: cfg, store: store, pool: pool}
}

func (r *Recorder) Start(ctx context.Context) error {
	if err := os.MkdirAll(r.cfg.OutputDir, 0755); err != nil {
		return err
	}
	outPattern := filepath.Join(r.cfg.OutputDir, "%Y%m%d_%H%M%S.mp4")
	if !r.cfg.UseStrftimeFilenames {
		outPattern = filepath.Join(r.cfg.OutputDir, "output_%03d.mp4")
	}
	args := []string{"-y"}
	if r.cfg.RtspTransportTCP {
		args = append(args, "-rtsp_transport", "tcp")
	}
	args = append(args, "-i", r.cfg.RtspURL, "-c", "copy", "-f", "segment", "-segment_time", strconv.Itoa(r.cfg.SegmentSeconds))
	if r.cfg.ResetTimestamps {
		args = append(args, "-reset_timestamps", "1")
	}
	args = append(args, "-strftime", "1", outPattern)
	log.Printf("running ffmpeg: %s %s", r.cfg.FfmpegPath, strings.Join(args, " "))
	cmd := exec.CommandContext(ctx, r.cfg.FfmpegPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return err
	}
	go func() { _ = cmd.Wait() }()
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	if err := watcher.Add(r.cfg.OutputDir); err != nil {
		return err
	}
	go r.handleEvents(ctx, watcher)
	return nil
}

func (r *Recorder) handleEvents(ctx context.Context, w *fsnotify.Watcher) {
	for {
		select {
		case ev := <-w.Events:
			if ev.Op&fsnotify.Create == fsnotify.Create || ev.Op&fsnotify.Rename == fsnotify.Rename || ev.Op&fsnotify.Write == fsnotify.Write {
				if strings.HasSuffix(strings.ToLower(ev.Name), ".mp4") {
					r.processFile(ctx, ev.Name)
				}
			}
		case <-ctx.Done():
			_ = w.Close()
			return
		}
	}
}

func (r *Recorder) processFile(ctx context.Context, path string) {
	for i := 0; i < 30; i++ {
		fi, err := os.Stat(path)
		if err == nil {
			if fi.Size() > 0 {
				break
			}
		}
		time.Sleep(2 * time.Second)
	}
	fi, err := os.Stat(path)
	if err != nil {
		return
	}
	base := filepath.Base(path)
	start := parseStartTime(base)
	end := start.Add(time.Duration(r.cfg.SegmentSeconds) * time.Second)
	key := filepath.ToSlash(filepath.Join("videos", fmt.Sprintf("%04d", start.Year()), fmt.Sprintf("%02d", start.Month()), fmt.Sprintf("%02d", start.Day()), base))
	_, err = r.store.UploadFile(ctx, key, path, "video/mp4")
	if err != nil {
		return
	}
	_ = db.InsertVideo(ctx, r.pool, db.Video{ObjectKey: key, StartTime: start, EndTime: end, SizeBytes: fi.Size()})
	now := time.Now().In(time.Local)
	sLoc := start.In(time.Local)
	if !(sLoc.Year() == now.Year() && sLoc.YearDay() == now.YearDay()) {
		_ = os.Remove(path)
	}
}

func parseStartTime(name string) time.Time {
	n := strings.TrimSuffix(name, filepath.Ext(name))
	t, err := time.Parse("20060102_150405", n)
	if err != nil {
		return time.Now()
	}
	return t
}
