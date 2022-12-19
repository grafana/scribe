package state

import (
	"net/url"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Handler struct {
	*ObjectStorageHandler
}

func NewS3Handler(client *s3.Client, u *url.URL) (*S3Handler, error) {
	h := NewObjectStorageHandler(
		&S3ObjectStorage{
			Client: client,
		},
		u.Host,
		u.Path,
	)

	return &S3Handler{
		ObjectStorageHandler: h,
	}, nil
}
