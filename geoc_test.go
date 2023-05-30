package geoc

import (
	"math"
	"testing"
)

func TestStringToCoord(t *testing.T) {
	testCases := []struct {
		input         string
		expectedCoord float64
		expectedError bool
	}{
		{`48°33'26.9604"N`, 48.557489, false},
		{`48°33'26,9604"N`, 48.557489, false},
		{`48-33-26.9604N`, 48.557489, false},
		{`48°33'27"N`, 48.5575, false},
		{`48-33-27 N`, 48.5575, false},
		{`48-3327N`, 48.5575, false},
		{`48°33.4493'N`, 48.557488, false},
		{`48-33.4493'N`, 48.557488, false},
		{`48-33,00'N`, 48.55, false},
		{`48°33'N`, 48.55, false},
		{`48-33'N`, 48.55, false},
		{`48.557489`, 48.557489, false},
		{`48,557489`, 48.557489, false},
		{`-48.557489`, -48.557489, false},
		{`48`, 48, false},
		{`48N`, 48, false},
		{`48 N`, 48, false},
		{`48  ° N`, 48, false},
		{`-48`, -48, false},
		{`+48`, 48, false},

		{`48"N`, 0, true},
		{`48'N`, 0, true},
		{`-48N`, 0, true},
		{`+48N`, 0, true},
		{`48°33.4493"N`, 0, true},
		{"invalid string", 0, true},
	}

	for _, tc := range testCases {
		coord, err := StringToCoord(tc.input)
		if tc.expectedError {
			if err == nil {
				t.Errorf("Expected error for %q string, got nil", tc.input)
			}
		} else {
			if err != nil {
				t.Errorf("Error %v, but excepted %f", err, tc.expectedCoord)
				continue
			}

			if math.Abs(coord-tc.expectedCoord) > 0.000001 {
				t.Errorf("For string %q expected coord is %f, but got %f", tc.input, tc.expectedCoord, coord)
			}
		}
	}
}
