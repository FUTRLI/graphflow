package rendering

import (
	"github.com/futrli/graphflow"
	"github.com/stretchr/testify/assert"
	"testing"
)

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
	sky := ctx.Get("Sky").(string)
	if sky == "Cloudy" {
		t.SetExitPath(graphflow.YES)
	} else {
		t.SetExitPath(graphflow.NO)
	}
	return nil
}

// TaskWithNoName is a Task struct
type TaskWithNoName struct {
	graphflow.Task
}

// Execute sets the ExecutionContext's Forecast value to "Sun"
func (t *TaskWithNoName) Execute(ctx *graphflow.ExecutionContext) error {
	ctx.Set("Forecast", "Nothing")
	return nil
}

func buildGraphflow() *graphflow.Graphflow {
	gf := new(graphflow.Graphflow)

	// create task instances
	start := gf.AddTask(new(graphflow.StartTask))
	isTheSkyCloudy := gf.AddTask(new(IsTheSkyCloudy))
	forecastRain := gf.AddTask(new(ForecastRain))
	forecastSun := gf.AddTask(new(ForecastSun))
	end := gf.AddTask(new(graphflow.EndTask))

	// add task paths to the graphflow
	gf.AddPath(start, graphflow.ALWAYS, isTheSkyCloudy)
	gf.AddPath(isTheSkyCloudy, graphflow.YES, forecastRain)
	gf.AddPath(isTheSkyCloudy, graphflow.NO, forecastSun)
	gf.AddPath(forecastRain, graphflow.ALWAYS, end)
	gf.AddPath(forecastSun, graphflow.ALWAYS, end)

	return gf
}

func TestGraphflowRenderGraph(t *testing.T) {
	gf := buildGraphflow()

	bytes, err := gf.RenderGraph()

	assert.NotEmpty(t, bytes.Bytes())
	assert.Nil(t, err)
}

func TestGraphflowRenderPathThroughGraph(t *testing.T) {
	ctx := new(graphflow.ExecutionContext)
	ctx.Set("Sky", "Clear")
	ctx.Set("Forecast", "")
	gf := buildGraphflow()

	bytes, err := gf.RenderPathThroughGraph(ctx, "Sky")

	assert.NotEmpty(t, bytes.Bytes())
	assert.Nil(t, err)

	assert.Equal(t, "Sun", ctx.Get("Forecast"))
}

func TestWorkflowWithNoPathsShouldRenderFine(t *testing.T) {
	var gf graphflow.Graphflow

	// create task instances
	start := new(graphflow.StartTask)
	forecastSun := new(ForecastSun)
	end := new(graphflow.EndTask)

	// add task instances to the graphflow
	gf.AddTask(start)
	gf.AddTask(forecastSun)
	gf.AddTask(end)

	_, err := gf.RenderGraph()

	assert.Nil(t, err)
}

func TestWorkflowWithNoPathsShouldRenderPathThroughGraphFine(t *testing.T) {
	var gf graphflow.Graphflow

	// create task instances
	start := new(graphflow.StartTask)
	forecastSun := new(ForecastSun)
	end := new(graphflow.EndTask)

	// add task instances to the graphflow
	gf.AddTask(start)
	gf.AddTask(forecastSun)
	gf.AddTask(end)

	_, err := gf.RenderPathThroughGraph(new(graphflow.ExecutionContext))

	assert.Nil(t, err)
}

func TestWorkflowWithTaskWithNoName(t *testing.T) {
	var gf graphflow.Graphflow

	// create task instances
	start := new(graphflow.StartTask)
	forecastNothing := new(TaskWithNoName)
	end := new(graphflow.EndTask)

	// add task instances to the graphflow
	gf.AddTask(start)
	gf.AddTask(forecastNothing)
	gf.AddTask(end)

	// add task paths to the graphflow
	gf.AddPath(start, graphflow.ALWAYS, forecastNothing)
	gf.AddPath(forecastNothing, graphflow.ALWAYS, end)

	_, err := gf.RenderPathThroughGraph(new(graphflow.ExecutionContext))

	assert.Nil(t, err)

	ctx := gf.GetContext()
	assert.Equal(t, "Nothing", ctx.Get("Forecast"))
}

func TestAddTaskGroups(t *testing.T) {
	var gf graphflow.Graphflow

	// create task instances
	start := new(graphflow.StartTask)
	forecastSun := new(ForecastSun)
	end := new(graphflow.EndTask)

	// add task instances to the graphflow
	gf.AddTask(start)
	gf.AddTask(forecastSun)
	gf.AddTask(end)

	taskGroup := gf.NewTaskGroup("my taskgroup")
	taskGroup.AddTasks(forecastSun)

	assert.Len(t, gf.TaskGroups(), 1)
	assert.Len(t, gf.TaskGroups()[0].Tasks(), 1)
	assert.Contains(t, gf.TaskGroups(), taskGroup)
	assert.Contains(t, gf.TaskGroups()[0].Tasks(), forecastSun)

	_, err := gf.RenderPathThroughGraph(new(graphflow.ExecutionContext))

	assert.Nil(t, err)
}

func TestTaskGroupsWithDuplicateTasks(t *testing.T) {
	var gf graphflow.Graphflow

	// create task instances
	start := new(graphflow.StartTask)
	forecastSun := new(ForecastSun)
	end := new(graphflow.EndTask)

	// add task instances to the graphflow
	gf.AddTask(start)
	gf.AddTask(forecastSun)
	gf.AddTask(end)

	taskGroup := gf.NewTaskGroup("my taskgroup")
	taskGroup.AddTasks(forecastSun)

	// don't allow tasks to appear in more than one TaskGroup
	duplicateTaskGroup := gf.NewTaskGroup("my other taskgroup")
	duplicateTaskGroup.AddTasks(forecastSun)

	_, err := gf.RenderPathThroughGraph(new(graphflow.ExecutionContext))

	assert.NotNil(t, err)
}
