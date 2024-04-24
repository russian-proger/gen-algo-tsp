package simplega

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"tsp/graph"
	"tsp/strategy"
)

// + Strategy
type Strategy struct {
	strategy.StrategyInterface
}

func (s *Strategy) Solve(input strategy.Input) strategy.Output {
	algorithm := Algorithm{graph: input.Graph}
	totalDistance, order := algorithm.Solve(input)
	return strategy.Output{
		TotalDistance: totalDistance,
		Order:         order,
	}
}

// - Strategy

// + Algorithm
type Algorithm struct {
	graph graph.Graph
}

func (algo *Algorithm) Solve(input strategy.Input) (totalDistance float64, order []int) {
	var graphSize int = len(algo.graph.Vertices)
	var population []*Target = generatePopulation(randomPopulationSize(graphSize), graphSize)
	var bufferPopulation []*Target = generatePopulation(len(population)*2, graphSize)
	var populationSize = len(population)

	const elitism float64 = 0.3
	const crookedTapeMeasuring float64 = 0.6
	const rangedSelection float64 = 0.1
	const mutationIntensityA = 0.05
	const mutationIntensityB = 0.005
	const generationsCount = 6000

	var lastBestScore float64 = 1e10
	for generationId := 1; generationId <= generationsCount; generationId++ {
		bufferPopulation = bufferPopulation[:0]

		// Calculating scores
		for i := range population {
			population[i].UpdateScore(algo.graph) // O(|P|*|G|)
		}

		// Sorting by score
		sort.Slice(population, func(i, j int) bool {
			return population[i].score < population[j].score // O(|P|*|G|)
		})

		if lastBestScore > 1/population[populationSize-1].score {
			fmt.Printf("Best Score (generation %v): %v\n", generationId, 1/population[populationSize-1].score)
			lastBestScore = 1 / population[populationSize-1].score
			input.SolutionChannel <- strategy.Output{
				TotalDistance: lastBestScore,
				Order:         population[populationSize-1].chromosome,
				GenerationId:  generationId,
				Terminate:     false,
			}
		}

		// Elitism - O(|P|)
		{
			for i := 0; float64(i)/elitism < float64(populationSize); i++ {
				bufferPopulation = append(bufferPopulation, population[populationSize-i-1])
			}
		}

		// Crooked Tape Measure - O(|P| * log(|P|) + |P| * |G|)
		{
			offset := len(bufferPopulation)
			prefixSum := make([]float64, populationSize)
			prefixSum[0] = population[0].score
			for i := 1; i < populationSize; i++ {
				prefixSum[i] = prefixSum[i-1] + population[i].score
			}
			count := 0
			for i := 0; float64(i) < float64(populationSize)*crookedTapeMeasuring; i++ {
				rnd := rand.Float64() * (prefixSum[populationSize-1])
				l, r := -1, populationSize-1
				for l < r-1 {
					x := (l + r) >> 1
					if rnd < prefixSum[x] {
						r = x
					} else {
						l = x
					}
				}
				bufferPopulation = append(bufferPopulation, population[r].Clone())
				count++
			}

			if count%2 != 0 {
				bufferPopulation = bufferPopulation[:len(bufferPopulation)-1]
				count--
			}

			for i := offset; i < offset+count; i += 2 {
				// fmt.Println(tempPopulation[i])
				bufferPopulation[i].Crossover(bufferPopulation[i+1])
			}

			for i := 0; i < count; i++ {
				bufferPopulation[offset+i].Mutate(mutationIntensityA, mutationIntensityB)
			}
		}

		// Ranged Selection
		{
			offset := len(bufferPopulation)
			prefixSum := make([]float64, populationSize)
			prefixSum[0] = 1
			for i := 1; i < populationSize; i++ {
				prefixSum[i] = prefixSum[i-1] + float64(i) + 1
			}
			count := 0
			for i := 0; float64(i) < float64(populationSize)*rangedSelection; i++ {
				rnd := rand.Float64() * (prefixSum[populationSize-1])
				l, r := -1, populationSize-1
				for l < r-1 {
					x := (l + r) >> 1
					if rnd < prefixSum[x] {
						r = x
					} else {
						l = x
					}
				}
				bufferPopulation = append(bufferPopulation, population[r].Clone())
				count++
			}

			if count%2 != 0 {
				bufferPopulation = bufferPopulation[:len(bufferPopulation)-1]
				count--
			}

			for i := offset; i < offset+count; i += 2 {
				// fmt.Println(tempPopulation[i])
				bufferPopulation[i].Crossover(bufferPopulation[i+1])
			}

			for i := 0; i < count; i++ {
				bufferPopulation[offset+i].Mutate(mutationIntensityA, mutationIntensityB)
			}
		}

		// Fix next population
		if len(bufferPopulation) > len(population) {
			bufferPopulation = bufferPopulation[:len(population)]
		}
		for len(bufferPopulation) < len(population) {
			bufferPopulation = append(bufferPopulation, generateTarget(graphSize))
		}

		// Update population
		bufferPopulation, population = population, bufferPopulation
	}

	var bestTarget *Target = nil
	for i := range population {
		if bestTarget == nil || population[i].Fitness(algo.graph) > bestTarget.Fitness(algo.graph) {
			bestTarget = population[i]
		}
	}
	order = bestTarget.chromosome

	input.SolutionChannel <- strategy.Output{
		GenerationId:  generationsCount,
		Order:         order,
		TotalDistance: 1 / bestTarget.Fitness(algo.graph),
		Terminate:     true,
	}

	fmt.Println("Genetic Algorithm was end!")
	fmt.Printf("Final Result: %v", 1/bestTarget.Fitness(algo.graph))

	return
}

