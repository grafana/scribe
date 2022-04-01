package pipeline

import (
	"fmt"
	"log"
	"regexp"
)

type FilterValue[T string | *regexp.Regexp] struct {
	v T
}

func (f *FilterValue[T]) String() string {
	switch x := any(f.v).(type) {
	case string:
		return x
	case *regexp.Regexp:
		return x.String()
	}

	return "Unknown type"
}

func StringFilter(v string) *FilterValue[string] {
	return &FilterValue[string]{
		v: v,
	}
}

func RegexpFilter(v *regexp.Regexp) *FilterValue[*regexp.Regexp] {
	return &FilterValue[*regexp.Regexp]{
		v: v,
	}
}

// Event is provided when defining a Shipwright pipeline to define the events that cause the pipeline to be ran.
// Some example events that might cause pipelines to be created:
// * Manual events with user input, like 'Promotions' in Drone. In this scenario, the user may have the ability to supply any keys/values as arguments, however, pipeline developers in Shipwright should be able to specifically define what fields are accepted. See https://docs.drone.io/promote/.
// * git and SCM-related events like 'Pull Reuqest', 'Commit', 'Tag'. Each one of these events has a unique set of arguments / filters. `Commit` may allow pipeline developers to filter by branch or message. Tags may allow developers to filter by name.
// * cron events, which may allow the pipeline in the CI service to be ran on a schedule.
// The Event type stores both the filters and a list of values that it provides to the pipeline.
// Client implementations of the pipeline (type Client) are expected to handle events that they are capable of handling.
// 'Handling' events means that the the arguments in the `Provides` key should be available before any first steps are ran. It will not typically be up to pipeline developers to decide what arguments an event provides.
// The only case where this may happen is if the event is a manual one, where users are able to submit the event with any arbitrary set of keys/values.
// The 'Filters' key is provided in the pipeline code and should not be populated when pre-defined in the Shipwright package.
type Event struct {
	Filters  map[string]fmt.Stringer
	Provides []Argument
}

type GitCommitFilters[T string | *regexp.Regexp] struct {
	Branch *FilterValue[T]
}

func GitCommitEvent[T string | *regexp.Regexp](filters GitCommitFilters[T]) Event {
	f := map[string]fmt.Stringer{}

	if filters.Branch != nil {
		f["branch"] = filters.Branch
	}

	return Event{
		Filters: f,
		Provides: []Argument{
			ArgumentCommitSHA,
			ArgumentBranch,
			ArgumentRemoteURL,
		},
	}
}

type GitTagFilters[T string | *regexp.Regexp] struct {
	Name *FilterValue[T]
}

func GitTagEvent[T string | *regexp.Regexp](filters GitTagFilters[T]) Event {
	f := map[string]fmt.Stringer{}
	log.Println("got git tag event...", filters)
	if filters.Name != nil {
		f["tag"] = filters.Name
	}

	return Event{
		Filters: f,
		Provides: []Argument{
			ArgumentCommitSHA,
			ArgumentCommitRef,
			ArgumentRemoteURL,
		},
	}
}
