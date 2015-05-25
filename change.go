// Package change implements an online change detection algorithm
package change

import (
	"errors"
	"math"
)

// ChangePoint is a potential change point found by Check().
type ChangePoint struct {
	Index       int
	Correlation float64
}

// Detector is a change detector.
type Detector struct {
	MinCorrelationCoef float64
	MinSampleSize      int
	MarkerWidth        int
}

// Check returns the index of a potential change point
func (d *Detector) Check(window []float64) *ChangePoint {
	n := len(window)

	marked := make([]float64, n)
	marked[n-1] = 1.0

	step := 1.0 / float64(d.MarkerWidth)
	start := n - d.MarkerWidth

	for i := start; i < n-1; i++ {
		marked[i] = marked[i-1] + step
	}

	corr := correlate(linearMeanFilter(differences(window), d.MarkerWidth), marked)
	if math.Abs(corr) > d.MinCorrelationCoef {
		return &ChangePoint{
			Index:       start,
			Correlation: corr,
		}
	}

	return nil
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

	sd := sd1 * sd2
	if sd == 0 {
		return sd
	}

	return cov / sd
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

// Stream monitors a stream of floats for changes
type Stream struct {
	windowSize int
	blockSize  int

	data []float64

	items int

	buffer []float64
	bufidx int

	detector *Detector
}

// NewStream constructs a new stream detector
func NewStream(windowSize, minSample, blockSize, width int, correlation float64) (*Stream, error) {
	if width > windowSize {
		return nil, errors.New("marker width cannot larger than window size")
	}

	detector := &Stream{
		windowSize: windowSize,
		blockSize:  blockSize,
		data:       make([]float64, windowSize),
		buffer:     make([]float64, blockSize),

		detector: &Detector{
			MinSampleSize:      minSample,
			MinCorrelationCoef: correlation,
			MarkerWidth:        width,
		},
	}

	return detector, nil
}

// Push adds a float to the stream and calls the change detector
func (s *Stream) Push(item float64) *ChangePoint {
	s.buffer[s.bufidx] = item
	s.bufidx++
	s.items++

	if s.bufidx < s.blockSize {
		return nil
	}

	copy(s.data[0:], s.data[s.blockSize:])
	copy(s.data[s.windowSize-s.blockSize:], s.buffer)
	s.bufidx = 0

	if s.items < s.windowSize {
		return nil
	}

	return s.detector.Check(s.data)
}

// Window returns the current data window.  This should be treated as read-only
func (s *Stream) Window() []float64 { return s.data }
