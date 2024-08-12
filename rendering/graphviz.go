package rendering

import (
	"bytes"
	"fmt"
	"github.com/futrli/graphflow"
	"github.com/goccy/go-graphviz"
	"github.com/goccy/go-graphviz/cgraph"
	"log"
)

// RenderGraph returns a buffer of bytes containing a graphviz png representation of all the Tasks and the Paths
// connecting them. Needs to be passed a GraphVizIntf implementation to render the graph (eg graphviz.New())
func RenderGraph(gf *graphflow.Graphflow) (bytes.Buffer, error) {
	return generateGraph(gf, false, "")
}

// RenderPathThroughGraph returns a buffer of bytes containing a graphviz png representation of all the Tasks and the Paths
// connecting them, with the actual path taken for the given ExecutionContext highlighted. Any Context Keys provided will be
// rendered with their values at the top of the image.
func RenderPathThroughGraph(context *graphflow.ExecutionContext, gf *graphflow.Graphflow, contextKeysToRender ...string) (bytes.Buffer, error) {
	if len(gf.Executed()) == 0 {
		err := gf.Run(context)
		if err != nil {
			var buf bytes.Buffer
			return buf, err
		}
	}
	return generateGraph(gf, true, contextKeysToRender...)
}

func generateGraph(gf *graphflow.Graphflow, showPath bool, contextKeysToRender ...string) (bytes.Buffer, error) {
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
	graphs := make(map[graphflow.TaskIntf]*cgraph.Graph)
	// first of all link each task to the parent graph
	for _, t := range gf.Tasks() {
		graphs[t] = parentGraph
	}
	// for each task group, create a sub-graph
	for _, tg := range gf.TaskGroups() {
		graph := parentGraph.SubGraph(fmt.Sprintf("cluster_%s", tg.Name()), 1)
		graph.SetLabel(tg.Name())
		graph.SetLabelJust("l")
		graph.SetStyle("filled")
		graph.SetBackgroundColor("lightgrey")
		for _, t := range tg.Tasks() {
			// link tasks to the subgraph instead
			graphs[t] = graph
		}
	}

	nodes := make(map[graphflow.TaskIntf]*cgraph.Node)
	for _, t := range gf.Tasks() {
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
			desc = fmt.Sprintf("%s\n%s = %v", desc, k, gf.GetContext().Get(k))
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
	for from, edge := range gf.Paths() {
		n1 := nodes[from]
		for label, to := range edge {
			n2 := nodes[to]
			e, err := parentGraph.CreateEdge("to", n1, n2)
			if err != nil {
				return buf, err
			}
			if label != graphflow.ALWAYS {
				e.SetLabel(graphflow.PathConditionName[label])
			}
			_, isEndTask := to.(*graphflow.EndTask)
			if isEndTask {
				n2.SetColorScheme("paired10")
				n2.SetColor("7") // orange
				n2.SetFontColor("")
			}
			if showPath {
				if !gf.Executed()[to] {
					n2.SetColorScheme("greys3")
					n2.SetColor("1") // grey
					n2.SetFontColor("2")
				}
			}
		}
		n1.SetColorScheme("paired10")
		n1.SetFontColor("")
		_, isStartTask := from.(*graphflow.StartTask)
		if isStartTask {
			n1.SetColor("7") // orange
		} else {
			n1.SetColor("9") // mauve
			for label := range edge {
				if label == graphflow.YES || label == graphflow.NO {
					n1.SetColor("3") // green
					break
				}
			}
		}
		if showPath {
			if !gf.Executed()[from] {
				n1.SetColorScheme("greys3")
				n1.SetColor("1") // grey
				n1.SetFontColor("2")
			}
		}
	}
	if err := g.Render(parentGraph, "png", &buf); err != nil {
		return buf, err
	}
	return buf, nil
}
