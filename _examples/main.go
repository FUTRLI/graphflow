package main

import (
	"fmt"
	"io/ioutil"

	"github.com/futrli/graphflow/_examples/tasks"

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
