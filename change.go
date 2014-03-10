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

import "github.com/dgryski/go-onlinestats"

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

// ChangePoint is a potential change point found by Check().
type ChangePoint struct {
	// Index is the offset into the data set of the suspected change point
	Index int

	// Difference is the difference in distribution means found by the Student's t-test
	Difference float64

	// Before is the statistics of the distribution before the change point
	Before Stats

	// After is the statistics of the distribution after the change point
	After Stats
}

// DefaultMinSampleSize is the minimum sample size to consider from the window being checked
const DefaultMinSampleSize = 30

// Detector is a change detector.  The default confidence level is passing a t-test with 80% confidence.
type Detector struct {
	MinSampleSize int
	TConf         onlinestats.Confidence
}

// Check returns the index of a potential change point
func (d *Detector) Check(window []float64) *ChangePoint {

	n := len(window)

	// The paper provides recursive formulas for computing the means and
	// standard deviations as we slide along the window.  This
	// implementation uses alternate math based on cumulative sums.

	// cumsum contains the cumulative sum of all elements <= i
	// cumsumsq contains the cumulative sum of squares of all elements <= i
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
		n1 := float64(l + 1)
		mean1 := cumsum[l] / n1

		n2 := float64(n - l - 1)
		sum2 := (sum - cumsum[l])
		mean2 := sum2 / n2

		sb := ((n1 * n2) / (n1 + n2)) * (mean1 - mean2) * (mean1 - mean2)
		if maxsb < sb {
			maxsb = sb
			maxsbIdx = l

			// The variances are calculated only if needed to
			// reduce the main in the main loop
			var1 := (cumsumsq[l] - (cumsum[l]*cumsum[l])/(n1-1)) / (n1 - 1)
			var2 := ((sumsq - cumsumsq[l]) - (sum2*sum2)/(n2-1)) / (n2 - 1)

			before.mean, before.variance, before.n = mean1, var1, l+1
			after.mean, after.variance, after.n = mean2, var2, n-l-1
		}
	}

	var diff float64

	if before.n > 0 {
		// we found a difference
		diff = onlinestats.TTest(before, after, onlinestats.Confidence(d.TConf))
	}

	cp := &ChangePoint{
		Index:      maxsbIdx,
		Difference: diff,
		Before:     before,
		After:      after,
	}

	return cp
}

// Confidence levels for the Student's t-test
const (
	Conf80   = onlinestats.Conf80
	Conf90   = onlinestats.Conf90
	Conf95   = onlinestats.Conf95
	Conf98   = onlinestats.Conf98
	Conf99   = onlinestats.Conf99
	Conf99p5 = onlinestats.Conf99p5
)
