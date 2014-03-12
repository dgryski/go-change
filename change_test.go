package change

import "testing"

func TestDetectChange(t *testing.T) {

	var tests = []struct {
		w   []float64
		idx int
	}{
		{
			[]float64{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
			0, // no change point found
		},

		{
			[]float64{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2},
			10, // the 1 before the first 2, due the scale factor
		},
		{
			[]float64{1, 1, 2, 2, 1, 1, 2, 2, 1, 1, 2, 3, 0, 1, 2, 2, 1, 1, 2, 2, 1, 1, 2},
			0, // change occurs but not statistically significant
		},
	}

	var detector = Detector{
		MinSampleSize: 5,
	}

	for _, tt := range tests {
		r := detector.Check(tt.w)
		if (r == nil || r.Confidence < 0.95) && tt.idx == 0 {
			// no difference found and no difference expected -- good
		} else if r.Confidence >= 0.95 && r.Index == tt.idx {
			// difference found at expected location -- good
		} else {
			t.Errorf("DetectChange confidence=%f index=%d, wanted %d", r.Confidence, r.Index, tt.idx)
		}
	}
}
