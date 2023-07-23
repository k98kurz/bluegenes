package bluegenes

import (
	"math"
)

type Neuron struct {
	Weights            []float64
	Bias               float64
	ActivationFunction func(float64) float64
	Value              float64
	activationCache    float64
}

func (n *Neuron) Activate(inputs []float64) float64 {
	min_size, _ := min(len(inputs), len(n.Weights))
	vals := make([]float64, min_size)
	for i := 0; i < min_size; i++ {
		vals[i] = inputs[i] * n.Weights[i]
	}
	total := reduce(vals, func(f1, f2 float64) float64 { return f1 + f2 })

	n.activationCache = total + n.Bias
	n.Value = n.ActivationFunction(n.activationCache)
	return n.Value
}

func NewNeuron(weights []float64, bias float64,
	activationFunc ...func(float64) float64) Neuron {
	actFunc := math.Tanh
	if len(activationFunc) > 0 {
		actFunc = activationFunc[0]
	}
	return Neuron{Weights: weights, Bias: bias, ActivationFunction: actFunc}
}

type Layer struct {
	Neurons []Neuron
}

func (l *Layer) FeedForward(inputs []float64) []float64 {
	vals := make([]float64, len(l.Neurons))
	for i, neuron := range l.Neurons {
		vals[i] = neuron.Activate(inputs)
	}
	return vals
}

func NewLayer(weights [][]float64, biases []float64,
	activationFunc ...func(float64) float64) Layer {
	actFunc := math.Tanh
	if len(activationFunc) > 0 {
		actFunc = activationFunc[0]
	}
	neurons := []Neuron{}
	for i, nWeights := range weights {
		neuron := NewNeuron(nWeights, biases[i], actFunc)
		neurons = append(neurons, neuron)
	}
	return Layer{Neurons: neurons}
}

type Network struct {
	Layers []Layer
}

func (n *Network) FeedForward(inputs []float64) []float64 {
	var vals []float64 = inputs
	for _, layer := range n.Layers {
		vals = layer.FeedForward(vals)
	}
	return vals
}

// Creates a new Network from the multidimensional slice of weights and
// biases, as well as the given activationFunc (or RELU by default).
func NewNetwork(weights [][][]float64, biases [][]float64,
	activationFunc ...func(float64) float64) Network {
	actFunc := math.Tanh
	if len(activationFunc) > 0 {
		actFunc = activationFunc[0]
	}
	layers := []Layer{}
	for j, nLayerWeights := range weights {
		layer := NewLayer(nLayerWeights, biases[j], actFunc)
		layers = append(layers, layer)
	}
	return Network{Layers: layers}
}

// Brain region.
type Region struct {
	Name       string
	Network    Network
	OutputHook func([]float64)
	Queue      chan []float64
}

// Creates a new Region with the name, network, queueSize, and optional output hook.
func NewRegion(name string, network Network, queueSize int, hook ...func([]float64)) Region {
	var outputhook func([]float64) = nil
	if len(hook) > 0 {
		outputhook = hook[0]
	}
	return Region{
		Name:       name,
		Network:    network,
		OutputHook: outputhook,
		Queue:      make(chan []float64, queueSize),
	}
}

// Adds the inputs to the processing queue. Drops if queue is full.
func (r *Region) AddInputs(inputs []float64) {
	select {
	case r.Queue <- inputs:
	default:
	}
}

// Pull inputs from queue and FeedForward each until queue is empty.
func (r *Region) Activate() {
	continueLoop := true
	for continueLoop {
		select {
		case inputs := <-r.Queue:
			result := r.Network.FeedForward(inputs)
			if r.OutputHook != nil {
				r.OutputHook(result)
			}
		default:
			continueLoop = false
		}
	}
}

// Brain structure. Controller processes inputs before forwarding to regions.
type Brain struct {
	Controller Network
	Regions    map[string]Region
	Queue      chan []float64
}

// Creates a new Brain with the controller, regions, and queueSize specified.
func NewBrain(controller Network, regions map[string]Region, queueSize int) Brain {
	return Brain{
		Controller: controller,
		Regions:    regions,
		Queue:      make(chan []float64, queueSize),
	}
}

// Adds the inputs to the processing queue. Drops if queue is full.
func (b *Brain) AddInputs(inputs []float64) {
	select {
	case b.Queue <- inputs:
	default:
	}
}

// Pull inputs from queue, FeedForward through Controller, then add result to
// queue for every region. Runs until Brain.Queue is empty.
func (b *Brain) Activate() {
	continueLoop := true
	for continueLoop {
		select {
		case inputs := <-b.Queue:
			inputs = b.Controller.FeedForward(inputs)
			for _, region := range b.Regions {
				region.AddInputs(inputs)
			}
		default:
			continueLoop = false
		}
	}
}

