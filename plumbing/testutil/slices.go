package testutil

func Int64SlicesEqual(a []int64, b []int64) bool {
	// sort.Slice(a, int64SortFunc(a))
	// sort.Slice(b, int64SortFunc(b))

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func int64SortFunc(int64s []int64) func(i, j int) bool {
	return func(i, j int) bool {
		return int64s[i] < int64s[j]
	}
}
