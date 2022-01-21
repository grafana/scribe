package plumbing

import "errors"

var (
	ErrorMissingArgument = errors.New("argument requested but not provided")
)
