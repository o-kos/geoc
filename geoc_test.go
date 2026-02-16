package geoc

import (
	"errors"
	"math"
	"testing"
)

func TestCoordFormatPositive(t *testing.T) {
	testCases := []struct {
		coord    string
		example  string
		expected string
	}{
		// DMS formats
		{`48-33.27N`, `48°33'27"N`, `48°33'16"N`},
		{`48-33.269604E`, `48°33'26.9604"E`, `48°33'16.1762"E`},
		{`48-55.7489N`, `48°33'26,9604"N`, `48°55'44,9340"N`},
		{`48-55.7489N`, `48-33-27N`, `48-55-45N`},
		{`48-55.7489N`, `48-33-27 N`, `48-55-45 N`},
		{`120-5749E`, `120-5749E`, `120-5749E`},
		{`12.9999999`, `012-5749E`, `013-0000E`},
		{`48.05`, `48°03'04"N`, `48°03'00"N`},
		{`48.999999`, `48°33'27"N`, `49°00'00"N`},
		{`-48.55`, `48-33-27`, `-48-33-00`},
		{`48.55`, `+48-33-27`, `+48-33-00`},

		// Negative coord flips location letter
		{`-48.557489`, `48°33'27"N`, `48°33'27"S`},
		{`-120.963611`, `120°57'49"E`, `120°57'49"W`},

		// MinDec formats
		{`48.557489`, `48°33.4493'N`, `48°33.4493'N`},
		{`48.55`, `48°33'N`, `48°33'N`},
		{`48.55`, `48-33N`, `48-33N`},
		{`48.05`, `48-03.0N`, `48-03.0N`},
		{`48.999999`, `48-33.0N`, `49-00.0N`},
		{`179.9999999`, `180-00.0E`, `180-00.0E`},
		{`89.9999999`, `90-00.0N`, `90-00.0N`},
		{`-48.55`, `48-33`, `-48-33`},
		{`48.55`, `+48-33`, `+48-33`},

		// DegDec formats
		{`48.557489`, `48.557489`, `48.557489`},
		{`48.557489`, `48,557489`, `48,557489`},
		{`-48.557489`, `-48.557489`, `-48.557489`},
		{`48.557489`, `+48.557489`, `+48.557489`},
		{`48.0`, `48N`, `48N`},
		{`48.0`, `48 N`, `48 N`},
		{`48.0`, `48`, `48`},
	}

	for _, tc := range testCases {
		t.Run(tc.coord+" → "+tc.example, func(t *testing.T) {
			coord, err := ParseCoord(tc.coord)
			if err != nil {
				t.Fatalf("Failed to parse coord %q: %v", tc.coord, err)
			}
			result, err := coord.Format(tc.example)
			if err != nil {
				t.Fatalf("Error %v for coord=%q, example=%q", err, tc.coord, tc.example)
			}
			if result != tc.expected {
				t.Fatalf("For coord=%q, example=%q: expected %q, got %q", tc.coord, tc.example, tc.expected, result)
			}
		})
	}
}

func TestCoordFormatNegative(t *testing.T) {
	testCases := []struct {
		coord   string
		example string
	}{
		{`48.0`, `invalid`},
		{`90.1`, `48°33'27"N`}, // latitude out of range
	}

	for _, tc := range testCases {
		t.Run(tc.coord+" "+tc.example, func(t *testing.T) {
			coord, err := ParseCoord(tc.coord)
			if err != nil {
				t.Fatalf("Failed to parse coord %q: %v", tc.coord, err)
			}
			if _, err := coord.Format(tc.example); err == nil {
				t.Fatalf("Expected error for coord=%q, example=%q, got nil", tc.coord, tc.example)
			}
		})
	}
}

func TestCoordFormatNegativeDirect(t *testing.T) {
	testCases := []struct {
		name        string
		coord       Coord
		example     string
		expectedErr error
	}{
		{
			name:        "loc_none_out_of_range_positive",
			coord:       Coord{Value: 180.1, Loc: LocNone},
			example:     `180.0`,
			expectedErr: ErrOutOfRange,
		},
		{
			name:        "loc_none_out_of_range_negative",
			coord:       Coord{Value: -180.1, Loc: LocNone},
			example:     `180.0`,
			expectedErr: ErrOutOfRange,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.coord.Format(tc.example)
			if err == nil {
				t.Fatalf("Expected error for coord=%v, example=%q, got nil", tc.coord, tc.example)
			}
			if !errors.Is(err, tc.expectedErr) {
				t.Fatalf("Expected error %q, got %q", tc.expectedErr, err)
			}
		})
	}
}

