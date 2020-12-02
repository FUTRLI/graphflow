// Package graphflow supports the building, execution and graphical rendering of simple linear workflows.
//
// There may be times when complex business logic is best represented in the form of a decision tree or a workflow
// made up of a series of yes/no questions and actions or decisions that should be taken.
//
// The graphflow package allows this logic to be built and represented as a self-documenting series of Task nodes with
// conditional Paths between them, allowing a project stakeholder to easily validate what has been built.
//
package graphflow

import (
	"bytes"
	"errors"
	"fmt"
	"log"

	"github.com/goccy/go-graphviz"
	"github.com/goccy/go-graphviz/cgraph"
)

// PathCondition is a type representing the condition that should be satisfied for a certain path
// to be followed when traversing the graph
type PathCondition int

const (
	// ALWAYS is the default PathCondition (as 0 is the default value of an int). A path with the ALWAYS PathCondition
	// set will always be followed. No other PathCondition can be added to a node if it has an ALWAYS path leaving it.
	ALWAYS PathCondition = 0
	// YES is the PathCondition that should be set by a question node if its condition passes
	YES PathCondition = 1
	// NO is the PathCondition that should be set by a question node if its condition fails
	NO PathCondition = 2
	// ERROR is the PathCondition that should be set by any node if you want a particular path to be followed in the event of an error
	ERROR PathCondition = 3
)

// PathConditionName is a map back to the textual name for each PathCondition
var PathConditionName = map[PathCondition]string{
	0: "ALWAYS",
	1: "YES",
	2: "NO",
	3: "ERROR",
}

// Graphflow represents a series of Tasks with defined Paths between them constructed as a simple workflow.
// Graphflow methods support its construction, execution and rendering as a graphflow png.
type Graphflow struct {
	context    *ExecutionContext
	tasks      []TaskIntf
	taskGroups []*TaskGroup
	paths      map[TaskIntf]map[PathCondition]TaskIntf
	executed   map[TaskIntf]bool
}

// ExecutionContext is a map of values of any type that is passed from Task to Task as the graphflow is executed
// Tasks are free to read from the context as well as write to it.
// During the graphflow execution one or more of the Tasks should store the result of the run in the ExecutionContext
// for retrieval after execution has completed.
type ExecutionContext struct {
	values map[string]interface{}
}

// GetContext retrieves the current ExecutionContext from the graphflow
func (gf *Graphflow) GetContext() *ExecutionContext {
	return gf.context
}

// AddTask adds a new Task (a struct implementing the TaskIntf interface) to the graphflow
func (gf *Graphflow) AddTask(task TaskIntf) TaskIntf {
	gf.tasks = append(gf.tasks, task)
	return task
}

// AddPath adds conditional Paths between graphflow Tasks
func (gf *Graphflow) AddPath(from TaskIntf, condition PathCondition, to TaskIntf) {
	if gf.paths == nil {
		gf.paths = make(map[TaskIntf]map[PathCondition]TaskIntf)
	}
	if gf.paths[from] == nil {
		gf.paths[from] = make(map[PathCondition]TaskIntf)
	}
	gf.paths[from][condition] = to
}

// Run passes the graphflow an ExecutionContext and, starting at the StartTask, follows conditional Paths
// through the graphflow, executing each Task until it reaches the EndTask.
func (gf *Graphflow) Run(context *ExecutionContext) error {
	gf.context = context
	gf.executed = make(map[TaskIntf]bool)
	err := gf.execute()
	if err != nil {
		return err
	}

	return nil
}

// Get retrieves a specific value from the ExecutionContext
func (ctx *ExecutionContext) Get(v string) interface{} {
	return ctx.values[v]
}

// Set sets a specific value in the ExecutionContext, updating it if if already exists
func (ctx *ExecutionContext) Set(key string, value interface{}) {
	if ctx.values == nil {
		ctx.values = make(map[string]interface{})
	}
	ctx.values[key] = value
}

// TaskIntf is the interface that all graphflow Tasks need to implement.
// New Tasks should be defined like so:
//
// Example:
//   type MyNewTask struct {
//	   graphflow.Task
//	 }
//
// This ensures that they include the implementation of SetExitPath(PathCondition)
// provided by graphflow.Task
type TaskIntf interface {
	SetExitPath(PathCondition)
	Execute(*ExecutionContext) error
	String() string
	ExitPath() PathCondition
}

// Task is a struct that all new Tasks should include in their definition.
//
// Example:
//   type MyNewTask struct {
//	   graphflow.Task
//	 }
type Task struct {
	exitPath PathCondition
}

// Execute is the empty default method on Task that needs to be overridden by
// Tasks you create
func (t *Task) Execute(c *ExecutionContext) error {
	return nil
}

// SetExitPath needs to be called by Tasks you create in their Execute() method if you want the ExitPath
// to be anything other than the default PathCondition, ALWAYS
func (t *Task) SetExitPath(path PathCondition) {
	t.exitPath = path
}

