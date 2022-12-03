package fs

import "github.com/grafana/scribe/pipeline"

// FileHasChanged creates a checksum for the file "file" and stores it.
// If the checksum does not exist in the scribe key store, then it will return false.
func FileHasChanged(file string) pipeline.CacheCondition {
	return func() bool {
		return false
	}
}

// Cache will store the directory or file located at `path` if the conditions return true.
// If all of the conditions return true, then the step is skipped and the directory is added to the local filesystem.
func Cache(path string, conditions ...pipeline.CacheCondition) pipeline.Cacher {
	return func(pipeline.Step) {}
}
