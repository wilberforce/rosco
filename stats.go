package rosco

import (
	"math"

	"github.com/montanaflynn/stats"
	log "github.com/sirupsen/logrus"
	"gonum.org/v1/gonum/stat"
)

// Stats structure
type Stats struct {
	Name        string  // name of the metric
	Value       float64 // value
	Max         float64
	Min         float64
	Mean        float64
	Stddev      float64
	TrendSlope  float64
	Trend       float64
	Oscillation float64
}

const (
	lambdaOscillationSwing = 300
	lambdaVoltageMetric    = "LambdaVoltage"
)

// NewStats generates stats from a sample of float64 values
func NewStats(name string, data []float64) *Stats {
	// the sample stats
	s := &Stats{
		Name:  name,
		Value: data[len(data)-1],
	}

	// get the sample stats
	s.Min, s.Max = findMinAndMax(data)
	s.TrendSlope, s.Trend = linearRegression(data)
	s.Mean, _ = stats.Mean(data)
	s.Stddev, _ = stats.StandardDeviation(data)

	// get the oscillation of the lambda voltage
	if name == lambdaVoltageMetric {
		s.Oscillation = countOscillations(data, lambdaOscillationSwing)
	}

	s.Min = convertNaNandRound(s.Min)
	s.Max = convertNaNandRound(s.Max)
	s.Mean = convertNaNandRound(s.Mean)
	s.Stddev = convertNaNandRound(s.Stddev)
	s.Trend = convertNaNandRound(s.Trend)
	s.TrendSlope = convertNaNandRound(s.TrendSlope)
	s.Oscillation = convertNaNandRound(s.Oscillation)

	log.Debugf("stats %+v", *s)
	return s
}

func convertNaNandRound(metric float64) float64 {
	if math.IsNaN(metric) {
		return 0.0
	}

	return math.Round(metric*100) / 100
}

func findMinAndMax(data []float64) (min float64, max float64) {
	min = data[0]
	max = data[0]
	for _, value := range data {
		if value < min {
			min = value
		}
		if value > max {
			max = value
		}
	}
	return min, max
}

func countOscillations(data []float64, swing float64) float64 {
	count := len(data)
	oscillations := 0.0

	if count > 1 {
		// get the first value in the array
		prev := data[0]

		// start at the second value
		for i := 1; i < count; i++ {
			// if the current value has moved +/- the swing value
			// count that as an oscillation
			if data[i] < (prev-swing) || data[i] > (prev+swing) {
				oscillations++
			}
		}
	}

	return oscillations
}

func linearRegression(data []float64) (float64, float64) {
	xs := make([]float64, len(data))
	for i := range xs {
		xs[i] = float64(i)
	}

	ys := data

	return stat.LinearRegression(xs, ys, nil, false)
}
