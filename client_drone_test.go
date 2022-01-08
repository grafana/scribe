package shipwright_test

import (
	"bufio"
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"pkg.grafana.com/shipwright/v1/plumbing"
	"pkg.grafana.com/shipwright/v1/plumbing/testutil"
)

func TestDroneClient(t *testing.T) {
	t.Run("It should generate a simple Drone pipeline",
		testutil.WithTimeout(time.Second*10, func(t *testing.T) {
			var (
				buf     = bytes.NewBuffer(nil)
				errBuff = bytes.NewBuffer(nil)
				ctx     = context.Background()
				path    = filepath.Clean("./demo/basic")
			)

			testutil.RunPipeline(ctx, t, buf, errBuff, &plumbing.Arguments{
				Mode: plumbing.RunModeDrone,
				Path: path,
			})

			t.Log(errBuff.String())

			expected, err := os.Open(filepath.Join(path, "drone_gen.yml"))
			if err != nil {
				t.Fatal(err)
			}

			rScanner := bufio.NewScanner(buf)
			eScanner := bufio.NewScanner(expected)

			for eScanner.Scan() {
				if rScanner.Scan() != true {
					t.Fatal("File size not equal")

					if !bytes.Equal(eScanner.Bytes(), rScanner.Bytes()) {
						t.Fatalf("Lines not equal: \n%s\n%s\n", string(eScanner.Bytes()), string(rScanner.Bytes()))
					}
				}
			}
		}),
	)
}
