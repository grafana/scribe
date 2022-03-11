package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/mitchellh/mapstructure"
	"gopkg.in/yaml.v2"
)

type DroneBuild struct {
	Pipelines []Pipeline
}

type Service struct {
	Name        string                  `yaml:"name"`
	Image       string                  `yaml:"image"`
	Volumes     []StepVolume            `yaml:"volumes"`
	Environment *map[string]interface{} `yaml:"environment"`
}

type Node struct {
	Type string `yaml:"type"`
}

type Platform struct {
	Arch string `yaml:"arch"`
	OS   string `yaml:"os"`
}

type PipelineVolume struct {
	Host struct {
		Path string `yaml:"path"`
	} `yaml:"host"`
	Name string `yaml:"name"`
	Temp struct {
		Medium string `yaml:"medium"`
	} `yaml:"temp"`
}

type StepVolume struct {
	Path string `yaml:"path"`
	Name string `yaml:"name"`
}

type Trigger struct {
	Event interface{} `yaml:"event"`
	Paths struct {
		Include []string `yaml:"include"`
		Exclude []string `yaml:"exclude"`
	} `yaml:"paths"`
	Branch string        `yaml:"branch"`
	Repo   []interface{} `yaml:"repo"`
	Cron   string        `yaml:"cron"`
}

type Pipeline struct {
	DependsOn []string         `yaml:"depends_on"`
	Kind      string           `yaml:"kind"`
	Name      string           `yaml:"name"`
	Node      Node             `yaml:"node"`
	Platform  Platform         `yaml:"platform"`
	Services  []Service        `yaml:"services"`
	Steps     []Step           `yaml:"steps"`
	Trigger   Trigger          `yaml:"trigger"`
	Type      string           `yaml:"type"`
	Volumes   []PipelineVolume `yaml:"volumes"`
}

type StepSettings struct {
	Params       []string `yaml:"params"`
	Repositories []string `yaml:"repositories"`
	Server       string   `yaml:"server"`
	Token        struct {
		FromSecret string `yaml:"from_secret"`
	} `yaml:"token"`
}

type Step struct {
	Commands    []string                `yaml:"commands"`
	Image       string                  `yaml:"image"`
	Name        string                  `yaml:"name"`
	DependsOn   *[]string               `yaml:"depends_on"`
	Environment *map[string]interface{} `yaml:"environment"`
	Detach      *bool                   `yaml:"detach"`
	When        *struct {
		Status []string `yaml:"status"`
	} `yaml:"when"`
	Failure  *string       `yaml:"failure"`
	Volumes  *[]StepVolume `yaml:"volumes"`
	Settings *StepSettings `yaml:"settings"`
}

type Secret struct {
	Kind string `yaml:"kind"`
	Name string `yaml:"name"`
	Get  struct {
		Name string `yaml:"name"`
		Path string `yaml:"path"`
	} `yaml:"get"`
}
type Signature struct {
	Kind string `yaml:"kind"`
	HMAC string `yaml:"hmac"`
}

type Build struct {
	Pipelines []Pipeline
	Secrets   []Secret
	Signature Signature
}

func (s *Step) CamelName() string {
	split := func(r rune) bool {
		return strings.ContainsRune(" _-", r)
	}
	p := strings.FieldsFunc(s.Name, split)
	c := []string{}
	for _, part := range p {
		c = append(c, strings.Title(part))
	}
	return strings.Join(c, "")
}

func parseYAML(filename string) (*Build, error) {
	build := Build{}
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	r := bytes.NewReader(data)
	dec := yaml.NewDecoder(r)
	var msi map[string]interface{}
	for dec.Decode(&msi) == nil {
		kind := msi["kind"].(string)
		switch kind {
		case "pipeline":
			var pipeline Pipeline
			err = mapstructure.WeakDecode(msi, &pipeline)
			if err != nil {
				return nil, err
			}
			build.Pipelines = append(build.Pipelines, pipeline)

		case "secret":
			var secret Secret
			err = mapstructure.WeakDecode(msi, &secret)
			if err != nil {
				return nil, err
			}
			build.Secrets = append(build.Secrets, secret)
		case "signature":
			var signature Signature
			err = mapstructure.Decode(msi, &signature)
			if err != nil {
				return nil, err
			}
			build.Signature = signature
		}
	}
	return &build, nil
}

func renderYAML(build Build) error {
	enc := yaml.NewEncoder(os.Stdout)
	for _, pipeline := range build.Pipelines {
		err := enc.Encode(pipeline)
		if err != nil {
			return err
		}
	}
	for _, secret := range build.Secrets {
		err := enc.Encode(secret)
		if err != nil {
			return err
		}
	}
	return enc.Encode(build.Signature)
}

const header = `package main
import (
	"github.com/grafana/shipwright"
	"github.com/grafana/shipwright/exec"
)

func main() {
	sw := shipwright.New()`

func renderGolang(build Build) error {
	fmt.Println(header)

	for _, pipeline := range build.Pipelines {
		for _, step := range pipeline.Steps {
			camel := step.CamelName()
			if len(step.Commands) > 0 {
				fmt.Printf("\n    step%s := exec.Run(\n", camel)
				for _, cmd := range step.Commands {
					fmt.Printf("      \"%s\",\n", cmd)
				}
				fmt.Println("    )")
			} else {
				fmt.Printf("\n    step%s := exec.Noop()\n", camel)
			}
			if step.Image != "" {
				fmt.Printf("      .WithImage(\"%s\")\n", step.Image)
			}
			if step.Name != "" {
				fmt.Printf("      .WithName(\"%s\")\n", step.Name)
			}
			if step.Environment != nil && len(*step.Environment) > 0 {
				fmt.Println("      .WithEnvironment(map[string]string{")
				for k, v := range *step.Environment {
					fmt.Printf("      \"%s\": \"%s\",\n", k, v)
				}
				fmt.Println("    })")
			}
		}
		fmt.Println("\n    sw.Run(")
		for _, step := range pipeline.Steps {
			fmt.Printf("      step%s,\n", step.CamelName())
		}
		fmt.Println("    )")
	}
	fmt.Println("}")
	return nil
}

func main() {
	droneYAML := "/home/malcolm/grafana/grafana/.drone.yml"
	build, err := parseYAML(droneYAML)
	if err != nil {
		panic(err)
	}
	/*
		err = renderYAML(*build)
		if err != nil {
			panic(err)
		}
	*/
	err = renderGolang(*build)
	if err != nil {
		panic(err)
	}
}
