package pipeline

// New creates a new Step that represents a pipeline.
func New(name string, id int64) Step[Pipeline] {
	return Step[Pipeline]{
		Name:   name,
		Serial: id,
	}
}
