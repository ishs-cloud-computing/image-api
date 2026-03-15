package service

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Service struct {
	client *s3.Client
	bucket string
}

func NewS3Service(cfg aws.Config, bucket string) *S3Service {
	return &S3Service{
		client: s3.NewFromConfig(cfg),
		bucket: bucket,
	}
}

func (s *S3Service) Upload(ctx context.Context, key string, body io.Reader, ContentType string) (string, error) {
	log.Printf("[S3] PutObject: bucket=%s, key=%s, contentType=%s", s.bucket, key, ContentType)

	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        body,
		ContentType: aws.String(ContentType),
	})
	if err != nil {
		return "", fmt.Errorf("S3 업로드 실패: %w", err)
	}

	url := fmt.Sprintf("https://%s.s3.amazonaws.com/%s", s.bucket, key)
	log.Printf("[S3] PutObject 완료: %s", url)
	return url, nil
}

func (s *S3Service) Download(ctx context.Context, key string) (io.ReadCloser, string, error) {
	log.Printf("[S3] GetObject: bucket=%s, key=%s", s.bucket, key)

	result, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, "", fmt.Errorf("S3 다운로드 실패: %w", err)
	}

	contentType := "application/octet-stream"
	if result.ContentType != nil {
		contentType = *result.ContentType
	}

	log.Printf("[S3] GetObject 완료: key=%s, contentType=%s", key, contentType)
	return result.Body, contentType, nil
}
