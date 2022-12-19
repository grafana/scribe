package state

import (
	"net/url"
	"strings"

	"cloud.google.com/go/storage"
)

// GCSHandler uses the S3 API but with a round-tripper that makes the API client compatible with the S3 api
type GCSHandler struct {
	*ObjectStorageHandler
}

func BucketAndPath(u *url.URL) (string, string) {
	return u.Host, strings.TrimPrefix(u.Path, "/")
}

func NewGCSHandler(client *storage.Client, u *url.URL) (*GCSHandler, error) {
	bucket, path := BucketAndPath(u)

	h := NewObjectStorageHandler(
		&GCSObjectStorage{
			Client: client,
		},
		bucket,
		path,
	)

	return &GCSHandler{
		ObjectStorageHandler: h,
	}, nil

}
