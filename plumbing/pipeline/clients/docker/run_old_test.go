package docker

// import (
// 	"strings"
// 	"testing"
// )
//
// func TestRunArgs(t *testing.T) {
// 	opts := RunOpts{
// 		Image:   "golang:latest",
// 		Command: "go",
// 		Args:    []string{"test", "./..."},
// 		Env:     []string{"ENV_1=1", "ENV_2=2"},
// 		Volumes: []string{"/var/volume1:/usr/share/volume1", "/var/volume2:/usr/share/volume2"},
// 	}
//
// 	expect := strings.Split("run --rm -v /var/volume1:/usr/share/volume1 -v /var/volume2:/usr/share/volume2 -e ENV_1=1 -e ENV_2=2 golang:latest go test ./...", " ")
//
// 	val := RunArgs(opts)
// 	t.Logf("Comparing (received) '%+v' to (expected) '%+v'", val, expect)
//
// 	for i, v := range expect {
// 		if val[i] != v {
// 			t.Fatalf("got '%s' at position '%d', expected '%s'", val[i], i, v)
// 		}
// 	}
// }
