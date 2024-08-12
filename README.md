# graphflow [![Go](https://github.com/futrli/graphflow/workflows/CI/badge.svg)](https://github.com/futrli/graphflow/actions) 

graphflow supports the building, execution and graphical rendering of simple linear workflows.

There may be times when complex business logic is best represented in the form of a decision tree or a workflow
made up of a series of yes/no questions and actions or decisions that should be taken.

The graphflow package allows this logic to be built and represented as a self-documenting series of Task nodes with
conditional Paths between them, allowing a project stakeholder to easily validate what has been built.

# Features

- A sensible API to construct a graphflow from Tasks and Paths
- Shared ExecutionContext passed between Tasks in which any data can be stored
- Rendering of the graphflow structure
- Rendering of the path taken through a graphflow, given a particular context

# Installation

```bash
$ go get github.com/futrli/graphflow
```

# Synopsis

### See the _examples directory for the Tasks used here:

```go
package main

import (
	"fmt"
	"io/ioutil"
	"log"

	"github.com/futrli/graphflow/_examples/tasks"

	"github.com/futrli/graphflow"
	"github.com/futrli/graphflow/rendering"
)

// my graphflow:

func buildGraphflow() *graphflow.Graphflow {
	gf := new(graphflow.Graphflow)

	// create task instances and add to graphflow
	start := gf.AddTask(new(graphflow.StartTask))
	isTheSkyCloudy := gf.AddTask(new(tasks.IsTheSkyCloudy))
	forecastFog := gf.AddTask(new(tasks.ForecastFog)) // leave this deliberately orphaned with no Paths in or out
	forecastRain := gf.AddTask(new(tasks.ForecastRain))
	forecastSun := gf.AddTask(new(tasks.ForecastSun))
	end := gf.AddTask(new(graphflow.EndTask))

	// add task paths to the graphflow
	gf.AddPath(start, graphflow.ALWAYS, isTheSkyCloudy)
	gf.AddPath(isTheSkyCloudy, graphflow.YES, forecastRain)
	gf.AddPath(isTheSkyCloudy, graphflow.NO, forecastSun)
	gf.AddPath(forecastRain, graphflow.ALWAYS, end)
	gf.AddPath(forecastSun, graphflow.ALWAYS, end)

	// create a task group
	forecastingGroup := gf.NewTaskGroup("Forecasting")

	// add forecasting tasks to this group
	forecastingGroup.AddTasks(forecastFog, forecastRain, forecastSun)

	return gf
}

func main() {

	ctx := new(graphflow.ExecutionContext)
	//ctx.Set("Sky", "Clear")
	ctx.Set("Sky", "Cloudy")
	ctx.Set("Forecast", "")

	gf := buildGraphflow()

	err := gf.Run(ctx)
	if err != nil {
		log.Fatalf("Error: %s\n", err)
	}

	fmt.Printf("Successfully forecasted: %v\n", ctx.Get("Forecast"))

	buf, _ := rendering.RenderGraph(gf)
	ioutil.WriteFile("./graph.png", buf.Bytes(), 0664)

	buf, _ = rendering.RenderPathThroughGraph(ctx, gf, "Sky")
	ioutil.WriteFile("./pathThroughGraph.png", buf.Bytes(), 0664)

}
```

### The above would output this graph.png, giving an overview of the graphflow:

<img src="https://github.com/FUTRLI/graphflow/raw/master/_examples/graph.png"></img>

### and this pathThroughGraph.png, showing the path taken through the graph:

<img src="https://github.com/FUTRLI/graphflow/raw/master/_examples/pathThroughGraph.png"></img>

