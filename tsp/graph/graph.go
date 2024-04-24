package graph

import (
	"math"
	"math/rand"
	"tsp/constants"
)

type Graph struct {
	Vertices []*Vertice
	Matrix   [][]float64
}
type Position [2]float64

type Edge struct {
	From *Vertice
	To   *Vertice
	Cost float64
}

type Vertice struct {
	Neighbours []Edge
	Position   Position
}

func (v *Vertice) Distance(o *Vertice) float64 {
	return math.Hypot(v.Position[0]-o.Position[0], v.Position[1]-o.Position[1])
}

func (v *Vertice) addNeighbour(o *Vertice) {
	v.Neighbours = append(v.Neighbours, Edge{
		From: v,
		To:   o,
		Cost: v.Distance(o),
	})
}

func randomPosition() (p Position) {
	for i := range p {
		p[i] = rand.Float64() * constants.MAX_SIZE
	}
	return
}

// Return random graph
//
//	size = random of [MIN_COUNT, MAX_COUNT]
func RandomGraph() (g Graph) {
	n := rand.Int()%(constants.MAX_COUNT-constants.MIN_COUNT+1) + constants.MIN_COUNT
	g.Vertices = make([]*Vertice, n)
	for i := 0; i < n; i++ {
		g.Vertices[i] = new(Vertice)
		g.Vertices[i].Position = randomPosition()
		for j := 0; j < i; j++ {
			g.Vertices[i].addNeighbour(g.Vertices[j])
			g.Vertices[j].addNeighbour(g.Vertices[i])
		}
	}

	g.Matrix = make([][]float64, n)

	for ii, iv := range g.Vertices {
		g.Matrix[ii] = make([]float64, n)
		for jj, jv := range g.Vertices {
			g.Matrix[ii][jj] = iv.Distance(jv)
		}
	}
	return
}
