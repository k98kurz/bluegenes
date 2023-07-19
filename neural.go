package bluegenes

import "math"

type Neuron struct {
	Weights            []float64
	Bias               float64
	ActivationFunction func(float64) float64
	Value              float64
}

func (n Neuron) Activate(inputs []float64) float64 {
	max_size, _ := max(len(inputs), len(n.Weights))
	min_size, _ := min(len(inputs), len(n.Weights))
	vals := make([]float64, max_size)
	for i := 0; i < min_size; i++ {
		vals[i] = inputs[i] * n.Weights[i]
	}
	total := reduce(vals, func(f1, f2 float64) float64 { return f1 + f2 })

	n.Value = n.ActivationFunction(total + n.Bias)
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

func (l Layer) FeedForward(inputs []float64) []float64 {
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

func (n Network) FeedForward(inputs []float64) []float64 {
	var vals []float64 = inputs
	for _, layer := range n.Layers {
		vals = layer.FeedForward(vals)
	}
	return vals
}

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
