package api

import (
	"time"
)

type VideoDTO struct {
	URL       string    `json:"url"`
	Key       string    `json:"key"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	SizeBytes int64     `json:"size_bytes"`
}
