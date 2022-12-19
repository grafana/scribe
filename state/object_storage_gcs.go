package state

import (
	"context"
	"errors"
	"io"

	"cloud.google.com/go/storage"
)

type GCSObjectStorage struct {
	Client *storage.Client
}

func (s *GCSObjectStorage) GetObject(ctx context.Context, bucket, key string) (*GetObjectResponse, error) {
	obj := s.Client.Bucket(bucket).Object(key)
	r, err := obj.NewReader(ctx)
	if err != nil {
		if errors.Is(err, storage.ErrObjectNotExist) {
			return nil, ErrorFileNotFound
		}
		return nil, err
	}

	return &GetObjectResponse{
		Body: r,
	}, nil
}

func (s *GCSObjectStorage) PutObject(ctx context.Context, bucket, key string, body io.Reader) error {
	obj := s.Client.Bucket(bucket).Object(key)
	w := obj.NewWriter(ctx)
	if _, err := io.Copy(w, body); err != nil {
		return err
	}

	return w.Close()
}
