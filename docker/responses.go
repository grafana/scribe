package docker

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
)

type buildLogs struct {
	Stream string `json:"stream"`
}

func WriteBuildLogs(body io.Reader, out io.Writer) error {
	scanner := bufio.NewScanner(body)
	for scanner.Scan() {
		l := &buildLogs{}
		if err := json.Unmarshal(scanner.Bytes(), l); err != nil {
			return err
		}
		io.WriteString(out, l.Stream)
	}

	return nil
}

func WriteBody(body io.Reader, out io.Writer) error {
	_, err := io.Copy(out, body)
	if err != nil {
		return err
	}

	return nil
}

type pushLogs struct {
	Progress string `json:"progress"`
}

func WritePushLogs(r io.Reader, out io.Writer) error {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		l := &pushLogs{}
		if err := json.Unmarshal(scanner.Bytes(), l); err != nil {
			return err
		}
		io.WriteString(out, fmt.Sprintf("%s\n", l.Progress))
	}

	return nil
}
