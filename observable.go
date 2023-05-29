package scribe

type Observable interface {
	C() chan bool
}
