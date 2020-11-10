package tasks

import "github.com/futrli/graphflow"

// ForecastSun is a Task struct
type ForecastSun struct {
	graphflow.Task
}

// String returns a description of the Task
func (t *ForecastSun) String() string {
	return "Forecast Sun"
}

// Execute sets the ExecutionContext's Forecast value to "Sun"
func (t *ForecastSun) Execute(ctx *graphflow.ExecutionContext) error {
	ctx.Set("Forecast", "Sun")
	return nil
}
