package geoc

import (
	"strings"
	"testing"
)

// findCase describes one input string and the expected list of coordinate
// substrings FindCoords should return for it (in order).
type findCase struct {
	name  string
	input string
	opts  []FindOption
	want  []string // expected Match.Text values, in order
}

func runFindCases(t *testing.T, cases []findCase) {
	t.Helper()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := FindCoords(tc.input, tc.opts...)
			if len(got) != len(tc.want) {
				gotTexts := make([]string, len(got))
				for i, m := range got {
					gotTexts[i] = m.Text
				}
				t.Fatalf("got %d matches %q, want %d %q",
					len(got), gotTexts, len(tc.want), tc.want)
			}
			for i, m := range got {
				if m.Text != tc.want[i] {
					t.Errorf("match %d: Text = %q, want %q", i, m.Text, tc.want[i])
				}
				if tc.input[m.Start:m.End] != m.Text {
					t.Errorf("match %d: s[%d:%d]=%q != Text=%q",
						i, m.Start, m.End, tc.input[m.Start:m.End], m.Text)
				}
			}
		})
	}
}

func TestFindCoordsPositive(t *testing.T) {
	runFindCases(t, []findCase{
		{
			name:  "dms_with_decimal_seconds",
			input: "msg 39 27 25.55N body",
			opts:  []FindOption{RequireDirection()},
			want:  []string{"39 27 25.55N"},
		},
		{
			name:  "decimal_comma_greek",
			input: "pos 35 32,00N here",
			opts:  []FindOption{RequireDirection()},
			want:  []string{"35 32,00N"},
		},
		{
			name:  "line_break_inside_coord",
			input: "lat 38-\n34.2N text",
			opts:  []FindOption{RequireDirection()},
			want:  []string{"38-\n34.2N"},
		},
		{
			name:  "dms_sign_separators",
			input: `at 35°32'15"N now`,
			opts:  []FindOption{RequireDirection()},
			want:  []string{`35°32'15"N`},
		},
		{
			name:  "truncated_deg_min",
			input: "x 40 30N y",
			opts:  []FindOption{RequireDirection()},
			want:  []string{"40 30N"},
		},
		{
			name:  "truncated_deg_only",
			input: "x 40N y",
			opts:  []FindOption{RequireDirection()},
			want:  []string{"40N"},
		},
		{
			name:  "truncated_decimal_deg",
			input: "x 40.5N y",
			opts:  []FindOption{RequireDirection()},
			want:  []string{"40.5N"},
		},
		{
			name:  "navtex_pair_with_dash_separator",
			input: "39 27 25.55N - 009 39 25E",
			opts:  []FindOption{RequireDirection()},
			want:  []string{"39 27 25.55N", "009 39 25E"},
		},
		{
			name:  "pair_with_comma_separator",
			input: "lat 39 27N, lon 40 30E end",
			opts:  []FindOption{RequireDirection()},
			want:  []string{"39 27N", "40 30E"},
		},
	})
}

func TestFindCoordsNegativeWithRequireDirection(t *testing.T) {
	// Inputs that look numeric but contain no N/S/E/W must yield zero
	// matches when RequireDirection() is active.
	cases := []findCase{
		{name: "iso_date", input: "2022 03 15", want: nil},
		{name: "navtex_message_number", input: "175/22 MAR 15", want: nil},
		{name: "distance_nm", input: "3 NM", want: nil},
		{name: "distance_metres", input: "1500 METRES", want: nil},
		{name: "zone_id", input: "BR-12", want: nil},
	}
	for i := range cases {
		cases[i].opts = []FindOption{RequireDirection()}
	}
	runFindCases(t, cases)
}

func TestFindCoordsBoundary(t *testing.T) {
	t.Run("position_excludes_surrounding_whitespace", func(t *testing.T) {
		input := "   39 27N   "
		got := FindCoords(input, RequireDirection())
		if len(got) != 1 {
			t.Fatalf("got %d matches, want 1", len(got))
		}
		m := got[0]
		if m.Text != "39 27N" {
			t.Errorf("Text = %q, want %q", m.Text, "39 27N")
		}
		if input[m.Start:m.End] != m.Text {
			t.Errorf("s[%d:%d]=%q != Text=%q", m.Start, m.End,
				input[m.Start:m.End], m.Text)
		}
		if m.Start != 3 || m.End != 9 {
			t.Errorf("Start/End = %d/%d, want 3/9", m.Start, m.End)
		}
	})

	t.Run("substring_invariant_on_complex_text", func(t *testing.T) {
		// Several coords mixed with noise; every returned match must
		// satisfy s[Start:End] == Text.
		input := "  alpha 39 27 25.55N - 009 39 25E beta 40N gamma  "
		got := FindCoords(input, RequireDirection())
		if len(got) == 0 {
			t.Fatal("no matches")
		}
		for i, m := range got {
			if input[m.Start:m.End] != m.Text {
				t.Errorf("match %d: s[%d:%d]=%q != Text=%q",
					i, m.Start, m.End, input[m.Start:m.End], m.Text)
			}
			if strings.HasPrefix(m.Text, " ") || strings.HasSuffix(m.Text, " ") {
				t.Errorf("match %d: Text %q has surrounding whitespace",
					i, m.Text)
			}
		}
	})

	t.Run("adjacent_coords_no_whitespace_separator", func(t *testing.T) {
		input := "39N40E"
		got := FindCoords(input, RequireDirection())
		want := []string{"39N", "40E"}
		if len(got) != len(want) {
			gotTexts := make([]string, len(got))
			for i, m := range got {
				gotTexts[i] = m.Text
			}
			t.Fatalf("got %d matches %q, want %d %q",
				len(got), gotTexts, len(want), want)
		}
		for i, m := range got {
			if m.Text != want[i] {
				t.Errorf("match %d: Text = %q, want %q", i, m.Text, want[i])
			}
			if input[m.Start:m.End] != m.Text {
				t.Errorf("match %d: s[Start:End] mismatch", i)
			}
		}
	})
}

