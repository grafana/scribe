package pipeline

import (
	"bufio"
	"fmt"
	"io"
)

type StdinReader struct {
	out io.Writer
	in  io.Reader
}

func NewStdinReader(in io.Reader, out io.Writer) *StdinReader {
	return &StdinReader{
		out: out,
		in:  in,
	}
}

func (s *StdinReader) Get(key string) (string, error) {
	fmt.Fprintf(s.out, "Argument '%[1]s' requested but not found. Please provide a value for '%[1]s': ", key)
	// Prompt for the value via stdin since it was not found
	scanner := bufio.NewScanner(s.in)
	scanner.Scan()

	if err := scanner.Err(); err != nil {
		return "", err
	}

	value := scanner.Text()
	fmt.Fprintf(s.out, "In the future, you can provide this value with the '-arg=%s=%s' argument\n", key, value)
	return value, nil
}
