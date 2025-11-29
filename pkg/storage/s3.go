package storage

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"github.com/yourusername/golf_messenger/internal/config"
)

type S3Client struct {
	client     *s3.Client
	bucketName string
}

func NewS3Client(cfg *config.AWSConfig) (*S3Client, error) {
	ctx := context.Background()

	var awsCfg aws.Config
	var err error

	if cfg.S3Endpoint != "" {
		awsCfg, err = config.LoadDefaultConfig(ctx,
			config.WithRegion(cfg.Region),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
				cfg.AccessKeyID,
				cfg.SecretAccessKey,
				"",
			)),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to load AWS config: %w", err)
		}

		s3Client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(cfg.S3Endpoint)
			o.UsePathStyle = true
		})

		return &S3Client{
			client:     s3Client,
			bucketName: cfg.S3BucketName,
		}, nil
	}

	awsCfg, err = config.LoadDefaultConfig(ctx,
		config.WithRegion(cfg.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.AccessKeyID,
			cfg.SecretAccessKey,
			"",
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	s3Client := s3.NewFromConfig(awsCfg)

	return &S3Client{
		client:     s3Client,
		bucketName: cfg.S3BucketName,
	}, nil
}

func (s *S3Client) UploadFile(ctx context.Context, file io.Reader, filename string, contentType string) (string, error) {
	ext := filepath.Ext(filename)
	key := fmt.Sprintf("avatars/%s%s", uuid.New().String(), ext)

	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucketName),
		Key:         aws.String(key),
		Body:        file,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file to S3: %w", err)
	}

	url := fmt.Sprintf("https://%s.s3.amazonaws.com/%s", s.bucketName, key)
	return url, nil
}

func (s *S3Client) DeleteFile(ctx context.Context, fileURL string) error {
	key, err := s.extractKeyFromURL(fileURL)
	if err != nil {
		return err
	}

	_, err = s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete file from S3: %w", err)
	}

	return nil
}

func (s *S3Client) extractKeyFromURL(fileURL string) (string, error) {
	baseURL := fmt.Sprintf("https://%s.s3.amazonaws.com/", s.bucketName)
	if len(fileURL) <= len(baseURL) {
		return "", fmt.Errorf("invalid S3 URL format")
	}

	key := fileURL[len(baseURL):]
	return key, nil
}
