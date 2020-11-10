# graphflow [![Go](https://github.com/futrli/graphflow/workflows/CI/badge.svg)](https://github.com/futrli/graphflow/actions) 

graphflow supports the building, execution and graphical rendering of simple linear workflows.

There may be times when complex business logic is best represented in the form of a decision tree or a workflow
made up of a series of yes/no questions and actions or decisions that can be taken.

In such a case it is difficult for a product stakeholder to precisely express what they expect and it's equally difficult for them to validate it once it's built.

The graphflow package allows this logic to be built and represented as a self-documenting series of Task nodes with
conditional Paths between them.

# Features

- A sensible API to construct a graphflow from Tasks and Paths
- Rendering of the graphflow structure
- Rendering of the path taken through a graphflow, given a particular context

# Installation

```bash
$ go get github.com/futrli/graphflow
```

# Synopsis

## See the _examples directory for the Tasks used here:

```go
package main

import (
	"fmt"
	"io/ioutil"

	"github.com/futrli/graphflow/examples/tasks"

	"github.com/futrli/graphflow"
)

// my graphflow:

func buildGraphflow() graphflow.Graphflow {
	var gf graphflow.Graphflow

	// create task instances
	start := new(graphflow.StartTask)
	isTheSkyCloudy := new(tasks.IsTheSkyCloudy)
	forecastFog := new(tasks.ForecastFog)
	forecastRain := new(tasks.ForecastRain)
	forecastSun := new(tasks.ForecastSun)
	end := new(graphflow.EndTask)

	// add task instances to the graphflow
	gf.AddTask(start)
	gf.AddTask(isTheSkyCloudy)
	gf.AddTask(forecastFog) // leave this deliberately orphaned with no Paths in or out
	gf.AddTask(forecastRain)
	gf.AddTask(forecastSun)
	gf.AddTask(end)

	// add task paths to the graphflow
	gf.AddPath(start, graphflow.ALWAYS, isTheSkyCloudy)
	gf.AddPath(isTheSkyCloudy, graphflow.YES, forecastRain)
	gf.AddPath(isTheSkyCloudy, graphflow.NO, forecastSun)
	gf.AddPath(forecastRain, graphflow.ALWAYS, end)
	gf.AddPath(forecastSun, graphflow.ALWAYS, end)

	return gf
}

func main() {

	ctx := new(graphflow.ExecutionContext)
	//ctx.Set("Sky", "Clear")
	ctx.Set("Sky", "Cloudy")
	ctx.Set("Forecast", "")

	gf := buildGraphflow()

	gf.Run(ctx)

	fmt.Printf("Successfully forecasted: %v\n", ctx.Get("Forecast"))

	buf, _ := gf.RenderGraph()
	ioutil.WriteFile("./graph.png", buf.Bytes(), 0664)

	buf, _ = gf.RenderPathThroughGraph(ctx, "Sky")
	ioutil.WriteFile("./pathThroughGraph.png", buf.Bytes(), 0664)

}
```

The above would output this graph.png, giving an overview of the graphflow:

<img src="https://github.com/FUTRLI/graphflow/raw/master/_examples/graph.png"></img>

and this pathThroughGraph.png, showing the path taken through the graph:

<img src="https://github.com/FUTRLI/graphflow/raw/master/_examples/pathThroughGraph.png"></img>

