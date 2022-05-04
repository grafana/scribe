package starlark

import (
	"bytes"
	"fmt"
	"strings"
)

type Starlark struct {
	buf   *bytes.Buffer
	index int
}

func (s *Starlark) Indent(change int) {
	if change < 0 {
		s.index += change
	}
	if s.index > 0 {
		s.buf.WriteString(strings.Repeat(" ", s.index))
	}
	if change > 0 {
		s.index += change
	}
}

func (s *Starlark) methodName(name, suffix string) string {
	r := strings.NewReplacer("-", "_", " ", "_", ":", "")
	name = r.Replace(name)
	if suffix != "" && !strings.HasSuffix(name, "_"+suffix) {
		name += "_" + suffix
	}
	return name
}

func (s *Starlark) MethodStart(name string) {
	s.Indent(2)
	s.buf.WriteString(fmt.Sprintf("def %s():\n", name))
}

func (s *Starlark) MethodEnd() {
	s.Indent(-2)
	s.buf.WriteString("\n")
}

func (s *Starlark) MethodCall(name, suffix string) {
	methodName := s.methodName(name, suffix)
	s.Indent(0)

	s.buf.WriteString(fmt.Sprintf("%s(),\n", methodName))
}

func (s *Starlark) Return() {
	s.Indent(0)
	s.buf.WriteString("return ")
}

func (s *Starlark) StartDict(indent bool) {
	if indent {
		s.Indent(2)
	} else {
		s.index += 2
	}
	s.buf.WriteString("{\n")
}

func (s *Starlark) DictFieldName(name string) {
	s.Indent(0)
	s.buf.WriteString(fmt.Sprintf(`"%s": `, name))
}

func (s *Starlark) EndDict(comma bool) {
	s.Indent(-2)
	if comma {
		s.buf.WriteString("},\n")
	} else {
		s.buf.WriteString("}\n")
	}
}

func (s *Starlark) StartArray() {
	s.index += 2
	s.buf.WriteString("[\n")
}

func (s *Starlark) EndArray() {
	s.Indent(-2)
	s.buf.WriteString("],\n")
}

func (s *Starlark) Bytes() []byte {
	return s.buf.Bytes()
}

func (s *Starlark) String() string {
	return s.buf.String()
}
