package graph

import (
	"image/color"
	"tsp/constants"

	"github.com/go-p5/p5"
)

type VisualizerInterface interface {
	Draw()
	Setup()
	Run()
}

type Visualizer struct {
	VisualizerInterface
	Graph      Graph
	Edges      []Edge
	Terminated bool
}

func (gv *Visualizer) Draw() {
	for _, v := range gv.Graph.Vertices {
		p5.Circle(v.Position[0], v.Position[1], 3)
	}
	for _, e := range gv.Edges {
		a := e.From.Position
		b := e.To.Position
		var c color.Color = color.Black
		var w float64 = 2
		if gv.Terminated {
			c = color.RGBA{R: 50, G: 200, B: 50, A: 255}
			w = 4
		}
		p5.Stroke(c)
		p5.StrokeWidth(w)
		p5.Line(a[0], a[1], b[0], b[1])
	}
}

func (gv *Visualizer) Setup() {
	p5.Canvas(constants.MAX_SIZE, constants.MAX_SIZE)
}

func (gv *Visualizer) Run() {
	p5.Run(gv.Setup, gv.Draw)
}
