package drone_test

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	shipwright "pkg.grafana.com/shipwright/v1"
	"pkg.grafana.com/shipwright/v1/plumbing"
	"pkg.grafana.com/shipwright/v1/plumbing/clients/drone"
	"pkg.grafana.com/shipwright/v1/plumbing/testutil"
	"pkg.grafana.com/shipwright/v1/plumbing/types"
)

func TestDroneClient(t *testing.T) {
	t.Run("It should generate a simple Drone pipeline",
		testutil.WithTimeout(time.Second*10, func(t *testing.T) {
			var (
				buf          = bytes.NewBuffer(nil)
				errBuff      = bytes.NewBuffer(nil)
				ctx          = context.Background()
				pipelinePath = "../../../demo/basic"
				path         = "./demo/basic"
			)

			testutil.RunPipeline(ctx, t, pipelinePath, io.MultiWriter(buf, os.Stdout), errBuff, &plumbing.Arguments{
				Mode: plumbing.RunModeDrone,
				Path: path,
			})

			expected, err := os.Open(filepath.Join(pipelinePath, "gen_drone.yml"))
			if err != nil {
				t.Fatal(err)
			}

			testutil.ReadersEqual(t, buf, expected)
		}),
	)
}

func TestDroneRun(t *testing.T) {
	t.Run("It should run sequential steps sequentially",
		testutil.WithTimeout(time.Second*5, func(t *testing.T) {
			t.SkipNow()

			t.Log("Creating new drone client...")
			sw := shipwright.NewDroneClient(&types.CommonOpts{})

			t.Log("Creating new test steps...")
			var (
				step1Chan = make(chan bool)
				step1     = testutil.NewTestStep(step1Chan)

				step2Chan = make(chan bool)
				step2     = testutil.NewTestStep(step2Chan)

				step3Chan = make(chan bool)
				step3     = testutil.NewTestStep(step3Chan)
			)

			t.Log("Running steps...")
			sw.Run(step1, step2, step3)

			go func() {
				t.Log("Done()")
				sw.Done()
				t.Log("done with Done()")
			}()

			var (
				expectedOrder = []int{1, 2, 3}
				order         = []int{}
			)

			t.Log("Waiting for order...")
			// Only watch for 3 channels
			for i := 0; i < 3; i++ {
				select {
				case <-step1Chan:
					order = append(order, 1)
				case <-step2Chan:
					order = append(order, 2)
				case <-step3Chan:
					order = append(order, 3)
				}
			}

			if !cmp.Equal(order, expectedOrder) {
				t.Fatal("Steps ran in unexpected order:", cmp.Diff(order, expectedOrder))
			}
		}))
}

func TestDroneTree(t *testing.T) {
	t.Run("It should set the root node once", func(t *testing.T) {
		sw := shipwright.NewDroneClient(&types.CommonOpts{})
		sw.Run(types.NoOpStep)
		client := sw.Client.(*drone.Client)

		if client.List == nil {
			t.Fatal("step list is nil")
		}
	})

	// t.Run("It should populate a new node", func(t *testing.T) {
	// 	sw := shipwright.NewDroneClient(&shipwright.CommonOpts{})
	// 	sw.Run(types.NoOpStep, types.NoOpStep)

	// 	client := sw.Client.(*shipwright.DroneClient)

	// 	if client.Root.FirstChild == nil {
	// 		t.Error("The root node should have a first child")
	// 	}

	// 	if client.Root.LastChild == nil {
	// 		t.Error("The root node should have a last child")
	// 	}

	// 	child := client.Root.FirstChild

	// 	if child.NextSibling != nil || child.PrevSibling != nil {
	// 		t.Error("The child node should not have any siblings")
	// 	}

	// 	if child.LastChild != nil || child.FirstChild != nil {
	// 		t.Error("The child node should not have any children")
	// 	}

	// 	if child.Parent == nil {
	// 		t.Error("The child node should have a parent")
	// 	}
	// })
}
