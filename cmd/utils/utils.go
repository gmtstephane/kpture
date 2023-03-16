package utils

import (
	"fmt"
	"io"
	"os"
)

const DefaultKubePath = "/dev/termination-log"

type CommandExample struct {
	Command     string
	Title       string
	Additionnal string
}

func CommandMardkown(commands []CommandExample) string {
	ret := "### Examples:\n"
	for _, c := range commands {
		ret += fmt.Sprintf("#### %s\n", c.Title)
		ret += fmt.Sprintf("```bash\n%s\n```\n", c.Command)
		if c.Additionnal != "" {
			ret += fmt.Sprintf("%s\n", c.Additionnal)
		}
	}
	return ret
}

type TerminationWriter struct {
	messagePath bool
	writer      io.Writer
}

func NewTerminationWriter(messagePath bool, path string) (*TerminationWriter, error) {
	if messagePath {
		f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0o666)
		if err != nil {
			return nil, err
		}
		return &TerminationWriter{messagePath: messagePath, writer: f}, nil
	}
	return &TerminationWriter{messagePath: messagePath}, nil
}

func (t *TerminationWriter) TerminationMessage(e error) error {
	if t.messagePath {
		_, err := t.writer.Write([]byte(e.Error()))
		if err != nil {
			return fmt.Errorf("%w : %s", err, e.Error())
		}
	}
	return e
}
