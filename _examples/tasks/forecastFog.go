package tasks

import "github.com/futrli/graphflow"

//ForecastFog is a Task struct
type ForecastFog struct {
	graphflow.Task
}

// String returns a description of the Task
func (t *ForecastFog) String() string {
	return "Forecast Fog"
}

// Execute sets the ExecutionContext's Forecast value to "Fog"
func (t *ForecastFog) Execute(ctx *graphflow.ExecutionContext) error {
	ctx.Set("Forecast", "Fog")
	return nil
}
