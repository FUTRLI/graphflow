package tasks

import "github.com/futrli/graphflow"

// ForecastRain is a Task struct
type ForecastRain struct {
	graphflow.Task
}

// String returns a description of the Task
func (t *ForecastRain) String() string {
	return "Forecast Rain"
}

// Execute sets the ExecutionContext's Forecast value to "Rain"
func (t *ForecastRain) Execute(ctx *graphflow.ExecutionContext) error {
	ctx.Set("Forecast", "Rain")
	return nil
}
