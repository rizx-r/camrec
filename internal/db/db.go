package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Video struct {
	ID        int64
	ObjectKey string
	StartTime time.Time
	EndTime   time.Time
	SizeBytes int64
	CreatedAt time.Time
}

func NewPool(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	return pgxpool.NewWithConfig(ctx, cfg)
}

func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `
        CREATE TABLE IF NOT EXISTS videos (
            id BIGSERIAL PRIMARY KEY,
            object_key TEXT NOT NULL,
            start_time TIMESTAMPTZ NOT NULL,
            end_time   TIMESTAMPTZ NOT NULL,
            size_bytes BIGINT NOT NULL,
            created_at TIMESTAMPTZ NOT NULL DEFAULT now()
        );
        CREATE INDEX IF NOT EXISTS idx_videos_start_time ON videos(start_time);
        CREATE INDEX IF NOT EXISTS idx_videos_created_at ON videos(created_at);
    `)
	return err
}

func InsertVideo(ctx context.Context, pool *pgxpool.Pool, v Video) error {
	_, err := pool.Exec(ctx, `INSERT INTO videos(object_key, start_time, end_time, size_bytes) VALUES($1,$2,$3,$4)`, v.ObjectKey, v.StartTime, v.EndTime, v.SizeBytes)
	return err
}

func ListAll(ctx context.Context, pool *pgxpool.Pool) ([]Video, error) {
	rows, err := pool.Query(ctx, `SELECT id, object_key, start_time, end_time, size_bytes, created_at FROM videos ORDER BY start_time ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Video
	for rows.Next() {
		var v Video
		if err := rows.Scan(&v.ID, &v.ObjectKey, &v.StartTime, &v.EndTime, &v.SizeBytes, &v.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func ListRange(ctx context.Context, pool *pgxpool.Pool, start, end time.Time) ([]Video, error) {
	rows, err := pool.Query(ctx, `SELECT id, object_key, start_time, end_time, size_bytes, created_at FROM videos WHERE start_time >= $1 AND end_time <= $2 ORDER BY start_time ASC`, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Video
	for rows.Next() {
		var v Video
		if err := rows.Scan(&v.ID, &v.ObjectKey, &v.StartTime, &v.EndTime, &v.SizeBytes, &v.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func ListLatest(ctx context.Context, pool *pgxpool.Pool, n int) ([]Video, error) {
	rows, err := pool.Query(ctx, `SELECT id, object_key, start_time, end_time, size_bytes, created_at FROM videos ORDER BY start_time DESC LIMIT $1`, n)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Video
	for rows.Next() {
		var v Video
		if err := rows.Scan(&v.ID, &v.ObjectKey, &v.StartTime, &v.EndTime, &v.SizeBytes, &v.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}
