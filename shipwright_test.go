package shipwright_test

import (
	"reflect"
	"testing"

	shipwright "pkg.grafana.com/shipwright/v1"
)

func TestNew(t *testing.T) {
	t.Run("New should return a CLIClient when provided the -mode=cli flag", func(t *testing.T) {
		args := []string{"-mode", "cli"}
		opts, err := shipwright.ParseCLIOpts(args)
		if err != nil {
			t.Fatal(err)
		}

		sw := shipwright.NewFromOpts(opts)

		if reflect.TypeOf(sw.Client) != reflect.TypeOf(&shipwright.CLIClient{}) {
			t.Fatalf("shipwright.Client is '%v',  not a CLIClient", reflect.TypeOf(sw.Client))
		}

		// Because reflect feels iffy to me, also make sure that it does not equal the same type as a different client
		if reflect.TypeOf(sw.Client) == reflect.TypeOf(&shipwright.ConfigClient{}) {
			t.Fatalf("shipwright.Client is '%v', not a CLIClient", reflect.TypeOf(&shipwright.ConfigClient{}))
		}
	})

	t.Run("New should return a ConfigClient when provided the -mode=config flag", func(t *testing.T) {
		args := []string{"-mode", "config"}
		opts, err := shipwright.ParseCLIOpts(args)
		if err != nil {
			t.Fatal(err)
		}

		sw := shipwright.NewFromOpts(opts)

		if reflect.TypeOf(sw.Client) != reflect.TypeOf(&shipwright.ConfigClient{}) {
			t.Fatalf("shipwright.Client is '%v',  not a ConfigClient", reflect.TypeOf(sw.Client))
		}

		// Because reflect feels iffy to me, also make sure that it does not equal the same type as a different client
		if reflect.TypeOf(sw.Client) == reflect.TypeOf(&shipwright.CLIClient{}) {
			t.Fatalf("shipwright.Client is '%v', not a ConfigClient", reflect.TypeOf(&shipwright.ConfigClient{}))
		}
	})
}
