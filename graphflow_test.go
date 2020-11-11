package graphflow

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ForecastRain is a Task struct
type ForecastRain struct {
	Task
}

// String returns a description of the Task
func (t *ForecastRain) String() string {
	return "Forecast Rain"
}

// Execute sets the ExecutionContext's Forecast value to "Rain"
func (t *ForecastRain) Execute(ctx *ExecutionContext) error {
	ctx.Set("Forecast", "Rain")
	return nil
}

// ForecastSun is a Task struct
type ForecastSun struct {
	Task
}

// String returns a description of the Task
func (t *ForecastSun) String() string {
	return "Forecast Sun"
}

// Execute sets the ExecutionContext's Forecast value to "Sun"
func (t *ForecastSun) Execute(ctx *ExecutionContext) error {
	ctx.Set("Forecast", "Sun")
	return nil
}

// IsTheSkyCloudy is a Task struct
type IsTheSkyCloudy struct {
	Task
}

// String returns a description of the Task
func (t *IsTheSkyCloudy) String() string {
	return "Is the sky cloudy?"
}

// Execute retrieves the value of "Sky" from the ExecutionContext and checks whether it's Cloudy
func (t *IsTheSkyCloudy) Execute(ctx *ExecutionContext) error {
	sky := ctx.Get("Sky").(string)
	if sky == "Cloudy" {
		t.SetExitPath(YES)
	} else {
		t.SetExitPath(NO)
	}
	return nil
}

// TaskWithNoName is a Task struct
type TaskWithNoName struct {
	Task
}

// Execute sets the ExecutionContext's Forecast value to "Sun"
func (t *TaskWithNoName) Execute(ctx *ExecutionContext) error {
	ctx.Set("Forecast", "Nothing")
	return nil
}

func buildGraphflow() Graphflow {
	var gf Graphflow

	// create task instances
	start := new(StartTask)
	isTheSkyCloudy := new(IsTheSkyCloudy)
	forecastRain := new(ForecastRain)
	forecastSun := new(ForecastSun)
	end := new(EndTask)

	// add task instances to the graphflow
	gf.AddTask(start)
	gf.AddTask(isTheSkyCloudy)
	gf.AddTask(forecastRain)
	gf.AddTask(forecastSun)
	gf.AddTask(end)

	// add task paths to the graphflow
	gf.AddPath(start, ALWAYS, isTheSkyCloudy)
	gf.AddPath(isTheSkyCloudy, YES, forecastRain)
	gf.AddPath(isTheSkyCloudy, NO, forecastSun)
	gf.AddPath(forecastRain, ALWAYS, end)
	gf.AddPath(forecastSun, ALWAYS, end)

	return gf
}

func TestGraphflowCloudySky(t *testing.T) {
	ctx := new(ExecutionContext)
	ctx.Set("Sky", "Cloudy")
	ctx.Set("Forecast", "")

	gf := buildGraphflow()

	err := gf.Run(ctx)

	assert.Nil(t, err)
	assert.Equal(t, "Rain", ctx.Get("Forecast"))
	assert.Same(t, ctx, gf.GetContext())
}

func TestGraphflowClearSky(t *testing.T) {
	ctx := new(ExecutionContext)
	ctx.Set("Sky", "Clear")
	ctx.Set("Forecast", "")

	gf := buildGraphflow()

	err := gf.Run(ctx)

	assert.Nil(t, err)
	assert.Equal(t, "Sun", ctx.Get("Forecast"))
}

func TestGraphflowRenderGraph(t *testing.T) {
	gf := buildGraphflow()

	bytes, err := gf.RenderGraph()

	assert.NotEmpty(t, bytes.Bytes())
	assert.Nil(t, err)
}

func TestGraphflowRenderPathThroughGraph(t *testing.T) {
	ctx := new(ExecutionContext)
	ctx.Set("Sky", "Clear")
	ctx.Set("Forecast", "")
	gf := buildGraphflow()

	bytes, err := gf.RenderPathThroughGraph(ctx, "Sky")

	assert.NotEmpty(t, bytes.Bytes())
	assert.Nil(t, err)

	assert.Equal(t, "Sun", ctx.Get("Forecast"))
}

func TestGraphWithNoStartTask(t *testing.T) {
	var gf Graphflow

	// create task instances
	forecastSun := new(ForecastSun)
	end := new(EndTask)

	// add task instances to the graphflow
	gf.AddTask(forecastSun)
	gf.AddTask(end)

	// add task paths to the graphflow
	gf.AddPath(forecastSun, ALWAYS, end)

	err := gf.Run(new(ExecutionContext))

	assert.NotNil(t, err)
}

func TestGraphWithNoEndTask(t *testing.T) {
	var gf Graphflow

	// create task instances
	start := new(StartTask)
	forecastSun := new(ForecastSun)

	// add task instances to the graphflow
	gf.AddTask(start)
	gf.AddTask(forecastSun)

	// add task paths to the graphflow
	gf.AddPath(start, ALWAYS, forecastSun)

	err := gf.Run(new(ExecutionContext))

	assert.NotNil(t, err)
}

func TestALWAYSConditionWithYESThrowsError(t *testing.T) {
	var gf Graphflow

	// create task instances
	start := new(StartTask)
	forecastSun := new(ForecastSun)
	end := new(EndTask)

	// add task instances to the graphflow
	gf.AddTask(start)
	gf.AddTask(forecastSun)
	gf.AddTask(end)

	// add task paths to the graphflow
	gf.AddPath(start, ALWAYS, forecastSun)
	gf.AddPath(forecastSun, ALWAYS, end)
	gf.AddPath(forecastSun, YES, end) // this should cause Run() to error

	err := gf.Run(new(ExecutionContext))

	assert.NotNil(t, err)
}