// ExitPath returns the current ExitPath for this task
func (t *Task) ExitPath() PathCondition {
	return t.exitPath
}

// String is the default implementation of the String() method that needs to be overridden by Tasks you create.
// It should return a meaningful description of your Task that'll be output in the graphviz png
func (t *Task) String() string {
	return "Unnamed Task"
}

// StartTask is a Task provided by the package. Every graphflow must include a StartTask.
type StartTask struct {
	Task
}

// String returns the name of the StartTask
func (t *StartTask) String() string {
	return "Start"
}

// EndTask is a Task provided by the package. Every graphflow must include an EndTask.
type EndTask struct {
	Task
}

// String returns the name of the EndTask
func (t *EndTask) String() string {
	return "End"
}

// RenderGraph returns a buffer of bytes containing a graphviz png representation of all the Tasks and the Paths
// connecting them
func (gf *Graphflow) RenderGraph() (bytes.Buffer, error) {
	return gf.generateGraph(false, "")
}

// RenderPathThroughGraph returns a buffer of bytes containing a graphviz png representation of all the Tasks and the Paths
// connecting them, with the actual path taken for the given ExecutionContext highlighted. Any Context Keys provided will be
// rendered with their values at the top of the image.
func (gf *Graphflow) RenderPathThroughGraph(context *ExecutionContext, contextKeysToRender ...string) (bytes.Buffer, error) {
	if len(gf.executed) == 0 {
		err := gf.Run(context)
		if err != nil {
			var buf bytes.Buffer
			return buf, err
		}
	}
	return gf.generateGraph(true, contextKeysToRender...)
}

// TaskGroup can have Tasks added to it, meaning they'll be rendered together with a box around them and a label set to the
// TaskGroup's name
type TaskGroup struct {
	name  string
	tasks []TaskIntf
}

// AddTasks allows Tasks to be added to a TaskGroup
func (t *TaskGroup) AddTasks(tasks ...TaskIntf) {
	t.tasks = append(t.tasks, tasks...)
}

// NewTaskGroup creates a new TaskGroup with the provided name and adds it to the graphflow. Add Tasks to the *TaskGroup it returns
// for them to be rendered together with a box around them and a label set to the TaskGroup's name
func (gf *Graphflow) NewTaskGroup(name string) *TaskGroup {
	taskGroup := &TaskGroup{
		name: name,
	}
	gf.taskGroups = append(gf.taskGroups, taskGroup)
	return taskGroup
}

// BFS (Breadth-First Search) is one of the most widely known algorithms to traverse a graph.
// Starting from a node, it first traverses all its directly linked tasks, then processes the
// tasks linked to those, and so on.
//
// Here we'll implement it using a queue

type taskQueue struct {
	tasks []TaskIntf
}

func (q *taskQueue) new() *taskQueue {
	q.tasks = []TaskIntf{}
	return q
}

func (q *taskQueue) enqueue(t TaskIntf) {
	q.tasks = append(q.tasks, t)
}

func (q *taskQueue) dequeue() TaskIntf {
	task := q.tasks[0]
	q.tasks = q.tasks[1:len(q.tasks)]
	return task
}

func (q *taskQueue) isEmpty() bool {
	return len(q.tasks) == 0
}

func (gf *Graphflow) execute() error {
	err := gf.validateTasks()
	if err != nil {
		return err
	}
	t, err := gf.findStartTask()
	if err != nil {
		return err
	}
	q := taskQueue{}
	q.new()
	q.enqueue(t)
	visited := make(map[TaskIntf]bool)
	for {
		if q.isEmpty() {
			break
		}
		task := q.dequeue()
		visited[task] = true
		err := task.Execute(gf.context)
		if err != nil {
			return err
		}

		near := gf.paths[task]

		for path, to := range near {
			if path != task.ExitPath() {
				continue
			}
			if !visited[to] {
				q.enqueue(to)
				visited[to] = true
			}
		}
	}
	gf.executed = visited
	return nil
}

func (gf *Graphflow) findStartTask() (TaskIntf, error) {
	for _, task := range gf.tasks {
		_, isStartTask := task.(*StartTask)
		if isStartTask {
			return task, nil
		}
	}
	return nil, errors.New("Workflow needs to contain a task of type StartTask")
}

func (gf *Graphflow) findEndTask() (TaskIntf, error) {
	for _, task := range gf.tasks {
		_, isEndTask := task.(*EndTask)
		if isEndTask {
			return task, nil
		}
	}
	return nil, errors.New("Workflow needs to contain a task of type EndTask")
}

