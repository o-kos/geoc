package geoc

import (
	"errors"
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
		input       string
		expectedLat float64
		expectedLon float64
		expectedErr error
	}{
		// Positive cases
		{`48-33N; 048-33.0E`, 48.55, 48.55, nil},
		{`48-3327N; 120-5749E`, 48.5575, 120.963611, nil},
		{`48-33-27N; 120-5749E`, 48.5575, 120.963611, nil},
		{`48°33'26,9604"N; 48-33-26.9604E`, 48.557489, 48.557489, nil},
		{`48°33'27"N; 48-33-27 E`, 48.5575, 48.5575, nil},
		{`48-33,00'N; 48°33'E`, 48.55, 48.55, nil},

		// Negative cases
		{`48N; 48N`, 0, 0, ErrInvalidString},
		{`48-3327N; 120-5760E`, 0, 0, ErrOutOfRange},
		{`48-3327N; 48°33.4493'E`, 0, 0, ErrInvalidString},
		{`48-33'N; 48.557489`, 0, 0, ErrInvalidString},
	}

	for _, tc := range testCases {
		point, err := ParsePoint(tc.input)
		if tc.expectedErr != nil {
			if err == nil {
				t.Errorf("Expected error for %q string, got nil", tc.input)
				continue
			}
			if !errors.Is(err, tc.expectedErr) {
				t.Errorf("For %q: expected error %q, got %q", tc.input, tc.expectedErr, err)
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
		expectedError bool
	}{
		// DMS formats
		{`48-33.27N`, `48°33'27"N`, "", false},
		{`48-33.269604E`, `48°33'26.9604"E`, "", false},
		{`48-55.7489N`, `48°33'26,9604"N`, "", false},
		{`48-55.7489N`, `48-33-27N`, "", false},
		{`48-55.7489N`, `48-33-27 N`, "", false},
		{`120-5749E`, `120-5749E`, "", false},

		// Negative coord flips location letter
		{`-48.557489`, `48°33'27"N`, `48°33'27"S`, false},
		{`-120.963611`, `120°57'49"E`, `120°57'49"W`, false},

		// MinDec formats
		{`48.557489`, `48°33.4493'N`, "", false},
		{`48.55`, `48°33'N`, "", false},
		{`48.55`, `48-33N`, "", false},

		// DegDec formats
		{`48.557489`, `48.557489`, "", false},
		{`48.557489`, `48,557489`, "", false},
		{`-48.557489`, `-48.557489`, "", false},
		{`48.0`, `48N`, "", false},
		{`48.0`, `48 N`, "", false},
		{`48.0`, `48`, "", false},

		// Negative cases
		{`48.0`, `invalid`, "", true},
		{`91.0`, `48°33'27"N`, "", true},    // latitude out of range
		{`-181.0`, `120°57'49"E`, "", true}, // longitude out of range
	}

	for _, tc := range testCases {
		coord, err := ParseCoord(tc.coord)
		if err != nil {
			t.Errorf("Failed to parse coord %q: %v", tc.coord, err)
			continue
		}
		result, err := coord.Format(tc.example)
		if tc.expectedError {
			if err == nil {
				t.Errorf("Expected error for coord=%q, example=%q, got nil", tc.coord, tc.example)
			}
		} else {
			if err != nil {
				t.Errorf("Error %v for coord=%q, example=%q", err, tc.coord, tc.example)
				continue
			}
			expected := tc.expected
			if expected == "" {
				expected = tc.example
			}
			if result != expected {
				t.Errorf("For coord=%q, example=%q: expected %q, got %q", tc.coord, tc.example, expected, result)
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
		input       string
		expectedLat float64
		expectedLon float64
		expectedErr error
	}{
		// Positive cases
		{`48-33-27N; 120-5749E`, 48.5575, 120.963611, nil},
		{`48-33N; 048-33.0E`, 48.55, 48.55, nil},
		{`48°33'26,9604"N; 48-33-26.9604E`, 48.557489, 48.557489, nil},
		{`48-33-27N 120-5749E`, 48.5575, 120.963611, nil},

		// Negative cases
		{`48N; 48N`, 0, 0, ErrOutOfRange},               // both lat
		{`48-3327N; 48°33.4493'E`, 0, 0, ErrOutOfRange}, // format mismatch
		{`invalid`, 0, 0, ErrOutOfRange},
	}

	for _, tc := range testCases {
		point, err := ParsePoint(tc.input)
		if tc.expectedErr != nil {
			if err == nil {
				t.Errorf("Expected error for %q, got nil", tc.input)
			}
			if !errors.Is(err, tc.expectedErr) {
				t.Errorf("For %q: expected error %v, got %v", tc.input, tc.expectedErr, err)
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
