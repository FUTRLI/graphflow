// Package graphflow supports the building, execution and graphical rendering of simple linear workflows
//
// There may be times when complex business logic is best represented in the form of a decision tree or a workflow
// comprised of some "context" and a series of yes/no questions and actions or decisions that can be made against this
// context.
//
// A naive representation of this in code could be a error-prone block of deeply nested if / else conditions. In such a
// case it is difficult for a product stakeholder to precisely express what they expect and it's equally difficult for them
// to validate it once it's built.
//
// The graphflow package allows this logic to instead be represented as a self-documenting series of Task nodes with
// conditional Paths between them. It is self-documenting because it includes the ability to output a graphviz png
// representation of the entire graph as well as the particular path taken through the graph when provided with a given
// context.
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
	context  *ExecutionContext
	tasks    []TaskIntf
	paths    map[TaskIntf]map[PathCondition]TaskIntf
	executed map[TaskIntf]bool
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
func (gf *Graphflow) AddTask(task TaskIntf) {
	gf.tasks = append(gf.tasks, task)
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
	getExitPath() PathCondition
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

func (t *Task) getExitPath() PathCondition {
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
		gf.Run(context)
	}
	return gf.generateGraph(true, contextKeysToRender...)
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
			if path != task.getExitPath() {
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
	graph, err := g.Graph()
	if err != nil {
		return buf, err
	}
	defer func() {
		if err := graph.Close(); err != nil {
			log.Fatal(err)
		}
		g.Close()
	}()
	nodes := make(map[string]*cgraph.Node)
	for _, t := range gf.tasks {
		n, err := graph.CreateNode(t.String())
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
		nodes[fmt.Sprintf("%T", t)] = n
	}
	if showPath {
		desc := ""
		for _, k := range contextKeysToRender {
			desc = fmt.Sprintf("%s\n%s = %v", desc, k, gf.context.Get(k))
		}
		if desc != "" {
			desc = fmt.Sprintf("This is the path taken when:\n%s", desc)
			n, err := graph.CreateNode(desc)
			if err != nil {
				return buf, err
			}
			n.SetShape(cgraph.UnderlineShape)
			n.SetMargin(0.2)
		}
	}
	for from, edge := range gf.paths {
		n1 := nodes[fmt.Sprintf("%T", from)]
		for label, to := range edge {
			n2 := nodes[fmt.Sprintf("%T", to)]
			e, err := graph.CreateEdge("to", n1, n2)
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
	if err := g.Render(graph, graphviz.PNG, &buf); err != nil {
		return buf, err
	}
	return buf, nil
}