func TestFindCoordsOnlyLatOnlyLon(t *testing.T) {
	input := "lat 39 27N body 40 30E end"

	t.Run("only_lat", func(t *testing.T) {
		got := FindCoords(input, OnlyLat())
		if len(got) != 1 {
			t.Fatalf("got %d matches, want 1", len(got))
		}
		if got[0].Text != "39 27N" {
			t.Errorf("Text = %q, want %q", got[0].Text, "39 27N")
		}
		if got[0].Coord.Loc != LocLat {
			t.Errorf("Loc = %s, want Lat", got[0].Coord.Loc)
		}
	})

	t.Run("only_lon", func(t *testing.T) {
		got := FindCoords(input, OnlyLon())
		if len(got) != 1 {
			t.Fatalf("got %d matches, want 1", len(got))
		}
		if got[0].Text != "40 30E" {
			t.Errorf("Text = %q, want %q", got[0].Text, "40 30E")
		}
		if got[0].Coord.Loc != LocLon {
			t.Errorf("Loc = %s, want Lon", got[0].Coord.Loc)
		}
	})
}

func TestFindCoordsUnitSuffixRejected(t *testing.T) {
	// Direction letters glued to a unit abbreviation ("1 N.M.", "5 S.M.")
	// must not be accepted as coordinates — the trailing dot makes the
	// naive one-byte letter-glue check pass, so the regex needs a proper
	// look-ahead for known distance units.
	cases := []findCase{
		{name: "nautical_mile_dotted", input: "1 N.M.", want: nil},
		{name: "statute_mile_dotted", input: "5 S.M.", want: nil},
		{name: "nautical_mile_radius", input: "RADIUS 1 N.M.", want: nil},
		{name: "nautical_mile_within", input: "WITHIN 10 N.M. OF", want: nil},
		// Counter-cases: real coords must keep matching.
		{
			name:  "real_coord_pair",
			input: "1N 2E",
			want:  []string{"1N", "2E"},
		},
		{
			name:  "direction_with_period_then_word",
			input: "1N. AREA",
			want:  []string{"1N"},
		},
	}
	for i := range cases {
		cases[i].opts = []FindOption{RequireDirection()}
	}
	runFindCases(t, cases)
}

func TestFindCoordsNoCrossLineGreedy(t *testing.T) {
	// A noise number followed by a real coordinate on the next line must
	// not collapse into a single oversized candidate that swallows the
	// real coord ("BR-117\n55-54.0N" used to lose the 55-54.0N match
	// because deg=117 ate the linebreak).
	cases := []findCase{
		{
			name:  "navtex_zone_then_pair",
			input: "BR-117\n55-54.0N 019-03.0E",
			want:  []string{"55-54.0N", "019-03.0E"},
		},
		{
			name:  "header_line_then_pair",
			input: "FROM 192200 UTC\n34-35.6N 139-49.2E",
			want:  []string{"34-35.6N", "139-49.2E"},
		},
		// Counter-case: a hyphen-bridged linebreak inside one coord must
		// stay as a single match.
		{
			name:  "intra_coord_linebreak_after_dash",
			input: "lat 38-\n34.2N text",
			want:  []string{"38-\n34.2N"},
		},
	}
	for i := range cases {
		cases[i].opts = []FindOption{RequireDirection()}
	}
	runFindCases(t, cases)
}

func TestFindCoordsGluedToTerminator(t *testing.T) {
	// A fully-specified coord glued to a non-coord word (NAVTEX terminator
	// "NNN", message "END", etc.) must still be returned. The trailing
	// letters can't themselves form a coordinate, so the strict
	// letter-glue check is too aggressive in this shape.
	cases := []findCase{
		{
			name:  "navtex_terminator_NNN",
			input: "124-50-00ENNN",
			want:  []string{"124-50-00E"},
		},
		// Counter-case: a minimal coord "3 N" glued to "M" (3 NM = 3
		// nautical miles) must keep being rejected — the candidate has
		// no min/sec to anchor it as a real coord.
		{
			name:  "distance_nm_minimal",
			input: "3 NM",
			want:  nil,
		},
	}
	for i := range cases {
		cases[i].opts = []FindOption{RequireDirection()}
	}
	runFindCases(t, cases)
}

func TestCoordRegexSourceEmbeddable(t *testing.T) {
	// CoordRegexSource() output must be safe to embed inside a user-defined
	// named group: it must not contain capture-group syntax of its own.
	src := CoordRegexSource()
	if src == "" {
		t.Fatal("CoordRegexSource returned empty string")
	}
	if strings.Contains(src, "(?P<") {
		t.Errorf("source contains named capture group: %q", src)
	}
	// Plain ( that is not part of (?: or (?= etc indicates a capturing group.
	for i := 0; i < len(src)-1; i++ {
		if src[i] == '(' && src[i+1] != '?' {
			t.Errorf("source contains capturing group at offset %d: %q", i, src)
			break
		}
	}
}
