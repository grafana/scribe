package state

// Without returns a list of Arguments equal to `args` without the args in the `exclude` list.
// complexity o(n^2)
func Without(args []Argument, exclude []Argument) []Argument {
	ret := []Argument{}
	for _, v := range args {
		found := false
		for _, excl := range exclude {
			if v == excl {
				found = true
			}
		}
		if !found {
			ret = append(ret, v)
		}
	}

	return ret
}

// EqualArgs checks that the argument list a and b are equal.
// We go out of our way to check equality despite the order.
// complexity o(n^2)
func EqualArgs(a []Argument, b []Argument) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		found := false
		for n := range b {
			if a[i] == b[n] {
				found = true
			}
		}
		if !found {
			return false
		}
	}

	return true
}
