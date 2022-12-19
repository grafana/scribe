package state_test

import (
	"net/url"
	"testing"

	"github.com/grafana/scribe/state"
)

func TestBucketAndPath(t *testing.T) {
	type result struct {
		bucket, path string
	}
	res := map[string]result{
		"gs://bucket/path":           {"bucket", "path"},
		"gs://bucket/path/1/2/3":     {"bucket", "path/1/2/3"},
		"gs://the-bucket/path/1/2/3": {"the-bucket", "path/1/2/3"},
	}

	for k, v := range res {
		u, _ := url.Parse(k)
		bucket, path := state.BucketAndPath(u)
		if bucket != v.bucket {
			t.Errorf("got: '%s', expected: '%s'", bucket, v.bucket)
		}
		if path != v.path {
			t.Errorf("got: '%s', expected: '%s'", path, v.path)
		}
	}
}