func TestCoordString(t *testing.T) {
	testCases := []struct {
		name     string
		coord    Coord
		expected string
	}{
		{
			name:     "lat_positive",
			coord:    Coord{Value: 48.5575, Loc: LocLat},
			expected: "48-33.4N",
		},
		{
			name:     "lat_negative",
			coord:    Coord{Value: -48.5575, Loc: LocLat},
			expected: "48-33.4S",
		},
		{
			name:     "lon_positive_with_three_digit_deg",
			coord:    Coord{Value: 120.963611, Loc: LocLon},
			expected: "120-57.8E",
		},
		{
			name:     "lon_negative_with_three_digit_deg",
			coord:    Coord{Value: -120.963611, Loc: LocLon},
			expected: "120-57.8W",
		},
		{
			name:     "none_uses_decimal_degrees",
			coord:    Coord{Value: 48.557489, Loc: LocNone},
			expected: "48.557489",
		},
		{
			name:     "none_negative_uses_sign",
			coord:    Coord{Value: -48.557489, Loc: LocNone},
			expected: "-48.557489",
		},
		{
			name:     "boundary_lat_90",
			coord:    Coord{Value: 90, Loc: LocLat},
			expected: "90-00.0N",
		},
		{
			name:     "boundary_lon_180",
			coord:    Coord{Value: 180, Loc: LocLon},
			expected: "180-00.0E",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.coord.String()
			if result != tc.expected {
				t.Fatalf("Expected %q, got %q", tc.expected, result)
			}
		})
	}
}

