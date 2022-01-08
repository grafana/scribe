package plumbing

import "strconv"

type OptionalInt struct {
	Value int
	Valid bool
}

func (o *OptionalInt) String() string {
	if o.Valid {
		return strconv.Itoa(o.Value)
	}

	return ""
}

func (o *OptionalInt) Set(v string) error {
	if v == "" {
		return nil
	}

	i, err := strconv.Atoi(v)
	if err != nil {
		return err
	}

	o.Value = i
	o.Valid = true

	return nil
}
