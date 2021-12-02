package git

type Repository struct{}

type DescribeOpts struct {
	Tags   bool
	Dirty  bool
	Always bool
}

func (r *Repository) Describe(opts *DescribeOpts) string {
	return ""
}
