package geoc

import (
	"math"
	"strings"
	"testing"
)

type testCase struct {
	input         string
	expectedCoord float64
	expectedError bool
}

func TestStringToCoord(t *testing.T) {
	// Positive cases
	testCases := []testCase{
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
		{`98`, 98, false},
		{`98E`, 98, false},
		{`120-5749E`, 120.963611, false},

		// Negative cases
		{`98N`, 0, true},
		{`120-5760E`, 0, true},
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
				t.Errorf("Error %v, but excepted %f ", err, tc.expectedCoord)
				continue
			}
			if math.Abs(coord-tc.expectedCoord) > 0.000001 {
				t.Errorf("For string %q expected coord is %f, but got %f", tc.input, tc.expectedCoord, coord)
			}
		}
	}
}

type testCasePoint struct {
	input         string
	expectedLat   float64
	expectedLon   float64
	expectedError bool
}

func TestStringToPoint(t *testing.T) {
	testCases := []testCasePoint{
		// Positive cases
		{`48-33N; 048-33.0E`, 48.55, 48.55, false},
		{`48-3327N; 120-5749E`, 48.5575, 120.963611, false},
		{`48-33-27N; 120-5749E`, 48.5575, 120.963611, false},
		{`48°33'26,9604"N; 48-33-26.9604E`, 48.557489, 48.557489, false},
		{`48°33'27"N; 48-33-27 E`, 48.5575, 48.5575, false},
		{`48-33,00'N; 48°33'E`, 48.55, 48.55, false},

		// Negative cases
		{`48N; 48N`, 0, 0, true},
		{`48-3327N; 120-5760E`, 0, 0, true},
		{`48-3327N; 48°33.4493'E`, 0, 0, true},
		{`48-33'N; 48.557489`, 0, 0, true},
	}

	for _, tc := range testCases {
		cl := strings.Split(tc.input, "; ")
		if len(cl) != 2 {
			t.Errorf("Invalid input %q", tc.input)
			continue
		}

		point, err := StringToPoint(cl[0], cl[1])
		if tc.expectedError {
			if err == nil {
				t.Errorf("Expected error for %q string, got nil", tc.input)
			}
		} else {
			if err != nil {
				t.Errorf("Error %v, but excepted [%.2f %2f]", err, tc.expectedLat, tc.expectedLon)
				continue
			}

			if math.Abs(point.Lat-tc.expectedLat) > 0.000001 {
				t.Errorf("For string %q expected lat is %f, but got %f", tc.input, tc.expectedLat, point.Lat)
			}
			if math.Abs(point.Lon-tc.expectedLon) > 0.000001 {
				t.Errorf("For string %q expected lat is %f, but got %f", tc.input, tc.expectedLon, point.Lon)
			}
		}
	}
}
