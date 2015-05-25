package offline

import "testing"

func TestCheck(t *testing.T) {

	var tests = []struct {
		w   []float64
		idx int
	}{
		{
			[]float64{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
			-1, // no change point found
		},
		{
			[]float64{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2},
			10, // the first 2
		},
		/*
			{
				[]float64{1, 1, 2, 2, 1, 1, 2, 2, 1, 1, 2, 3, 0, 1, 2, 2, 1, 1, 2, 2, 1, 1, 2},
				-1, // change occurs but not statistically significant
			},
		*/
	}

	var detector = Detector{
		MarkerWidth:        1,
		MinCorrelationCoef: 0.1,
	}

	for _, tt := range tests {
		changes, err := detector.Check(tt.w)
		if err != nil {
			t.Errorf("Check(%#v): %s", tt.w, err)
		}

		if len(changes) == 0 && tt.idx == -1 {
			// no difference found and no difference expected -- good
		} else if len(changes) == 1 && changes[0].Index == tt.idx {
			// difference found at expected location -- good
		} else {
			if tt.idx == -1 {
				t.Errorf("Check(%#v) => %#v, expected no change", tt.w, changes)
			} else {
				t.Errorf("Check(%#v) => %#v, expected one change at [%d]", tt.w, changes, tt.idx)
			}
		}
	}
}
