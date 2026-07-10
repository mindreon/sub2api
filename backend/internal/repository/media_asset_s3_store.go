package repository

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/Wei-Shaw/sub2api/internal/media"
)

// MediaAssetS3Config 是媒体结果转存所需的 S3 兼容存储配置。
type MediaAssetS3Config struct {
	Endpoint        string
	Region          string
	Bucket          string
	AccessKeyID     string
	SecretAccessKey string
	Prefix          string // 对象键前缀，如 "media/"
	PublicBaseURL   string // 公共访问域名（配置则直接拼 URL，否则用预签名）
	ForcePathStyle  bool
}

// IsConfigured 判断转存所需的关键字段是否齐备。
func (c MediaAssetS3Config) IsConfigured() bool {
	return c.Bucket != "" && c.AccessKeyID != "" && c.SecretAccessKey != ""
}

const (
	mediaAssetMaxDownloadBytes = 512 << 20 // 单个视频转存下载上限 512MB
	mediaAssetPresignTTL       = 7 * 24 * time.Hour
	mediaAssetDownloadTimeout  = 5 * time.Minute
)

// S3MediaAssetStore 用 S3 兼容存储实现 media.AssetStore（下载上游视频→转存→签发链接）。
type S3MediaAssetStore struct {
	client        *s3.Client
	bucket        string
	prefix        string
	publicBaseURL string
	http          *http.Client
}

// NewMediaAssetS3Store 构造媒体转存存储。未配置关键字段时返回 (nil, nil)，
// 调用方据此走降级路径（保留上游直链）。
func NewMediaAssetS3Store(ctx context.Context, cfg MediaAssetS3Config) (media.AssetStore, error) {
	if !cfg.IsConfigured() {
		return nil, nil
	}
	region := cfg.Region
	if region == "" {
		region = "auto"
	}
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(region),
		awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}
	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		if cfg.Endpoint != "" {
			o.BaseEndpoint = &cfg.Endpoint
		}
		if cfg.ForcePathStyle {
			o.UsePathStyle = true
		}
		o.APIOptions = append(o.APIOptions, v4.SwapComputePayloadSHA256ForUnsignedPayloadMiddleware)
		o.RequestChecksumCalculation = aws.RequestChecksumCalculationWhenRequired
	})
	return &S3MediaAssetStore{
		client:        client,
		bucket:        cfg.Bucket,
		prefix:        cfg.Prefix,
		publicBaseURL: strings.TrimRight(cfg.PublicBaseURL, "/"),
		http:          &http.Client{Timeout: mediaAssetDownloadTimeout},
	}, nil
}

func (s *S3MediaAssetStore) objectKey(key string) string {
	return s.prefix + key
}

// Rehost 下载 srcURL 并上传到自有存储，返回可访问链接。
func (s *S3MediaAssetStore) Rehost(ctx context.Context, key, srcURL, contentType string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, srcURL, nil)
	if err != nil {
		return "", fmt.Errorf("build download request: %w", err)
	}
	resp, err := s.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("download source: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("download source: unexpected status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, mediaAssetMaxDownloadBytes+1))
	if err != nil {
		return "", fmt.Errorf("read source body: %w", err)
	}
	if len(data) > mediaAssetMaxDownloadBytes {
		return "", fmt.Errorf("source exceeds max size %d bytes", mediaAssetMaxDownloadBytes)
	}

	objKey := s.objectKey(key)
	if contentType == "" {
		contentType = "video/mp4"
	}
	if _, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      &s.bucket,
		Key:         &objKey,
		Body:        bytes.NewReader(data),
		ContentType: &contentType,
	}); err != nil {
		return "", fmt.Errorf("s3 put object: %w", err)
	}

	if s.publicBaseURL != "" {
		return s.publicBaseURL + "/" + objKey, nil
	}
	return s.PresignedURL(ctx, key)
}

// PresignedURL 为已存储对象签发临时可访问链接。
func (s *S3MediaAssetStore) PresignedURL(ctx context.Context, key string) (string, error) {
	if s.publicBaseURL != "" {
		return s.publicBaseURL + "/" + s.objectKey(key), nil
	}
	objKey := s.objectKey(key)
	presign := s3.NewPresignClient(s.client)
	out, err := presign.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: &s.bucket,
		Key:    &objKey,
	}, s3.WithPresignExpires(mediaAssetPresignTTL))
	if err != nil {
		return "", fmt.Errorf("presign url: %w", err)
	}
	return out.URL, nil
}
