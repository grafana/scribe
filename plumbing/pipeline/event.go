package pipeline

type Event int

const (
	EventGitPush Event = iota
	EventManual
)
