package state_test

import (
	"testing"

	"github.com/grafana/scribe/pipeline"
	"github.com/grafana/scribe/state"
)

func TestEqual(t *testing.T) {
	var (
		arg1 = state.NewInt64Argument("1")
		arg2 = state.NewStringArgument("2")
		arg3 = state.NewDirectoryArgument("3")
		arg4 = state.NewFileArgument("4")
	)
	t.Run("simpel equality check", func(t *testing.T) {
		a := []state.Argument{arg1, arg2, arg3, arg4}
		b := []state.Argument{arg1, arg2, arg3, arg4}

		if !state.EqualArgs(a, b) {
			t.Errorf("%+v != %+v", a, b)
		}
	})
	t.Run("unequal ordering, same elements", func(t *testing.T) {
		a := []state.Argument{arg4, arg3, arg2, arg1}
		b := []state.Argument{arg1, arg2, arg3, arg4}

		if !state.EqualArgs(a, b) {
			t.Errorf("%+v != %+v", a, b)
		}
	})
}

func TestWithout(t *testing.T) {
	var (
		arg1 = state.NewInt64Argument("1")
		arg2 = state.NewInt64Argument("2")
		arg3 = state.NewInt64Argument("3")
		arg4 = state.NewInt64Argument("4")
		arg5 = state.NewInt64Argument("5")
		arg6 = state.NewInt64Argument("6")

		in = []state.Argument{arg1, arg2, arg3, arg4, arg5, arg6}
	)

	t.Run("simple removal", func(t *testing.T) {
		res := state.Without(in, []state.Argument{arg4, arg5, arg6})
		ex := []state.Argument{arg1, arg2, arg3}
		if !state.EqualArgs(res, ex) {
			t.Errorf("%v != %v", res, ex)
		}
	})

	t.Run("complex removal", func(t *testing.T) {
		res := state.Without(in, []state.Argument{arg1, arg3, arg5})
		ex := []state.Argument{arg2, arg4, arg6}
		if !state.EqualArgs(res, ex) {
			t.Errorf("%v != %v", res, ex)
		}
	})

	t.Run("removal with more exclusions than in the list", func(t *testing.T) {
		res := state.Without(in, []state.Argument{arg1, arg2, arg3, arg5, arg6})
		ex := []state.Argument{arg4}
		if !state.EqualArgs(res, ex) {
			t.Errorf("%v != %v", res, ex)
		}
	})

	t.Run("2022-10-12 bug, ArgumentSourceFS not being excluded", func(t *testing.T) {
		// '[{... Key:source}]' without '[{... Key:build-id} {... Key:source} {... Key:pipeline-go-mod} {... Key:docker-socket} {... Key:workdir}]' is '[{... Key:source}]'
		in := []state.Argument{pipeline.ArgumentSourceFS}
		ex := []state.Argument{}
		res := state.Without(in, []state.Argument{pipeline.ArgumentBuildID, pipeline.ArgumentSourceFS, pipeline.ArgumentPipelineGoModFS, pipeline.ArgumentDockerSocketFS})
		if !state.EqualArgs(ex, res) {
			t.Errorf("%v != %v", res, ex)
		}
	})
}
