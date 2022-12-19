package state

import (
	"context"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3ObjectStorage struct {
	Client *s3.Client
}

func (s *S3ObjectStorage) GetObject(ctx context.Context, bucket, key string) (*GetObjectResponse, error) {
	res, err := s.Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		return nil, err
	}

	return &GetObjectResponse{
		Body: res.Body,
	}, nil
}

func (s *S3ObjectStorage) PutObject(ctx context.Context, bucket, key string, body io.Reader) error {
	if _, err := s.Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   body,
	}); err != nil {
		return err
	}

	return nil
}
