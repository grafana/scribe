package fs

import "pkg.grafana.com/shipwright/v1/types"

// FileHasChanged creates a checksum for the file "file" and stores it.
// If the checksum does not exist in the shipwright key store, then it will return false.
func FileHasChanged(file string) types.CacheCondition {
	return func() bool {
		return false
	}
}

// Cache will store the directory or file located at `path` if the conditions return true.
// If all of the conditions return true, then the step is skipped and the directory is added to the local filesystem.
func Cache(path string, conditions ...types.CacheCondition) types.Cacher {
	return func(types.Step) {}
}
