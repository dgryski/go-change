package offline

import (
	"errors"
	"math"
)

type ChangePoint struct {
	Index       int
	Correlation float64
}

type Detector struct {
	MarkerWidth        int
	MinCorrelationCoef float64
}

func (d *Detector) Check(series []float64) ([]ChangePoint, error) {
	n := len(series)
	width := d.MarkerWidth

	if width%2 == 0 {
		return nil, errors.New("marker width must be odd")
	}

	if width > n {
		return nil, errors.New("marker width cannot be larger than the series size")
	}

	series = linearMeanFilter(differences(series), width)

	resChan := make(chan ChangePoint)

	for i := 0; i < n; i++ {
		go func(pos int, resChan chan<- ChangePoint) {
			marked := createMarkedSeries(n, pos, width)

			corr := correlate(series, marked)
			resChan <- ChangePoint{pos, corr}
		}(i, resChan)
	}

	var changes []ChangePoint
	for i := 0; i < n; i++ {
		change := <-resChan

		if math.Abs(change.Correlation) > d.MinCorrelationCoef {
			changes = append(changes, change)
		}
	}

	return changes, nil
}

func correlate(series1, series2 []float64) float64 {
	m1, sd1 := stats(series1)
	m2, sd2 := stats(series2)

	n := len(series1)
	cov := 0.0
	for i := 0; i < n; i++ {
		cov += (series1[i] - m1) * (series2[i] - m2)
	}
	cov /= float64(n)

	return cov / (sd1 * sd2)
}

func stats(series []float64) (float64, float64) {
	sum, sqsum := 0.0, 0.0
	for _, val := range series {
		sum += val
		sqsum += val * val
	}

	mean := sum / float64(len(series))
	stddev := math.Sqrt(sqsum/float64(len(series)) - mean*mean)

	return mean, stddev
}

func differences(series []float64) []float64 {
	n := len(series)

	if n == 1 {
		return series
	}

	d := make([]float64, n)
	d[0] = series[1] - series[0]

	for i := 1; i < n; i++ {
		d[i] = series[i] - series[i-1]
	}

	return d
}

// take one point (kernel) and some points around it (support), average them
func linearMeanFilter(series []float64, width int) []float64 {
	if width == 1 {
		return series
	}

	support := int(math.Floor(float64(width) / 2.0)) // that many points around the kernel

	n := len(series)
	n2 := n + support*2

	// symmetrically extended series
	// needed to obtain the full window for the first few entries
	ext := make([]float64, n2)

	// left extension
	for i := 0; i < support; i++ {
		ext[i] = series[i]
	}

	// body
	copy(ext[support:], series)

	// right extension
	for i := n2 - support; i < n2; i++ {
		ext[i] = series[i-support*2]
	}

	for i := 0; i < n; i++ {
		mean, _ := stats(ext[i : i+support*2+1])

		series[i] = mean
	}

	return series
}

func createMarkedSeries(n, pos, width int) []float64 {
	if pos > n-1 || pos < 0 {
		// should never happen
		panic("marker position out of boundaries")
	}

	d := make([]float64, n)
	for i := 0; i < n; i++ {
		d[i] = 0.0
	}
	// set peak value
	d[pos] = 1.0

	if width == 1 {
		return d
	}

	around := int(math.Ceil(float64(width)/2.0)) - 1 // that many points around the peak
	step := 1.0 / float64(around+1)

	// go left from the peak
	cnt := 1
	for i := pos - around; i < pos; i++ {
		if i < 0 {
			continue
		}
		d[i] = step * float64(cnt)
		cnt += 1
	}

	// go right from the peak
	cnt = 1
	for i := pos + 1; i <= pos+around; i++ {
		if i > n-1 {
			break
		}
		d[i] = 1.0 - step*float64(cnt)
		cnt += 1
	}

	return d
}
