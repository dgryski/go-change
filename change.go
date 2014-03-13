// Package change implements an online change detection algorithm
/*
http://excelsior.cs.ucsb.edu/papers/as06.pdf

How effective is this algorithm?  The authors state
"[T]he estimation algorithm can be shown theoretically to give the correct
prediction and approach an unbiased estimator of the true changeover point if
the window length approaches infinity and the two distributions are
sufficiently dissimilar."

The algorithm works by examining the distributions on either side of a
suspected change point, and finding the index in the window where the two
distributions are most dissimilar.  As there is not guaranteed to be a change
point in the window, this implementation also performs a Student's t-test on
the two distributions to reduce the rate of false positives.

*/
package change

import (
	"math"

	"github.com/dgryski/go-onlinestats"
)

// Stats are some descriptive statistics for a block of items.  It implements the interface needed by the t-test method of onlinestats.
type Stats struct {
	mean     float64
	variance float64
	n        int
}

// Mean returns the mean of the data set
func (s Stats) Mean() float64 { return s.mean }

// Var returns the variance of the data set
func (s Stats) Var() float64 { return s.variance }

// Len returns the number of items in the data set
func (s Stats) Len() int { return s.n }

// Stddev returns the standard deviation of the sample
func (s Stats) Stddev() float64 { return math.Sqrt(s.variance) }

// ChangePoint is a potential change point found by Check().
type ChangePoint struct {
	// Index is the offset into the data set of the suspected change point
	Index int

	// Difference is the difference in distribution means found by the Student's t-test
	Difference float64

	// Confidence is the confidence returned by a Student's t-test
	Confidence float64

	// Before is the statistics of the distribution before the change point
	Before Stats

	// After is the statistics of the distribution after the change point
	After Stats
}

// DefaultMinSampleSize is the minimum sample size to consider from the window being checked
const DefaultMinSampleSize = 30

// Detector is a change detector.
type Detector struct {
	MinSampleSize int
	MinConfidence float64
}

// Check returns the index of a potential change point
func (d *Detector) Check(window []float64) *ChangePoint {

	n := len(window)

	// The paper provides recursive formulas for computing the means and
	// standard deviations as we slide along the window.  This
	// implementation uses alternate math based on cumulative sums.

	// cumsum contains the cumulative sum of all elements <= i
	// cumsumsq contains the cumulative sum of squares of all elements <= i
	// TODO(dgryski): move this to a move numerically stable algorithm
	cumsum := make([]float64, n)
	cumsumsq := make([]float64, n)

	var sum, sumsq float64
	for i, v := range window {
		sum += v
		sumsq += v * v
		cumsum[i] = sum
		cumsumsq[i] = sumsq
	}

	// sb is our between-class scatter, the degree of dissimilarity of the
	// two distributions.  This value is always positive, so we can set 0
	// as the minimum and know that any valid value will be larger
	var maxsb float64
	var maxsbIdx int

	// The paper also provides a metric sw, for 'within-class scatter',
	// which depends on the standard-deviation of the samples. It suggests
	// finding the point that minimizes the ratio sw/sb.  However, it then
	// proves that this is equivalent to maximizing sb.  The calculation of
	// sb depends only on the means of the two samples, and not of the
	// variances.  However, we calculate the variances so that we can pass
	// them to the T test later on.

	var before, after Stats

	// sane default
	minSampleSize := d.MinSampleSize
	if minSampleSize == 0 {
		minSampleSize = DefaultMinSampleSize
	}

	for l := minSampleSize; l < (n - minSampleSize + 1); l++ {
		lidx := l - 1
		n1 := float64(l)
		mean1 := cumsum[lidx] / n1

		n2 := float64(n - l)
		sum2 := (sum - cumsum[lidx])
		mean2 := sum2 / n2

		sb := ((n1 * n2) / (n1 + n2)) * (mean1 - mean2) * (mean1 - mean2)
		if maxsb < sb {
			maxsb = sb
			maxsbIdx = l

			// The variances are calculated only if needed to
			// reduce the math in the main loop
			var1 := (cumsumsq[lidx] - (cumsum[lidx]*cumsum[lidx])/(n1)) / (n1 - 1)
			var2 := ((sumsq - cumsumsq[lidx]) - (sum2*sum2)/(n2)) / (n2 - 1)

			before.mean, before.variance, before.n = mean1, var1, l
			after.mean, after.variance, after.n = mean2, var2, n-l
		}
	}

	var conf float64
	if before.n > 0 {
		// we found a difference
		conf = onlinestats.Welch(before, after)
	}

	// not above our threshold
	if conf <= d.MinConfidence {
		return nil
	}

	cp := &ChangePoint{
		Index:      maxsbIdx,
		Difference: after.Mean() - before.Mean(),
		Confidence: conf,
		Before:     before,
		After:      after,
	}

	return cp
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
func NewStream(windowSize int, minSample int, blockSize int, confidence float64) *Stream {
	return &Stream{
		windowSize: windowSize,
		blockSize:  blockSize,
		data:       make([]float64, windowSize),
		buffer:     make([]float64, blockSize),

		detector: &Detector{
			MinSampleSize: minSample,
			MinConfidence: confidence,
		},
	}
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