func TestPointFormat(t *testing.T) {
	testCases := []struct {
		name        string
		point       Point
		latFmt      string
		lonFmt      string
		separator   string
		expected    string
		expectedErr error
	}{
		{
			name:      "dms_with_semicolon",
			point:     Point{Lat: Coord{Value: 48.5575, Loc: LocLat}, Lon: Coord{Value: 120.963611, Loc: LocLon}},
			latFmt:    `48°33'27"N`,
			lonFmt:    `120°57'49"E`,
			separator: "; ",
			expected:  `48°33'27"N; 120°57'49"E`,
		},
		{
			name:      "negative_flips_ns_we",
			point:     Point{Lat: Coord{Value: -48.5575, Loc: LocLat}, Lon: Coord{Value: -120.963611, Loc: LocLon}},
			latFmt:    `48°33'27"N`,
			lonFmt:    `120°57'49"E`,
			separator: " ",
			expected:  `48°33'27"S 120°57'49"W`,
		},
		{
			name:      "derive_location_from_format_when_none",
			point:     Point{Lat: Coord{Value: 48.55}, Lon: Coord{Value: 48.55}},
			latFmt:    `48-33N`,
			lonFmt:    `048-33.0E`,
			separator: " | ",
			expected:  `48-33N | 048-33.0E`,
		},
		{
			name:        "invalid_lat_format",
			point:       Point{Lat: Coord{Value: 48.5575, Loc: LocLat}, Lon: Coord{Value: 120.963611, Loc: LocLon}},
			latFmt:      `invalid`,
			lonFmt:      `120°57'49"E`,
			separator:   " ",
			expectedErr: ErrInvalidString,
		},
		{
			name:        "invalid_lon_format",
			point:       Point{Lat: Coord{Value: 48.5575, Loc: LocLat}, Lon: Coord{Value: 120.963611, Loc: LocLon}},
			latFmt:      `48°33'27"N`,
			lonFmt:      `invalid`,
			separator:   " ",
			expectedErr: ErrInvalidString,
		},
		{
			name:        "lat_out_of_range",
			point:       Point{Lat: Coord{Value: 90.1, Loc: LocLat}, Lon: Coord{Value: 120.963611, Loc: LocLon}},
			latFmt:      `48°33'27"N`,
			lonFmt:      `120°57'49"E`,
			separator:   " ",
			expectedErr: ErrOutOfRange,
		},
		{
			name:        "lon_out_of_range",
			point:       Point{Lat: Coord{Value: 48.5575, Loc: LocLat}, Lon: Coord{Value: 180.1, Loc: LocLon}},
			latFmt:      `48°33'27"N`,
			lonFmt:      `120°57'49"E`,
			separator:   " ",
			expectedErr: ErrOutOfRange,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := tc.point.Format(tc.latFmt, tc.lonFmt, tc.separator)
			if tc.expectedErr != nil {
				if err == nil {
					t.Fatalf("Expected error %q, got nil", tc.expectedErr)
				}
				if !errors.Is(err, tc.expectedErr) {
					t.Fatalf("Expected error %q, got %q", tc.expectedErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if result != tc.expected {
				t.Fatalf("Expected %q, got %q", tc.expected, result)
			}
		})
	}
}

func TestPointString(t *testing.T) {
	testCases := []struct {
		name     string
		point    Point
		expected string
	}{
		{
			name:     "default_positive",
			point:    Point{Lat: Coord{Value: 48.5575, Loc: LocLat}, Lon: Coord{Value: 120.963611, Loc: LocLon}},
			expected: "48-33.4N 120-57.8E",
		},
		{
			name:     "default_negative",
			point:    Point{Lat: Coord{Value: -48.5575, Loc: LocLat}, Lon: Coord{Value: -120.963611, Loc: LocLon}},
			expected: "48-33.4S 120-57.8W",
		},
		{
			name:     "derive_location_from_default_formats",
			point:    Point{Lat: Coord{Value: 48.55}, Lon: Coord{Value: 48.55}},
			expected: "48-33.0N 048-33.0E",
		},
		{
			name:     "boundary_lat_90",
			point:    Point{Lat: Coord{Value: 90, Loc: LocLat}, Lon: Coord{Value: 120.963611, Loc: LocLon}},
			expected: "90-00.0N 120-57.8E",
		},
		{
			name:     "boundary_lon_180",
			point:    Point{Lat: Coord{Value: 48.5575, Loc: LocLat}, Lon: Coord{Value: 180, Loc: LocLon}},
			expected: "48-33.4N 180-00.0E",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.point.String()
			if result != tc.expected {
				t.Fatalf("Expected %q, got %q", tc.expected, result)
			}
		})
	}
}

func TestParseCoordPositive(t *testing.T) {
	testCases := []struct {
		input         string
		expectedCoord float64
		expectedLoc   Location
	}{
		// DMS
		{`48°33'26.9604"N`, 48.557489, LocLat},
		{`48°33'27"N`, 48.5575, LocLat},
		{`48-33-27N`, 48.5575, LocLat},
		{`48-33-27 N`, 48.5575, LocLat},
		{`48-3327N`, 48.5575, LocLat},
		{`48-33-26.9604N`, 48.557489, LocLat},
		{`120-5749E`, 120.963611, LocLon},
		{`48°33'26,9604"N`, 48.557489, LocLat},

		// MinDec
		{`48°33.4493'N`, 48.557488, LocLat},
		{`48-33.4493'N`, 48.557488, LocLat},
		{`48-33,00'N`, 48.55, LocLat},
		{`48°33'N`, 48.55, LocLat},
		{`48-33'N`, 48.55, LocLat},
		{`48-33N`, 48.55, LocLat},

		// DegDec
		{`48`, 48, LocNone},
		{`48.557489`, 48.557489, LocNone},
		{`48,557489`, 48.557489, LocNone},
		{`-48.557489`, -48.557489, LocNone},
		{`-48`, -48, LocNone},
		{`+48`, 48, LocNone},
		{`98`, 98, LocNone},
		{`48N`, 48, LocLat},
		{`48 N`, 48, LocLat},
		{`48  ° N`, 48, LocLat},
		{`98E`, 98, LocLon},

		// Boundary values
		{`90N`, 90, LocLat},
		{`90S`, -90, LocLat},
		{`180E`, 180, LocLon},
		{`180W`, -180, LocLon},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			coord, err := ParseCoord(tc.input)
			if err != nil {
				t.Fatalf("Error %v for %q", err, tc.input)
			}
			if math.Abs(coord.Value-tc.expectedCoord) > 0.000001 {
				t.Fatalf("For %q: expected %f, got %f", tc.input, tc.expectedCoord, coord.Value)
			}
			if coord.Loc != tc.expectedLoc {
				t.Fatalf("For %q: expected Loc=%v, got %v", tc.input, tc.expectedLoc, coord.Loc)
			}
		})
	}
}

