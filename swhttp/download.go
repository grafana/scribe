package swhttp

import (
	"bytes"
	"context"
	"io"
	"net/http"
)

// Download downloads the file at the provided URL using the default client. It returns the response in a bytes.Buffer.
func Download(ctx context.Context, url string) (*bytes.Buffer, error) {
	return DownloadWithClient(ctx, DefaultClient, url)
}

// DownloadWithClient downloads the file at the provided URL using the provided client. It returns the response in a bytes.Buffer if successful.
func DownloadWithClient(ctx context.Context, client http.Client, url string) (*bytes.Buffer, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	res, err := client.Do(req)
	if err := HandleResponse(res, err); err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, res.Body); err != nil {
		return nil, err
	}

	return buf, nil
}
