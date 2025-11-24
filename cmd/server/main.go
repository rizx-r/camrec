package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"camrec/internal/config"
	"camrec/internal/db"
	"camrec/internal/handler"
	"camrec/internal/recorder"
	"camrec/internal/router"
	"camrec/internal/storage"
)

func main() {
	cfgPath := "config.yaml"
	if p := os.Getenv("CAMREC_CONFIG"); p != "" {
		cfgPath = p
	}
	cfg, err := config.Load(cfgPath)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	store, err := storage.NewMinio(ctx, storage.MinioConfig{
		Endpoint:  cfg.MinIO.Endpoint,
		AccessKey: cfg.MinIO.AccessKey,
		SecretKey: cfg.MinIO.SecretKey,
		Bucket:    cfg.MinIO.Bucket,
		UseSSL:    cfg.MinIO.UseSSL,
		Region:    cfg.MinIO.Region,
		Public:    cfg.Server.PublicBucketPolicy,
	})
	if err != nil {
		log.Fatal(err)
	}
	err = store.EnsureBucket(ctx)
	if err != nil {
		log.Fatal(err)
	}

	pool, err := db.NewPool(ctx, cfg.Postgres.DSN)
	if err != nil {
		log.Fatal(err)
	}
	err = db.Migrate(ctx, pool)
	if err != nil {
		log.Fatal(err)
	}

	rec := recorder.New(recorder.Config{
		FfmpegPath:           cfg.Recorder.FfmpegPath,
		RtspURL:              cfg.Recorder.RTSPURL,
		OutputDir:            cfg.Recorder.OutputDir,
		SegmentSeconds:       int(cfg.Recorder.SegmentSeconds),
		ResetTimestamps:      true,
		RtspTransportTCP:     true,
		UseStrftimeFilenames: true,
	}, store, pool)

	err = rec.Start(ctx)
	if err != nil {
		log.Fatal(err)
	}

	h := handler.NewVideoHandler(store, pool, cfg)
	r := router.New(h)
	srv := &http.Server{Addr: cfg.Server.Addr, Handler: r}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig
	cancel()
	ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelShutdown()
	_ = srv.Shutdown(ctxShutdown)
}