// Derivative for Tanh. Used in backpropagation. Tanh is default activation
// function used for the second half of epochs during automatic training
// (preceded by LeakyReLU).
func Tanhdx(x float64) float64 {
	return 1.0 - math.Pow(math.Tanh(x), 2.0)
}

// Rectified Linear Unit activation function.
func ReLU(x float64) float64 {
	if x > 0.0 {
		return x
	}
	return 0.0
}

// Derivative of ReLU. Used in backpropagation.
func RelUdx(x float64) float64 {
	if x > 0 {
		return 1.0
	}
	return 0.0
}

// Like ReLU but allows small negative values rather than replacing them with 0.
// Default where an activation function is an optional parameter and for first
// half of epochs during automatic training (math.Tanh thereafter).
func LeakyReLU(x float64) float64 {
	if x > 0.0 {
		return x
	}
	return 0.01 * x
}

// Derivative of LeakyReLU. Used in backpropagation.
func LeakyReLUdx(x float64) float64 {
	if x > 0.0 {
		return 1.0
	}
	return 0.01
}

// Like ReLU but smoother around 0.
func Softplus(x float64) float64 {
	return math.Log(1.0 + math.Exp(x))
}

// Derivative of Softplus. Used in backpropagation.
func Softplusdx(x float64) float64 {
	return 1.0 / (1.0 + math.Exp(-x))
}

// Exponential Linear Unit. Can be superior to ReLU for training classification
// models. Optional hyperparameter `a` set to 1.0 by default.`
func ELU(x float64, a ...float64) float64 {
	if x > 0.0 {
		return x
	}
	if len(a) > 0 {
		return a[0] * (math.Exp(x) - 1)
	}
	return math.Exp(x) - 1
}

// Returns an activation function that calls ELU with the supplied
// hyperparameter `a`.
func MakeELU(a float64) func(float64) float64 {
	return func(x float64) float64 {
		return ELU(x, a)
	}
}

// Derivative of ELU. Used in backpropagation.
func ELUdx(x float64, a ...float64) float64 {
	if x > 0.0 {
		return 1.0
	}
	if len(a) > 0 {
		return a[0] * math.Exp(x)
	}
	return math.Exp(x)
}

// Returns a derivation function that calls ELUdx with theh supplied
// hyperparameter `a`.
func MakeELUdx(a float64) func(float64) float64 {
	return func(x float64) float64 {
		return ELUdx(x, a)
	}
}

// Computes squared error loss for a given target `t` and observed result `z`.
func SEL(t float64, z float64) float64 {
	return math.Pow((t - z), 2)
}

// Computes the derivative of the squared error loss with respect to the
// difference between `t` and `z`.
func SELdx(t float64, z float64) float64 {
	x := t - z
	return 2 * x
}

// Computes the logistic loss for a given target `t` and observed result `z`.
func LogLoss(t float64, z float64) float64 {
	return -z*math.Log(t) - (1-z)*math.Log(1-t)
}

// Expresses a Gene as a Neuron. Gene.Bases encode Bias and Weights.
func ExpressGeneAsNeuron(gene *Gene[float64],
	activationFunc ...func(float64) float64) Neuron {
	gene.Mu.RLock()
	defer gene.Mu.RUnlock()
	actFunc := LeakyReLU
	if len(activationFunc) > 0 {
		actFunc = activationFunc[0]
	}
	if len(gene.Bases) == 0 {
		return Neuron{ActivationFunction: actFunc}
	} else if len(gene.Bases) == 1 {
		return Neuron{ActivationFunction: actFunc, Bias: gene.Bases[0]}
	}
	return Neuron{ActivationFunction: actFunc, Bias: gene.Bases[0],
		Weights: gene.Bases[1:]}
}

// Expresses an Allele as a neural Layer. Allele.Genes encode Neurons.
func ExpressAlleleAsLayer(allele *Allele[float64],
	activationFunc ...func(float64) float64) Layer {
	allele.Mu.RLock()
	defer allele.Mu.RUnlock()
	actFunc := LeakyReLU
	if len(activationFunc) > 0 {
		actFunc = activationFunc[0]
	}
	neurons := []Neuron{}
	for _, gene := range allele.Genes {
		neurons = append(neurons, ExpressGeneAsNeuron(gene, actFunc))
	}
	return Layer{Neurons: neurons}
}

// Expresses a Chromosome as a neural Network. Chromosome.Alleles encode Layers.
func ExpressChromosomeAsNetwork(chromosome *Chromosome[float64],
	activationFunc ...func(float64) float64) Network {
	chromosome.Mu.RLock()
	defer chromosome.Mu.RUnlock()
	actFunc := LeakyReLU
	if len(activationFunc) > 0 {
		actFunc = activationFunc[0]
	}
	layers := []Layer{}
	for _, allele := range chromosome.Alleles {
		layers = append(layers, ExpressAlleleAsLayer(allele, actFunc))
	}
	return Network{Layers: layers}
}
