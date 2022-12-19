package state

import (
	"context"
	"io"
)

type GetObjectResponse struct {
	Body io.ReadCloser
}

type ObjectStorage interface {
	GetObject(ctx context.Context, bucket, key string) (*GetObjectResponse, error)
	PutObject(ctx context.Context, bucket, key string, body io.Reader) error
}
