package change

import "testing"

func TestDetectChange(t *testing.T) {

	var tests = []struct {
		w   []float64
		idx int
	}{
		{
			[]float64{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
			0, // no change point found
		},
		{
			[]float64{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2},
			10, // the first 2
		},
	}

	var detector = Detector{
		MinSampleSize:      5,
		MinCorrelationCoef: 0.8,
		MarkerWidth:        10,
	}

	for _, tt := range tests {
		r := detector.Check(tt.w)
		if r == nil && tt.idx == 0 {
			t.Log("Check(): no change expected, no change found")
		} else if r == nil {
			t.Errorf("Check(): nothing, wanted %d", tt.idx)
		} else if r.Index == tt.idx {
			t.Logf("Check(): corr=%f index=%d", r.Correlation, r.Index)
		} else {
			t.Errorf("Check(): corr=%f index=%d, wanted %d", r.Correlation, r.Index, tt.idx)
		}
	}
}
