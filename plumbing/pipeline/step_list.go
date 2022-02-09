package pipeline

// List is a very primitive implementation of the Collection interface
// Each element in a single row is considered a child of the row above it.
// There are no leafs.
type List struct {
	rows [][]Step
}

func NewList() *List {
	return &List{
		rows: make([][]Step, 0),
	}
}

// Append adds a single step to the bottom of this tree.
// The default behavior of Append is to add the provided as a dependency of every step in the bottom row of the tree
func (t *List) Append(steps ...Step) error {
	if len(t.rows) != 0 {
		// Each element in the row before this one is a dependency of every step that was just added.
		for i := range steps {
			row := t.rows[len(t.rows)-1]
			steps[i].Dependencies = row
		}
	}

	t.rows = append(t.rows, steps)

	return nil
}

func (t *List) AppendLineage(steps ...Step) error {
	for _, v := range steps {
		if err := t.Append(v); err != nil {
			return err
		}
	}

	return nil
}

func (t *List) Walk(wf func(Step) error) error {
	for _, row := range t.rows {
		for _, col := range row {
			if err := wf(col); err != nil {
				return err
			}
		}
	}

	return nil
}

func (t *List) String() string {
	return ""
}
