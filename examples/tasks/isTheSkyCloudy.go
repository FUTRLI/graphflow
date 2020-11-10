package tasks

import (
	"errors"

	"github.com/futrli/graphflow"
)

// IsTheSkyCloudy is a Task struct
type IsTheSkyCloudy struct {
	graphflow.Task
}

// String returns a description of the Task
func (t *IsTheSkyCloudy) String() string {
	return "Is the sky cloudy?"
}

// Execute retrieves the value of "Sky" from the ExecutionContext and checks whether it's Cloudy
func (t *IsTheSkyCloudy) Execute(ctx *graphflow.ExecutionContext) error {
	sky, ok := ctx.Get("Sky").(string)
	if !ok {
		return errors.New("Failed to retrieve \"Sky\" from ExecutionContext and cast it to string")
	}
	if sky == "Cloudy" {
		t.SetExitPath(graphflow.YES)
	} else {
		t.SetExitPath(graphflow.NO)
	}
	return nil
}
