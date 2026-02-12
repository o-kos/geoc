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
		// Positive cases
		{`48.557489`, 48.557489, false},

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
		coord, err := ParseCoord(tc.input)
		if tc.expectedError {
			if err == nil {
				t.Errorf("Expected error for %q string, got nil", tc.input)
			}
		} else {
			if err != nil {
				t.Errorf("Error %v, but excepted %f ", err, tc.expectedCoord)
				continue
			}
			if math.Abs(coord.Value-tc.expectedCoord) > 0.000001 {
				t.Errorf("For string %q expected coord is %f, but got %f", tc.input, tc.expectedCoord, coord.Value)
			}
		}
	}
}

func TestStringToPoint(t *testing.T) {
	testCases := []struct {
		input         string
		expectedLat   float64
		expectedLon   float64
		expectedError bool
	}{
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
		point, err := ParsePoint(tc.input)
		if tc.expectedError {
			if err == nil {
				t.Errorf("Expected error for %q string, got nil", tc.input)
			}
		} else {
			if err != nil {
				t.Errorf("Error %v, but excepted [%.2f %2f]", err, tc.expectedLat, tc.expectedLon)
				continue
			}

			if math.Abs(point.Lat.Value-tc.expectedLat) > 0.000001 {
				t.Errorf("For string %q expected lat is %f, but got %f", tc.input, tc.expectedLat, point.Lat.Value)
			}
			if math.Abs(point.Lon.Value-tc.expectedLon) > 0.000001 {
				t.Errorf("For string %q expected lat is %f, but got %f", tc.input, tc.expectedLon, point.Lon.Value)
			}
		}
	}
}

func TestCoordToString(t *testing.T) {
	testCases := []struct {
		coord         string
		example       string
		expected      string // empty if equal to example
		expectedError error
	}{
		// DMS formats
		{`48-33.27N`, `48°33'27"N`, "", nil},
		{`48-33.269604E`, `48°33'26.9604"E`, "", nil},
		{`48-55.7489N`, `48°33'26,9604"N`, "", nil},
		{`48-55.7489N`, `48-33-27N`, "", nil},
		{`48-55.7489N`, `48-33-27 N`, "", nil},
		{`120-5749E`, `120-5749E`, "", nil},

		// Negative coord flips location letter
		{`-48.557489`, `48°33'27"N`, `48°33'27"S`, nil},
		{`-120.963611`, `120°57'49"E`, `120°57'49"W`, nil},

		// MinDec formats
		{`48.557489`, `48°33.4493'N`, "", nil},
		{`48.55`, `48°33'N`, "", nil},
		{`48.55`, `48-33N`, "", nil},

		// DegDec formats
		{`48.557489`, `48.557489`, "", nil},
		{`48.557489`, `48,557489`, "", nil},
		{`-48.557489`, `-48.557489`, "", nil},
		{`48.0`, `48N`, "", nil},
		{`48.0`, `48 N`, "", nil},
		{`48.0`, `48`, "", nil},

		// Negative cases
		{`48.0`, `invalid`, "", true},
		{`91.0`, `48°33'27"N`, "", true},    // latitude out of range
		{`-181.0`, `120°57'49"E`, "", true}, // longitude out of range
	}

	for _, tc := range testCases {
		result, err := Format(tc.coord, tc.example)
		if tc.expectedError {
			if err == nil {
				t.Errorf("Expected error for coord=%f, example=%q, got nil", tc.coord, tc.example)
			}
		} else {
			if err != nil {
				t.Errorf("Error %v for coord=%f, example=%q", err, tc.coord, tc.example)
				continue
			}
			if result != tc.expectedStr {
				t.Errorf("For coord=%f, example=%q: expected %q, got %q", tc.coord, tc.example, tc.expectedStr, result)
			}
		}
	}
}

func TestParseCoord(t *testing.T) {
	testCases := []struct {
		input         string
		expectedCoord float64
		expectedLoc   Location
		expectedError bool
	}{
		// DMS
		{`48°33'27"N`, 48.5575, Lat, false},
		{`48-33-27N`, 48.5575, Lat, false},
		{`120-5749E`, 120.963611, Lon, false},
		{`48°33'26,9604"N`, 48.557489, Lat, false},

		// MinDec
		{`48°33.4493'N`, 48.557488, Lat, false},
		{`48-33N`, 48.55, Lat, false},

		// DegDec
		{`48.557489`, 48.557489, None, false},
		{`48,557489`, 48.557489, None, false},
		{`-48.557489`, -48.557489, None, false},
		{`48N`, 48, Lat, false},
		{`98E`, 98, Lon, false},

		// Negative cases
		{`98N`, 0, None, true},
		{`invalid`, 0, None, true},
	}

	for _, tc := range testCases {
		coord, err := ParseCoord(tc.input)
		if tc.expectedError {
			if err == nil {
				t.Errorf("Expected error for %q, got nil", tc.input)
			}
		} else {
			if err != nil {
				t.Errorf("Error %v for %q", err, tc.input)
				continue
			}
			if math.Abs(coord.Value-tc.expectedCoord) > 0.000001 {
				t.Errorf("For %q: expected %f, got %f", tc.input, tc.expectedCoord, coord.Value)
			}
			if coord.Loc != tc.expectedLoc {
				t.Errorf("For %q: expected Loc=%d, got %d", tc.input, tc.expectedLoc, coord.Loc)
			}
		}
	}
}

func TestParsePoint(t *testing.T) {
	testCases := []struct {
		input         string
		expectedLat   float64
		expectedLon   float64
		expectedError bool
	}{
		// Positive cases
		{`48-33-27N; 120-5749E`, 48.5575, 120.963611, false},
		{`48-33N; 048-33.0E`, 48.55, 48.55, false},
		{`48°33'26,9604"N; 48-33-26.9604E`, 48.557489, 48.557489, false},
		{`48-33-27N 120-5749E`, 48.5575, 120.963611, false}, // no explicit separator

		// Negative cases
		{`48N; 48N`, 0, 0, true},               // both lat
		{`48-3327N; 48°33.4493'E`, 0, 0, true}, // format mismatch
		{`invalid`, 0, 0, true},
	}

	for _, tc := range testCases {
		point, err := ParsePoint(tc.input)
		if tc.expectedError {
			if err == nil {
				t.Errorf("Expected error for %q, got nil", tc.input)
			}
		} else {
			if err != nil {
				t.Errorf("Error %v for %q", err, tc.input)
				continue
			}
			if math.Abs(point.Lat.Value-tc.expectedLat) > 0.000001 {
				t.Errorf("For %q: expected lat %f, got %f", tc.input, tc.expectedLat, point.Lat.Value)
			}
			if math.Abs(point.Lon.Value-tc.expectedLon) > 0.000001 {
				t.Errorf("For %q: expected lon %f, got %f", tc.input, tc.expectedLon, point.Lon.Value)
			}
		}
	}
}
