package starlark

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/drone/drone-yaml/yaml"
)

type Expected struct {
	buf bytes.Buffer
}

func (e *Expected) Add(line string) {
	e.buf.WriteString(line + "\n")
}
func (e *Expected) AddNoNewline(line string) {
	e.buf.WriteString(line)
}
func (e *Expected) String() string {
	return e.buf.String()
}

func samplePipeline() (*yaml.Pipeline, string) {
	step, expectedStep := sampleStep()

	expected := Expected{}
	expected.Add(`def sample_pipeline():`)
	expected.Add(`  return {`)
	expected.Add(`    "kind": "pipeline",`)
	expected.Add(`    "type": "docker",`)
	expected.Add(`    "name": "sample",`)
	expected.Add(`    "platform": {`)
	expected.Add(`      "os": "linux",`)
	expected.Add(`      "arch": "amd64",`)
	expected.Add(`    },`)
	expected.Add(`    "steps": [`)
	expected.Add(`      sample_step(),`)
	expected.Add(`    ],`)
	expected.Add(`  }`)
	expected.Add(``)
	expected.AddNoNewline(expectedStep)

	return &yaml.Pipeline{
		Name: "sample",
		Kind: "pipeline",
		Type: "docker",
		Platform: yaml.Platform{
			OS:   "linux",
			Arch: "amd64",
		},
		Steps: []*yaml.Container{
			step,
		},
	}, expected.String()
}

func sampleStep() (*yaml.Container, string) {
	expected := Expected{}
	expected.Add(`def sample_step():`)
	expected.Add(`  return {`)
	expected.Add(`    "commands": [`)
	expected.Add(`      "do",`)
	expected.Add(`      "this",`)
	expected.Add(`      "then",`)
	expected.Add(`      "that",`)
	expected.Add(`    ],`)
	expected.Add(`    "environment": {`)
	expected.Add(`      "KEY": "this-value",`)
	expected.Add(`      "SECRET": {`)
	expected.Add(`        "from_secret": "secret",`)
	expected.Add(`      },`)
	expected.Add(`    },`)
	expected.Add(`    "name": "sample-step",`)
	expected.Add(`  }`)
	expected.Add(``)

	return &yaml.Container{
		Name:     "sample-step",
		Commands: []string{"do", "this", "then", "that"},
		Environment: map[string]*yaml.Variable{
			"KEY": {
				Value: "this-value",
			},
			"SECRET": {
				Secret: "secret",
			},
		},
	}, expected.String()
}

func TestMarshalString(t *testing.T) {
	p, _ := samplePipeline()
	v := reflect.ValueOf(p.Kind)
	//ty := reflect.TypeOf(v)

	sl := NewStarlark()
	sl.MarshalString(v)
	assertString(t, sl.String(), `"pipeline",`+"\n")
}

// func TestMarshalStep(t *testing.T) {
// 	step, expectedStep := sampleStep()
//
// 	s := NewStarlark()
// 	s.MarshalStep(step)
// 	assertString(t, s.String(), expectedStep)
// }

func TestMarshalPipeline(t *testing.T) {
	pipeline, expectedPipeline := samplePipeline()

	s := NewStarlark()
	s.MarshalPipeline(pipeline)
	assertString(t, s.String(), expectedPipeline)
}

func assertString(t *testing.T, got, expected string) {
	t.Helper()
	if got != expected {
		t.Errorf("Expected: <%s>\nGot: <%s>", expected, got)
	}
}
