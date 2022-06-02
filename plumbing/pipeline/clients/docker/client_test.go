package docker

// func TestVolumeValue(t *testing.T) {
// 	wd, _ := os.Getwd()
//
// 	expect := func(t *testing.T, value, expect string) func(t *testing.T) {
// 		return func(t *testing.T) {
// 			t.Helper()
// 			v, err := volumeValue(value)
// 			if err != nil {
// 				t.Fatal(err)
// 			}
//
// 			if v != expect {
// 				t.Fatalf("Unexpected value from 'volumeValue'. Expected '%s', received '%s'", expect, v)
// 			}
// 		}
// 	}
//
// 	t.Run("It should use the same absolute path if an absolute path is provided", expect(t, "/some/absolute/path", "/some/absolute/path:/some/absolute/path"))
// 	t.Run("It should mount the volume relative to /var/scribe if a relative path is provided", expect(t, "./some/relative/path", fmt.Sprintf("%s:/var/scribe/some/relative/path", path.Join(wd, "./some/relative/path"))))
// 	t.Run("It should preserve a fully-formatted volume mount", expect(t, "/absolute/path:/other/path", "/absolute/path:/other/path"))
// }
