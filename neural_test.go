package bluegenes

import (
	"math"
	"testing"
)

func fuzzyEqual(f1, f2 float64) bool {
	diff := math.Abs(f1 - f2)
	avg := math.Abs(f1+f2) / 2.0
	return diff/avg < 0.01
}

func TestNeural(t *testing.T) {
	t.Run("Neuron", func(t *testing.T) {
		t.Run("Activate", func(t *testing.T) {
			t.Parallel()
			neuron := Neuron{
				Weights:            []float64{0.1, 0.2},
				Bias:               0.1,
				ActivationFunction: math.Tanh,
			}
			observed := neuron.Activate([]float64{1.0, 1.0})
			expected := math.Tanh(0.1 + 0.1 + 0.2)
			if !fuzzyEqual(expected, observed) {
				t.Fatal("Neuron.Activate returned wrong value")
			}
		})
		t.Run("NewNeuron", func(t *testing.T) {
			t.Parallel()
			neuron := NewNeuron([]float64{0.1, 0.2}, 0.1, math.Tanh)
			observed := neuron.Activate([]float64{1.0, 1.0})
			expected := math.Tanh(0.1 + 0.1 + 0.2)
			if !fuzzyEqual(expected, observed) {
				t.Fatal("Neuron.Activate returned wrong value")
			}
			// default value
			neuron = NewNeuron([]float64{0.1, 0.2}, 0.1)
			observed = neuron.Activate([]float64{1.0, 1.0})
			expected = math.Tanh(0.1 + 0.1 + 0.2)
			if !fuzzyEqual(expected, observed) {
				t.Fatal("Neuron.Activate returned wrong value")
			}
		})
	})
	t.Run("Layer", func(t *testing.T) {
		t.Parallel()
		neurons := []Neuron{
			NewNeuron([]float64{0.1, 0.2}, 0.1),
			NewNeuron([]float64{0.3, 0.4}, 0.3),
		}
		layer := Layer{
			Neurons: neurons,
		}
		result := layer.FeedForward([]float64{1.0, 1.0})
		if len(result) != 2 {
			t.Fatalf("expected 2 results, found %d\n", len(result))
		}
		expected := math.Tanh(0.1 + 0.1 + 0.2)
		observed := result[0]
		if !fuzzyEqual(expected, observed) {
			t.Errorf("expected %f, observed %f\n", expected, observed)
		}

		expected = math.Tanh(0.3 + 0.3 + 0.4)
		observed = result[1]
		if !fuzzyEqual(expected, observed) {
			t.Errorf("expected %f, observed %f\n", expected, observed)
		}

		layer = NewLayer([][]float64{{0.1, 0.2}, {0.3, 0.4}}, []float64{0.1, 0.3})
		result = layer.FeedForward([]float64{2.0, 2.0})
		if len(result) != 2 {
			t.Fatalf("expected 2 results, found %d\n", len(result))
		}
		expected = math.Tanh(0.1 + 2.0*0.1 + 2.0*0.2)
		observed = result[0]
		if !fuzzyEqual(expected, observed) {
			t.Errorf("expected %f, observed %f\n", expected, observed)
		}

		expected = math.Tanh(0.3 + 2.0*0.3 + 2.0*0.4)
		observed = result[1]
		if !fuzzyEqual(expected, observed) {
			t.Errorf("expected %f, observed %f\n", expected, observed)
		}
	})
	t.Run("Network", func(t *testing.T) {
		t.Parallel()
		neurons1 := []Neuron{
			NewNeuron([]float64{0.1, 0.2}, 0.1),
			NewNeuron([]float64{0.3, 0.4}, 0.3),
		}
		neurons2 := []Neuron{NewNeuron([]float64{0.5, 0.5}, 0.0)}
		layer1 := Layer{
			Neurons: neurons1,
		}
		layer2 := Layer{
			Neurons: neurons2,
		}
		network := Network{Layers: []Layer{layer1, layer2}}
		result := network.FeedForward([]float64{1.0, 1.0})
		if len(result) != 1 {
			t.Fatalf("expected 1 results, found %d\n", len(result))
		}
		expected := math.Tanh(0.5*math.Tanh(0.1+0.1+0.2) + 0.5*math.Tanh(0.3+0.3+0.4))
		observed := result[0]
		if !fuzzyEqual(expected, observed) {
			t.Errorf("expected %f, observed %f\n", expected, observed)
		}

		network = NewNetwork([][][]float64{
			{{0.1, 0.2}, {0.3, 0.4}},
			{{0.5, 0.5}},
		}, [][]float64{
			{0.1, 0.3},
			{0.0},
		})
		result = network.FeedForward([]float64{1.0, 1.0})
		if len(result) != 1 {
			t.Fatalf("expected 1 results, found %d\n", len(result))
		}
		expected = math.Tanh(0.5*math.Tanh(0.1+0.1+0.2) + 0.5*math.Tanh(0.3+0.3+0.4))
		observed = result[0]
		if !fuzzyEqual(expected, observed) {
			t.Errorf("expected %f, observed %f\n", expected, observed)
		}
	})
}