func TestALWAYSConditionWithNOThrowsError(t *testing.T) {
	var gf Graphflow

	// create task instances
	start := new(StartTask)
	forecastSun := new(ForecastSun)
	end := new(EndTask)

	// add task instances to the graphflow
	gf.AddTask(start)
	gf.AddTask(forecastSun)
	gf.AddTask(end)

	// add task paths to the graphflow
	gf.AddPath(start, ALWAYS, forecastSun)
	gf.AddPath(forecastSun, ALWAYS, end)
	gf.AddPath(forecastSun, NO, end) // this should cause Run() to error

	err := gf.Run(new(ExecutionContext))

	assert.NotNil(t, err)
}

func TestYESConditionWithoutNOThrowsError(t *testing.T) {
	var gf Graphflow

	// create task instances
	start := new(StartTask)
	forecastSun := new(ForecastSun)
	end := new(EndTask)

	// add task instances to the graphflow
	gf.AddTask(start)
	gf.AddTask(forecastSun)
	gf.AddTask(end)

	// add task paths to the graphflow
	gf.AddPath(start, ALWAYS, forecastSun)
	gf.AddPath(forecastSun, YES, end)

	err := gf.Run(new(ExecutionContext))

	assert.NotNil(t, err)
}

func TestNOConditionWithoutYESThrowsError(t *testing.T) {
	var gf Graphflow

	// create task instances
	start := new(StartTask)
	forecastSun := new(ForecastSun)
	end := new(EndTask)

	// add task instances to the graphflow
	gf.AddTask(start)
	gf.AddTask(forecastSun)
	gf.AddTask(end)

	// add task paths to the graphflow
	gf.AddPath(start, ALWAYS, forecastSun)
	gf.AddPath(forecastSun, NO, end)

	err := gf.Run(new(ExecutionContext))

	assert.NotNil(t, err)
}

func TestWorkflowWithNoPathsShouldRunFine(t *testing.T) {
	var gf Graphflow

	// create task instances
	start := new(StartTask)
	forecastSun := new(ForecastSun)
	end := new(EndTask)

	// add task instances to the graphflow
	gf.AddTask(start)
	gf.AddTask(forecastSun)
	gf.AddTask(end)

	err := gf.Run(new(ExecutionContext))

	assert.Nil(t, err)
}

func TestWorkflowWithNoPathsShouldRenderFine(t *testing.T) {
	var gf Graphflow

	// create task instances
	start := new(StartTask)
	forecastSun := new(ForecastSun)
	end := new(EndTask)

	// add task instances to the graphflow
	gf.AddTask(start)
	gf.AddTask(forecastSun)
	gf.AddTask(end)

	_, err := gf.RenderGraph()

	assert.Nil(t, err)
}

func TestWorkflowWithNoPathsShouldRenderPathThroughGraphFine(t *testing.T) {
	var gf Graphflow

	// create task instances
	start := new(StartTask)
	forecastSun := new(ForecastSun)
	end := new(EndTask)

	// add task instances to the graphflow
	gf.AddTask(start)
	gf.AddTask(forecastSun)
	gf.AddTask(end)

	_, err := gf.RenderPathThroughGraph(new(ExecutionContext))

	assert.Nil(t, err)
}

func TestWorkflowWithTaskWithNoName(t *testing.T) {
	var gf Graphflow

	// create task instances
	start := new(StartTask)
	forecastNothing := new(TaskWithNoName)
	end := new(EndTask)

	// add task instances to the graphflow
	gf.AddTask(start)
	gf.AddTask(forecastNothing)
	gf.AddTask(end)

	// add task paths to the graphflow
	gf.AddPath(start, ALWAYS, forecastNothing)
	gf.AddPath(forecastNothing, ALWAYS, end)

	_, err := gf.RenderPathThroughGraph(new(ExecutionContext))

	assert.Nil(t, err)

	ctx := gf.GetContext()
	assert.Equal(t, "Nothing", ctx.Get("Forecast"))
}

func TestAddTaskGroups(t *testing.T) {
	var gf Graphflow

	// create task instances
	start := new(StartTask)
	forecastSun := new(ForecastSun)
	end := new(EndTask)

	// add task instances to the graphflow
	gf.AddTask(start)
	gf.AddTask(forecastSun)
	gf.AddTask(end)

	taskGroup := gf.NewTaskGroup("my taskgroup")
	taskGroup.AddTasks(forecastSun)

	assert.Len(t, gf.taskGroups, 1)
	assert.Len(t, gf.taskGroups[0].tasks, 1)
	assert.Contains(t, gf.taskGroups, taskGroup)
	assert.Contains(t, gf.taskGroups[0].tasks, forecastSun)

	_, err := gf.RenderPathThroughGraph(new(ExecutionContext))

	assert.Nil(t, err)
}

func TestTaskGroupsWithDuplicateTasks(t *testing.T) {
	var gf Graphflow

	// create task instances
	start := new(StartTask)
	forecastSun := new(ForecastSun)
	end := new(EndTask)

	// add task instances to the graphflow
	gf.AddTask(start)
	gf.AddTask(forecastSun)
	gf.AddTask(end)

	taskGroup := gf.NewTaskGroup("my taskgroup")
	taskGroup.AddTasks(forecastSun)

	// don't allow tasks to appear in more than one TaskGroup
	duplicateTaskGroup := gf.NewTaskGroup("my other taskgroup")
	duplicateTaskGroup.AddTasks(forecastSun)

	_, err := gf.RenderPathThroughGraph(new(ExecutionContext))

	assert.NotNil(t, err)
}
