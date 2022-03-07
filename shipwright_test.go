package shipwright_test

import (
	"reflect"
	"testing"

	"github.com/sirupsen/logrus"
	shipwright "github.com/grafana/shipwright"
	"github.com/grafana/shipwright/plumbing"
	"github.com/grafana/shipwright/plumbing/pipeline"
	"github.com/grafana/shipwright/plumbing/pipeline/clients/cli"
	"github.com/grafana/shipwright/plumbing/pipeline/clients/drone"
	"github.com/grafana/shipwright/plumbing/plog"
)

func TestNew(t *testing.T) {
	t.Run("New should return a CLIClient when provided the -mode=cli flag", func(t *testing.T) {
		cliArgs := []string{"-mode", "cli"}
		args, err := plumbing.ParseArguments(cliArgs)
		if err != nil {
			t.Fatal(err)
		}

		sw := shipwright.NewFromOpts(pipeline.CommonOpts{
			Log:  plog.New(logrus.DebugLevel),
			Args: args,
		})

		if reflect.TypeOf(sw.Client) != reflect.TypeOf(&cli.Client{}) {
			t.Fatalf("shipwright.Client is '%v', not a CLIClient", reflect.TypeOf(sw.Client))
		}

		// Because reflect feels iffy to me, also make sure that it does not equal the same type as a different client
		if reflect.TypeOf(sw.Client) == reflect.TypeOf(&drone.Client{}) {
			t.Fatalf("shipwright.Client is '%v', not a CLIClient", reflect.TypeOf(&drone.Client{}))
		}
	})

	t.Run("New should return a DroneClient when provided the -mode=drone flag", func(t *testing.T) {
		cliArgs := []string{"-mode", "drone"}
		args, err := plumbing.ParseArguments(cliArgs)
		if err != nil {
			t.Fatal(err)
		}

		sw := shipwright.NewFromOpts(pipeline.CommonOpts{
			Log:  plog.New(logrus.DebugLevel),
			Args: args,
		})

		if reflect.TypeOf(sw.Client) != reflect.TypeOf(&drone.Client{}) {
			t.Fatalf("shipwright.Client is '%v', not a DroneClient", reflect.TypeOf(sw.Client))
		}

		// Because reflect feels iffy to me, also make sure that it does not equal the same type as a different client
		if reflect.TypeOf(sw.Client) == reflect.TypeOf(&cli.Client{}) {
			t.Fatalf("shipwright.Client is '%v', not a DroneClient", reflect.TypeOf(&cli.Client{}))
		}
	})
}