func (algo *Algorithm) MakeSelection() []*Target {
	return make([]*Target, 0)
}

// - Algorithm

// + Target

type Target struct {
	chromosome        []int
	encodedChromosome []int
	score             float64
}

func (target *Target) UpdateChromosome(chromosome []int) {
	target.chromosome = chromosome
	target.encodedChromosome = EncodeChromosome(chromosome)
}

func (target *Target) UpdateEncodedChromosome(encodedChromosome []int) {
	target.encodedChromosome = encodedChromosome
	target.chromosome = DecodeChromosome(encodedChromosome)
}

func (target *Target) Fitness(g graph.Graph) float64 {
	path := target.chromosome
	result := 0.
	for i := 0; i < len(path); i++ {
		result += g.Matrix[path[i]][path[(i+1)%len(path)]]
	}
	return 1 / result
}

func (target *Target) Clone() (result *Target) {
	result = new(Target)
	encodedChromosome := make([]int, len(target.encodedChromosome))
	copy(encodedChromosome, target.encodedChromosome)
	result.UpdateEncodedChromosome(encodedChromosome)
	return
}

func (target *Target) UpdateScore(g graph.Graph) {
	target.score = target.Fitness(g)
}

func (target *Target) Mutate(intensityA float64, intensityB float64) {
	// Mutation A
	for i := range target.encodedChromosome {
		if rand.Float64() < intensityA {
			target.encodedChromosome[i] = rand.Intn(i + 1)
		}
	}
	target.UpdateEncodedChromosome((target.encodedChromosome))

	// Mutation B
	for i := range target.chromosome {
		if rand.Float64() < intensityB {
			j := rand.Intn(len(target.chromosome))
			target.chromosome[i], target.chromosome[j] = target.chromosome[j], target.chromosome[i]
		}
	}
	target.UpdateChromosome(target.chromosome)
}

func (target *Target) Crossover(other *Target) {
	n := max(rand.Intn(len(target.chromosome)), 3)
	changers := make([]int, n)

	// fmt.Printf("Crossover (was %v and %v) ", target.encodedChromosome, other.encodedChromosome)
	for i := range changers {
		changers[i] = rand.Intn(len(target.chromosome))
	}
	sort.Slice(changers, func(i, j int) bool {
		return changers[i] < changers[j]
	})
	state := rand.Intn(2)
	for i, j := 0, 0; i < len(target.chromosome); i++ {
		for j < n && changers[j] == i {
			j++
			state = 1 - state
		}
		if state == 1 {
			target.encodedChromosome[i], other.encodedChromosome[i] = other.encodedChromosome[i], target.encodedChromosome[i]
		}
	}
	target.UpdateEncodedChromosome(target.encodedChromosome)
	other.UpdateEncodedChromosome(other.encodedChromosome)
	// fmt.Printf(", (now %v and %v)\n", target.encodedChromosome, other.encodedChromosome)
}

// - Target

// + functions

// generate new target with n-sized subsequence
func generateTarget(n int) (target *Target) {
	target = new(Target)

	encodedChromosome := make([]int, n)
	for i := range encodedChromosome {
		encodedChromosome[i] = rand.Int() % (i + 1)
	}

	target.UpdateEncodedChromosome(encodedChromosome)

	return
}

// generate given sized new population
func generatePopulation(n int, m int) (population []*Target) {
	population = make([]*Target, n)

	for i := 0; i < n; i++ {
		population[i] = generateTarget(m)
	}

	return
}

func randomPopulationSize(graphSize int) int {
	var maxSize int = 1
	if graphSize > 15 {
		maxSize = math.MaxInt
	} else {
		for i := 2; i <= graphSize; i++ {
			maxSize *= i
		}
	}
	return min(2500, maxSize)
}

// encode chromosome in next form:
//
//	encodedChromosome[i] - count of chromosome[j] < chromosome[i] (j < i)
func EncodeChromosome(chromosome []int) (encodedChromosome []int) {
	n := len(chromosome)
	encodedChromosome = make([]int, n)

	for i, v := range chromosome {
		encodedChromosome[i] = 0
		for j := 0; j < i; j++ {
			if chromosome[j] < v {
				encodedChromosome[i]++
			}
		}
	}

	return
}

// decode chrosome to simple sequence
var decodeUsed []bool

func DecodeChromosome(encodedChromosome []int) (chromosome []int) {
	n := len(encodedChromosome)
	chromosome = make([]int, n)

	if cap(decodeUsed) < n {
		decodeUsed = make([]bool, n)
	} else {
		decodeUsed = decodeUsed[:n]
	}

	for i := range decodeUsed {
		decodeUsed[i] = false
	}

	for i := n - 1; i >= 0; i-- {
		t := 0
		for k := 0; k < encodedChromosome[i]; t++ {
			if !decodeUsed[t] {
				k++
			}
		}
		for decodeUsed[t] {
			t++
		}
		chromosome[i] = t
		decodeUsed[t] = true
	}

	return
}

// decode chrosome to simple sequence
var stree = segtree{}

func FastDecodeChromosome(encodedChromosome []int) (chromosome []int) {
	n := len(encodedChromosome)

	chromosome = make([]int, n)

	stree.reset(n)

	for i := n - 1; i >= 0; i-- {
		t := stree.getk(encodedChromosome[i])
		chromosome[i] = t
		stree.set(t, 0)
	}

	return
}

// - functions

func Test() {
	population := generatePopulation(10, 5)
	for _, v := range population {
		fmt.Printf("target: %v\n", v)
		fmt.Printf("decoded: %v\n", FastDecodeChromosome(v.encodedChromosome))
	}
}
