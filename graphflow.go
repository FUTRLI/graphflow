package graphflow

import (
	"bytes"
	"errors"
	"fmt"
	"log"

	"github.com/goccy/go-graphviz"
	"github.com/goccy/go-graphviz/cgraph"
)

type PathCondition int

const (
	ALWAYS PathCondition = 0
	YES    PathCondition = 1
	NO     PathCondition = 2
	ERROR  PathCondition = 3
)

var PathConditionName = map[PathCondition]string{
	0: "ALWAYS",
	1: "YES",
	2: "NO",
	3: "ERROR",
}

type Graphflow struct {
	context  *ExecutionContext
	tasks    []TaskIntf
	paths    map[TaskIntf]map[PathCondition]TaskIntf
	executed map[TaskIntf]bool
}

func (gf *Graphflow) GetContext() *ExecutionContext {
	return gf.context
}

func (gf *Graphflow) AddTask(task TaskIntf) {
	gf.tasks = append(gf.tasks, task)
}

func (gf *Graphflow) AddPath(from TaskIntf, condition PathCondition, to TaskIntf) {
	if gf.paths == nil {
		gf.paths = make(map[TaskIntf]map[PathCondition]TaskIntf)
	}
	if gf.paths[from] == nil {
		gf.paths[from] = make(map[PathCondition]TaskIntf)
	}
	gf.paths[from][condition] = to
}

func (gf *Graphflow) Run(context *ExecutionContext) error {
	gf.context = context
	err := gf.execute()
	if err != nil {
		return err
	}

	return nil
}

type ExecutionContext struct {
	values map[string]interface{}
}

func (ctx *ExecutionContext) Get(v string) interface{} {
	return ctx.values[v]
}

func (ctx *ExecutionContext) Set(key string, value interface{}) {
	if ctx.values == nil {
		ctx.values = make(map[string]interface{})
	}
	ctx.values[key] = value
}

type TaskIntf interface {
	SetExitPath(PathCondition)
	Execute(*ExecutionContext) error
	String() string
	getExitPath() PathCondition
}

type Task struct {
	Name     string
	exitPath PathCondition
}

func (t *Task) Execute(c *ExecutionContext) error {
	return nil
}

func (t *Task) SetExitPath(path PathCondition) {
	t.exitPath = path
}

func (t *Task) getExitPath() PathCondition {
	return t.exitPath
}

func (t *Task) String() string {
	return "Unnamed TaskIntf"
}

type StartTask struct {
	Task
}

func (t *StartTask) String() string {
	return "Start"
}

type EndTask struct {
	Task
}

func (t *EndTask) String() string {
	return "End"
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

func (q *taskQueue) size() int {
	return len(q.tasks)
}

func (gf *Graphflow) execute() error {
	q := taskQueue{}
	q.new()
	t, err := gf.findStartTask()
	if err != nil {
		return err
	}
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

func (wf *Graphflow) findStartTask() (TaskIntf, error) {
	for _, task := range wf.tasks {
		_, isStartTask := task.(*StartTask)
		if isStartTask {
			return task, nil
		}
	}
	return nil, errors.New("Workflow needs to contain a task of type StartTask")
}

func (gf *Graphflow) RenderGraph() (bytes.Buffer, error) {
	return gf.generateGraph(false)
}

func (gf *Graphflow) RenderPathThroughGraph() (bytes.Buffer, error) {
	return gf.generateGraph(true)
}

func (gf *Graphflow) generateGraph(showPath bool) (bytes.Buffer, error) {
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
		for k, v := range gf.context.values {
			str, ok := v.(string)
			if ok {
				desc = fmt.Sprintf("%s\n%s = %s", desc, k, str)
			}
		}
		if desc != "" {
			desc = fmt.Sprintf("This is the path taken when:\n%s", desc)
			n, err := graph.CreateNode(desc)
			if err != nil {
				return buf, err
			}
			n.SetShape(cgraph.NoteShape)
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
