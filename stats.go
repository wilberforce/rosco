package rosco

import (
	"github.com/montanaflynn/stats"
	log "github.com/sirupsen/logrus"
	"gonum.org/v1/gonum/stat"
	"math"
)

// Stats structure
type Stats struct {
	Name        string
	Value       float64
	Count       int
	Max         float64
	Min         float64
	Mean        float64
	Stddev      float64
	Mode        float64
	ModeCount   float64
	TrendSlope  float64
	Trend       float64
	Oscillation float64
}

// NewStats generates stats from a sample of float64 values
func NewStats(name string, data []float64) *Stats {
	// the sample stats
	s := &Stats{
		Name:  name,
		Value: data[len(data)-1],
	}

	s.Count = len(data)

	// get the sample stats
	s.Min, s.Max = findMinAndMax(data)
	s.Mode, s.ModeCount = stat.Mode(data, nil)
	s.TrendSlope, s.Trend = linearRegression(data)
	s.Mean, _ = stats.Mean(data)
	s.Stddev, _ = stats.StandardDeviation(data)
	s.Oscillation, _ = stats.AutoCorrelation(data, 10)

	s.Min = convertNaNandRound(s.Min)
	s.Max = convertNaNandRound(s.Max)
	s.Mean = convertNaNandRound(s.Mean)
	s.Stddev = convertNaNandRound(s.Stddev)
	s.Mode = convertNaNandRound(s.Mode)
	s.ModeCount = convertNaNandRound(s.ModeCount)
	s.Trend = convertNaNandRound(s.Trend)
	s.TrendSlope = convertNaNandRound(s.TrendSlope)
	s.Oscillation = convertNaNandRound(s.Oscillation)

	log.Infof("stats %+v", *s)
	return s
}

func convertNaNandRound(metric float64) float64 {
	if math.IsNaN(metric) {
		return float64(0.00)
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

func linearRegression(data []float64) (float64, float64) {
	origin := false

	xs := make([]float64, len(data))
	for i := range xs {
		xs[i] = float64(i)
	}

	ys := data

	return stat.LinearRegression(xs, ys, nil, origin)
}