func TestParseCoordNegative(t *testing.T) {
	testCases := []struct {
		input         string
		expectedError error
	}{
		{`98N`, ErrOutOfRange},
		{`120-5760E`, ErrOutOfRange},
		{`90-01N`, ErrOutOfRange},
		{`180-01E`, ErrOutOfRange},
		{`48"N`, ErrInvalidString},
		{`48'N`, ErrInvalidString},
		{`-48N`, ErrInvalidCoord},
		{`+48N`, ErrInvalidCoord},
		{`48°33.4493"N`, ErrInvalidString},
		{`invalid`, ErrInvalidString},
		{`invalid string`, ErrInvalidString},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			_, err := ParseCoord(tc.input)
			if err == nil {
				t.Fatalf("Expected error for %q, got nil", tc.input)
			}
			if !errors.Is(err, tc.expectedError) {
				t.Fatalf("For %q: expected error %q, got %q", tc.input, tc.expectedError, err)
			}
		})
	}
}

func TestParsePointPositive(t *testing.T) {
	testCases := []struct {
		input       string
		expectedLat float64
		expectedLon float64
	}{
		{`48-33-27N; 120-5749E`, 48.5575, 120.963611},
		{`48-3327N; 120-5749E`, 48.5575, 120.963611},
		{`48-33N; 048-33.0E`, 48.55, 48.55},
		{`48°33'26,9604"N; 48-33-26.9604E`, 48.557489, 48.557489},
		{`48°33'27"N; 48-33-27 E`, 48.5575, 48.5575},
		{`48-33,00'N; 48°33'E`, 48.55, 48.55},
		{`48-33-27N 120-5749E`, 48.5575, 120.963611},
		{`48N120E`, 48, 120},
		{` 48N;120E `, 48, 120},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			point, err := ParsePoint(tc.input)
			if err != nil {
				t.Fatalf("Error %v for %q", err, tc.input)
			}
			if math.Abs(point.Lat.Value-tc.expectedLat) > 0.000001 {
				t.Fatalf("For %q: expected lat %f, got %f", tc.input, tc.expectedLat, point.Lat.Value)
			}
			if math.Abs(point.Lon.Value-tc.expectedLon) > 0.000001 {
				t.Fatalf("For %q: expected lon %f, got %f", tc.input, tc.expectedLon, point.Lon.Value)
			}
		})
	}
}

func TestParsePointNegative(t *testing.T) {
	testCases := []struct {
		input       string
		expectedErr error
	}{
		{`98N; 120E`, ErrOutOfRange}, // latitude parsing error
		{`48-3327N; 120-5760E`, ErrOutOfRange},
		{`48N`, ErrInvalidString},                    // too few coords
		{`48N; 120E; 10N`, ErrInvalidString},         // too many coords
		{`120E; 48N`, ErrInvalidString},              // bad latitude location
		{`48N; 48N`, ErrInvalidString},               // both lat
		{`48-3327N; 48°33.4493'E`, ErrInvalidString}, // format mismatch
		{`48-33'N; 48.557489`, ErrInvalidString},
		{`abc 48N 120E`, ErrInvalidString},
		{`x48N;120E`, ErrInvalidString},
		{`48N xxx 120E`, ErrInvalidString},
		{`48N;120Ezzz`, ErrInvalidString},
		{`invalid`, ErrInvalidString},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			_, err := ParsePoint(tc.input)
			if err == nil {
				t.Fatalf("Expected error for %q, got nil", tc.input)
			}
			if !errors.Is(err, tc.expectedErr) {
				t.Fatalf("For %q: expected error %q, got %q", tc.input, tc.expectedErr, err)
			}
		})
	}
}

