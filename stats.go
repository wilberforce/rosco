package rosco

import (
	"fmt"
	"github.com/montanaflynn/stats"
	log "github.com/sirupsen/logrus"
	"gonum.org/v1/gonum/stat"
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
	s.Mean, s.Stddev = stat.MeanStdDev(data, nil)
	s.Mode, s.ModeCount = stat.Mode(data, nil)

	// round to 2 decimal places
	s.Mean, _ = stats.Mean(data)
	s.Stddev, _ = stats.StandardDeviation(data)
	s.Oscillation, _ = stats.AutoCorrelation(data, 10)

	s.Mean, _ = stats.Round(s.Mean, 2)
	s.Stddev, _ = stats.Round(s.Stddev, 2)
	s.Mode, _ = stats.Round(s.Mode, 2)
	s.TrendSlope, s.Trend = linearRegression(data)

	log.WithFields(log.Fields{"stats": fmt.Sprintf("%+v", *s)}).Info("diagnostic stats calculated")
	return s
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
