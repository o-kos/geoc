package geoc

import (
	"unicode"
)

// Match describes a coordinate located inside arbitrary text.
type Match struct {
	Start, End int    // byte offsets in the input string
	Text       string // raw substring from the input, equal to s[Start:End]
	Coord      Coord  // parsed coordinate; Loc is set when N/S/E/W is present
}

// FindOption configures FindCoords behavior.
type FindOption func(*findConfig)

type findConfig struct {
	requireDirection bool
	onlyLoc          Location
}

// RequireDirection makes FindCoords skip matches that lack an N/S/E/W letter.
// Useful for filtering out bare numeric tokens (dates, message numbers,
// distances) that would otherwise look like decimal-degree coordinates.
func RequireDirection() FindOption {
	return func(c *findConfig) { c.requireDirection = true }
}

// OnlyLat keeps only matches whose Coord.Loc is LocLat.
// Implies RequireDirection (latitude is determined by the N/S letter).
func OnlyLat() FindOption {
	return func(c *findConfig) {
		c.requireDirection = true
		c.onlyLoc = LocLat
	}
}

// OnlyLon keeps only matches whose Coord.Loc is LocLon.
// Implies RequireDirection (longitude is determined by the E/W letter).
func OnlyLon() FindOption {
	return func(c *findConfig) {
		c.requireDirection = true
		c.onlyLoc = LocLon
	}
}

// CoordRegexSource returns the regex source string that matches one
// coordinate. It contains no capture groups and no surrounding whitespace,
// so it can be safely embedded into user-defined patterns.
//
// The returned source matches the same set of substrings that FindCoords
// considers as coordinate candidates (parsing may still reject them).
func CoordRegexSource() string {
	return coordRegexCore
}

// FindCoords returns all coordinate occurrences inside s, in order of
// appearance. Candidates that fail to parse are skipped silently.
func FindCoords(s string, opts ...FindOption) []Match {
	cfg := findConfig{}
	for _, opt := range opts {
		opt(&cfg)
	}

	subNames := coordRegExp.SubexpNames()
	allIdx := coordRegExp.FindAllStringSubmatchIndex(s, -1)
	if len(allIdx) == 0 {
		return nil
	}

	out := make([]Match, 0, len(allIdx))
	for _, idx := range allIdx {
		groups := make([]string, len(subNames))
		for i := range subNames {
			gStart, gEnd := idx[2*i], idx[2*i+1]
			if gStart >= 0 {
				groups[i] = s[gStart:gEnd]
			}
		}

		cg, _ := coordGroupsFromMatch(groups, subNames)
		cg.normalizeDotDMS()
		cg.normalizeCompact()

		coord, err := cg.getCoord()
		if err != nil {
			continue
		}

		if cfg.requireDirection && cg.loc == "" {
			continue
		}
		if cfg.onlyLoc != LocNone && coord.Loc != cfg.onlyLoc {
			continue
		}

		start, end := trimMatchSpan(s, idx[0], idx[1])
		if start >= end {
			continue
		}

		out = append(out, Match{
			Start: start,
			End:   end,
			Text:  s[start:end],
			Coord: coord,
		})
	}

	return out
}

// trimMatchSpan returns [start, end] adjusted to exclude leading/trailing
// ASCII whitespace that the regex may have consumed via its outer (\s*)
// groups or via inner separators when no further coord component followed.
func trimMatchSpan(s string, start, end int) (int, int) {
	for start < end && unicode.IsSpace(rune(s[start])) {
		start++
	}
	for end > start && unicode.IsSpace(rune(s[end-1])) {
		end--
	}
	return start, end
}