func TestCoordGroupsErrorBranches(t *testing.T) {
	t.Run("newCoordGroups_too_many_coords", func(t *testing.T) {
		_, err := newCoordGroups("48N 49N")
		if err == nil || !errors.Is(err, ErrInvalidString) {
			t.Fatalf("Expected ErrInvalidString, got %v", err)
		}
	})

	t.Run("getLocation_bad_location_sign", func(t *testing.T) {
		cg := coordGroups{loc: "Q"}
		_, err := cg.getLocation()
		if err == nil || !errors.Is(err, ErrInvalidCoord) {
			t.Fatalf("Expected ErrInvalidCoord, got %v", err)
		}
	})

	t.Run("getDegrees_missing_degrees", func(t *testing.T) {
		cg := coordGroups{}
		_, err := cg.getDegrees(LocNone)
		if err == nil || !errors.Is(err, ErrInvalidCoord) {
			t.Fatalf("Expected ErrInvalidCoord, got %v", err)
		}
	})

	t.Run("getDegrees_decimal_with_minutes", func(t *testing.T) {
		cg := coordGroups{deg: "48.5", min: "30"}
		_, err := cg.getDegrees(LocNone)
		if err == nil || !errors.Is(err, ErrInvalidCoord) {
			t.Fatalf("Expected ErrInvalidCoord, got %v", err)
		}
	})

	t.Run("getDegrees_bad_degrees", func(t *testing.T) {
		cg := coordGroups{deg: "abc"}
		_, err := cg.getDegrees(LocNone)
		if err == nil || !errors.Is(err, ErrInvalidCoord) {
			t.Fatalf("Expected ErrInvalidCoord, got %v", err)
		}
	})

	t.Run("getMinutes_decimal_with_seconds", func(t *testing.T) {
		cg := coordGroups{min: "33.4", sec: "10"}
		_, err := cg.getMinutes()
		if err == nil || !errors.Is(err, ErrInvalidCoord) {
			t.Fatalf("Expected ErrInvalidCoord, got %v", err)
		}
	})

	t.Run("getMinutes_bad_minutes", func(t *testing.T) {
		cg := coordGroups{min: "abc"}
		_, err := cg.getMinutes()
		if err == nil || !errors.Is(err, ErrInvalidCoord) {
			t.Fatalf("Expected ErrInvalidCoord, got %v", err)
		}
	})

	t.Run("getSeconds_bad_seconds", func(t *testing.T) {
		cg := coordGroups{sec: "abc"}
		_, err := cg.getSeconds()
		if err == nil || !errors.Is(err, ErrInvalidCoord) {
			t.Fatalf("Expected ErrInvalidCoord, got %v", err)
		}
	})

	t.Run("getCoord_propagates_getLocation_error", func(t *testing.T) {
		cg := coordGroups{deg: "48", loc: "Q"}
		_, err := cg.getCoord()
		if err == nil || !errors.Is(err, ErrInvalidCoord) {
			t.Fatalf("Expected ErrInvalidCoord, got %v", err)
		}
	})

	t.Run("getCoord_propagates_getMinutes_error", func(t *testing.T) {
		cg := coordGroups{deg: "48", min: "abc"}
		_, err := cg.getCoord()
		if err == nil || !errors.Is(err, ErrInvalidCoord) {
			t.Fatalf("Expected ErrInvalidCoord, got %v", err)
		}
	})
}

func BenchmarkParseCoord(b *testing.B) {
	benchmarks := []struct {
		name  string
		input string
	}{
		{"DMS", `48°33'26.9604"N`},
		{"MinDec", `48°33.4493'N`},
		{"DegDec", `48.557489`},
		{"Compact", `120-5749E`},
	}
	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_, _ = ParseCoord(bm.input)
			}
		})
	}
}

func BenchmarkParsePoint(b *testing.B) {
	benchmarks := []struct {
		name  string
		input string
	}{
		{"DMS", `48°33'27"N; 120°57'49"E`},
		{"MinDec", `48-33N; 048-33.0E`},
	}
	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_, _ = ParsePoint(bm.input)
			}
		})
	}
}

func BenchmarkCoordFormat(b *testing.B) {
	coord := Coord{Value: 48.557489, Loc: LocLat}
	benchmarks := []struct {
		name    string
		example string
	}{
		{"ToDMS", `48°33'27"N`},
		{"ToMinDec", `48°33.4493'N`},
		{"ToDegDec", `48.557489`},
	}
	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_, _ = coord.Format(bm.example)
			}
		})
	}
}

func BenchmarkCoordString(b *testing.B) {
	benchmarks := []struct {
		name  string
		coord Coord
	}{
		{"Lat", Coord{Value: 48.5575, Loc: LocLat}},
		{"Lon", Coord{Value: 120.963611, Loc: LocLon}},
		{"None", Coord{Value: 48.557489, Loc: LocNone}},
	}
	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = bm.coord.String()
			}
		})
	}
}

func BenchmarkPointString(b *testing.B) {
	p := Point{
		Lat: Coord{Value: 48.5575, Loc: LocLat},
		Lon: Coord{Value: 120.963611, Loc: LocLon},
	}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = p.String()
	}
}
