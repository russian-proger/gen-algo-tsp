package strategy

import "tsp/graph"

type StrategyInterface interface {
	Solve(Input) Output
}

type Input struct {
	Graph           graph.Graph
	SolutionChannel chan Output
}

type Output struct {
	TotalDistance float64
	Order         []int
	GenerationId  int
	Terminate     bool
}

type Strategy struct {
	StrategyInterface
	tspStrategy StrategyInterface
}

func (tsp *Strategy) Solve(input Input) Output {
	return tsp.tspStrategy.Solve(input)
}
