package storage

import (
	"context"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Minio struct {
	client *minio.Client
	cfg    MinioConfig
}

type MinioConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	UseSSL    bool
	Region    string
	Public    bool
}

func NewMinio(ctx context.Context, cfg MinioConfig) (*Minio, error) {
	cl, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
		Region: cfg.Region,
	})
	if err != nil {
		return nil, err
	}
	return &Minio{client: cl, cfg: cfg}, nil
}

func (m *Minio) EnsureBucket(ctx context.Context) error {
	exists, err := m.client.BucketExists(ctx, m.cfg.Bucket)
	if err != nil {
		return err
	}
	if !exists {
		if err := m.client.MakeBucket(ctx, m.cfg.Bucket, minio.MakeBucketOptions{Region: m.cfg.Region}); err != nil {
			return err
		}
	}
	if m.cfg.Public {
		pol := "{\"Version\":\"2012-10-17\",\"Statement\":[{\"Effect\":\"Allow\",\"Principal\":{\"AWS\":\"*\"},\"Action\":[\"s3:GetObject\"],\"Resource\":[\"arn:aws:s3:::" + m.cfg.Bucket + "/**\"]}]}"
		_ = m.client.SetBucketPolicy(ctx, m.cfg.Bucket, pol)
	}
	return nil
}

func (m *Minio) UploadFile(ctx context.Context, objectName string, filePath string, contentType string) (minio.UploadInfo, error) {
	return m.client.FPutObject(ctx, m.cfg.Bucket, objectName, filePath, minio.PutObjectOptions{ContentType: contentType})
}

func (m *Minio) PresignURL(ctx context.Context, objectName string, expiry time.Duration) (string, error) {
	u, err := m.client.PresignedGetObject(ctx, m.cfg.Bucket, objectName, expiry, url.Values{})
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

func (m *Minio) PublicURL(objectName string) string {
	scheme := "http"
	if m.cfg.UseSSL {
		scheme = "https"
	}
	return scheme + "://" + m.cfg.Endpoint + "/" + m.cfg.Bucket + "/" + objectName
}
