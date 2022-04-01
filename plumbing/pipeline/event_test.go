package pipeline_test

import (
	"testing"
)

func TestManualEvents(t *testing.T) {
	// t.Run("A single manual event", func(t *testing.T) {
	// 	// A single manual event should define a filter.
	// 	pipelineEvent := pipeline.NewManualEvent(
	// 		pipeline.NewStringFilter("branch", "main", "dev"),
	// 	)

	// 	event := events.NewEvent("commit", map[string][]string{
	// 		"branch": []string{"main"},
	// 	})

	// 	if !pipelineEvent.Matches(event) {
	// 		t.Fatal("The pipeline event should match the event")
	// 	}
	// })
}
