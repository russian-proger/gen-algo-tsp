package main

import (
	"fmt"
	"net/http"
	"tsp/graph"
	"tsp/strategy"
	simplega "tsp/strategy/simple-genetic-algorithm"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

func main() {
	g := graph.RandomGraph()
	solver := simplega.Strategy{}

	solChannel := make(chan strategy.Output)
	go solver.Solve(strategy.Input{Graph: g, SolutionChannel: solChannel})

	solutions := make([]strategy.Output, 0)
	path := make([]graph.Edge, len(g.Vertices))
	gvisual := graph.Visualizer{Graph: g, Edges: path}
	go gvisual.Run()
	for solution := <-solChannel; !solution.Terminate; solution = <-solChannel {
		solutions = append(solutions, solution)
		for i := range solution.Order {
			path[i] = graph.Edge{
				From: g.Vertices[solution.Order[i]],
				To:   g.Vertices[solution.Order[(i+1)%len(path)]],
			}
		}
	}
	gvisual.Terminated = true

	report(solutions)

	fmt.Scanln()
}

func report(solutions []strategy.Output) {
	httpserver := func(w http.ResponseWriter, _ *http.Request) {
		var x []int = make([]int, 0)
		var y []opts.LineData = make([]opts.LineData, 0)
		for _, v := range solutions {
			x = append(x, v.GenerationId)
			y = append(y, opts.LineData{
				Value: v.TotalDistance,
			})
		}
		line := charts.NewLine()
		line.SetXAxis(x)
		line.AddSeries("Score", y)
		line.Render(w)
	}

	http.HandleFunc("/", httpserver)
	http.ListenAndServe(":8080", nil)
}