func (gf *Graphflow) validateTasks() error {
	_, err := gf.findEndTask()
	if err != nil {
		return err
	}
	for task := range gf.paths {
		conditions := []PathCondition{}
		for condition := range gf.paths[task] {
			conditions = append(conditions, condition)
		}
		if contains(conditions, ALWAYS) {
			if contains(conditions, YES) {
				return fmt.Errorf("Task %s cannot have an ALWAYS path as well as a YES path", task.String())
			} else if contains(conditions, NO) {
				return fmt.Errorf("Task %s cannot have an ALWAYS path as well as a NO path", task.String())
			}
		} else if contains(conditions, YES) {
			if !contains(conditions, NO) {
				return fmt.Errorf("Task %s has as a YES path but no NO path", task.String())
			}
		} else if contains(conditions, NO) {
			if !contains(conditions, YES) {
				return fmt.Errorf("Task %s has as a NO path but no YES path", task.String())
			}
		}
	}
	for _, taskGroup := range gf.taskGroups {
		for _, t := range taskGroup.tasks {
			for _, otherTaskGroup := range gf.taskGroups {
				if taskGroup == otherTaskGroup {
					continue
				}
				for _, otherTask := range otherTaskGroup.tasks {
					if t == otherTask {
						return fmt.Errorf("A Task can only exist in one TaskGroup but \"%s\" exists in \"%s\" and \"%s\"", t, taskGroup.name, otherTaskGroup.name)
					}
				}
			}
		}
	}
	return nil
}

func contains(conditions []PathCondition, condition PathCondition) bool {
	for _, c := range conditions {
		if c == condition {
			return true
		}
	}
	return false
}

func (gf *Graphflow) generateGraph(showPath bool, contextKeysToRender ...string) (bytes.Buffer, error) {
	var buf bytes.Buffer
	g := graphviz.New()
	parentGraph, err := g.Graph()
	if err != nil {
		return buf, err
	}
	defer func() {
		if err := parentGraph.Close(); err != nil {
			log.Fatal(err)
		}
		g.Close()
	}()
	graphs := make(map[TaskIntf]*cgraph.Graph)
	// first of all link each task to the parent graph
	for _, t := range gf.tasks {
		graphs[t] = parentGraph
	}
	// for each task group, create a sub-graph
	for _, tg := range gf.taskGroups {
		graph := parentGraph.SubGraph(fmt.Sprintf("cluster_%s", tg.name), 1)
		graph.SetLabel(tg.name)
		graph.SetLabelJust("l")
		graph.SetStyle("filled")
		graph.SetBackgroundColor("lightgrey")
		for _, t := range tg.tasks {
			// link tasks to the subgraph instead
			graphs[t] = graph
		}
	}

	nodes := make(map[TaskIntf]*cgraph.Node)
	for _, t := range gf.tasks {
		n, err := graphs[t].CreateNode(fmt.Sprintf("%p", t))
		n.SetLabel(t.String())
		if err != nil {
			return buf, err
		}
		n.SetStyle("filled")
		if showPath {
			n.SetColorScheme("greys3")
			n.SetColor("1")
			n.SetFontColor("2")
		} else {
			n.SetColorScheme("paired10")
			n.SetColor("6") // red
		}
		nodes[t] = n
	}
	if showPath {
		desc := ""
		for _, k := range contextKeysToRender {
			desc = fmt.Sprintf("%s\n%s = %v", desc, k, gf.context.Get(k))
		}
		if desc != "" {
			desc = fmt.Sprintf("This is the path taken when:\n%s", desc)
			n, err := parentGraph.CreateNode(desc)
			if err != nil {
				return buf, err
			}
			n.SetShape(cgraph.UnderlineShape)
			n.SetMargin(0.2)
		}
	}
	for from, edge := range gf.paths {
		n1 := nodes[from]
		for label, to := range edge {
			n2 := nodes[to]
			e, err := parentGraph.CreateEdge("to", n1, n2)
			if err != nil {
				return buf, err
			}
			if label != ALWAYS {
				e.SetLabel(PathConditionName[label])
			}
			_, isEndTask := to.(*EndTask)
			if isEndTask {
				n2.SetColorScheme("paired10")
				n2.SetColor("7") // orange
				n2.SetFontColor("")
			}
			if showPath {
				if !gf.executed[to] {
					n2.SetColorScheme("greys3")
					n2.SetColor("1") // grey
					n2.SetFontColor("2")
				}
			}
		}
		n1.SetColorScheme("paired10")
		n1.SetFontColor("")
		_, isStartTask := from.(*StartTask)
		if isStartTask {
			n1.SetColor("7") // orange
		} else {
			n1.SetColor("9") // mauve
			for label := range edge {
				if label == YES || label == NO {
					n1.SetColor("3") // green
					break
				}
			}
		}
		if showPath {
			if !gf.executed[from] {
				n1.SetColorScheme("greys3")
				n1.SetColor("1") // grey
				n1.SetFontColor("2")
			}
		}
	}
	if err := g.Render(parentGraph, graphviz.PNG, &buf); err != nil {
		return buf, err
	}
	return buf, nil
}
