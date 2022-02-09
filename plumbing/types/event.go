package types

type Event int

const (
	EventGitPush Event = iota
	EventManual
)
