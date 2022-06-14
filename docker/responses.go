package docker

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
)

type buildLogs struct {
	Stream string `json:"stream"`
}

func WriteBuildLogs(body io.Reader, out io.Writer) error {
	scanner := bufio.NewScanner(body)
	for scanner.Scan() {
		l := &buildLogs{}
		log.Println(string(scanner.Bytes()))
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

type ImageProgressDetail struct {
	Current int64 `json:"current"`
	Total   int64 `json:"total"`
}

type ImageProgress struct {
	Status         string              `json:"status"`
	ProgressDetail ImageProgressDetail `json:"progressDetail"`
	Progress       string              `json:"progress"`
}

func WriteImageLogs(r io.Reader, out io.Writer) error {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		l := &ImageProgress{}
		if err := json.Unmarshal(scanner.Bytes(), l); err != nil {
			return err
		}
		if l.Status == "Downloading" {
			io.WriteString(out, fmt.Sprintf("%20s | [%5dmb / %5dmb] %s\n", l.Status, l.ProgressDetail.Current/1024/1024, l.ProgressDetail.Total/1024/1024, l.Progress))
		} else {
			io.WriteString(out, fmt.Sprintf("%s\n", l.Status))
		}
	}

	return nil
}
