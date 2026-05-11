package geoc

import "regexp"

// unitSuffixRE matches a distance-unit abbreviation glued to a direction
// letter. When this matches the bytes immediately after the loc letter, the
// candidate is a distance fragment ("1 N.M.", "5 S.M.", "10 N MILE", "5 KM"),
// not a coordinate.
var unitSuffixRE = regexp.MustCompile(`(?i)^\.?[ \t]*(?:M\.|MILE\b|MI\b|KM\b)`)

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

// locGroupIndex is the position of the named "loc" group in
// coordRegExp.SubexpNames(). Cached at init for quick access in FindCoords.
var locGroupIndex = func() int {
	for i, name := range coordRegExp.SubexpNames() {
		if name == "loc" {
			return i
		}
	}
	return -1
}()

// FindCoords returns all coordinate occurrences inside s, in order of
// appearance. Candidates that fail to parse are skipped silently.
//
// FindCoords runs a manual scan loop rather than FindAllStringSubmatchIndex
// so that when a candidate is dropped (out-of-range degrees, glued unit
// suffix, RequireDirection mismatch) the search can recover a real coord
// hidden inside the dropped span: see TestFindCoordsNoCrossLineGreedy.
func FindCoords(s string, opts ...FindOption) []Match {
	cfg := findConfig{}
	for _, opt := range opts {
		opt(&cfg)
	}

	subNames := coordRegExp.SubexpNames()
	var out []Match
	pos := 0

	for pos < len(s) {
		rel := coordRegExp.FindStringSubmatchIndex(s[pos:])
		if rel == nil {
			break
		}
		idx := make([]int, len(rel))
		for i, v := range rel {
			if v < 0 {
				idx[i] = v
			} else {
				idx[i] = v + pos
			}
		}

		groups := make([]string, len(subNames))
		for i := range subNames {
			gStart, gEnd := idx[2*i], idx[2*i+1]
			if gStart >= 0 {
				groups[i] = s[gStart:gEnd]
			}
		}

		cg, _ := coordGroupsFromMatch(groups, subNames)
		cg.normalizeItalianMinFrac()
		cg.normalizeDotDMS()
		cg.normalizeCompact()

		if !acceptLetterGlue(s, idx[2*locGroupIndex+1], &cg) {
			pos = idx[0] + 1
			continue
		}

		coord, err := cg.getCoord()
		if err != nil {
			if isLatLonHyphenSeparator(s, idx[0]) {
				pos = idx[0] + 1
				continue
			}
			// Backtrack past the first whitespace inside the consumed
			// candidate so a noise prefix glued to a coord by a space
			// ("235 43-18.0N") doesn't trap us on a still-valid sub-
			// candidate ("35 43-18.0N"). Same shape as the cross-line
			// "BR-117\n55-54.0N" recovery, generalized.
			if next := firstSpacePast(s, idx[0], idx[1]); next > 0 {
				pos = next
			} else {
				pos = idx[0] + 1
			}
			continue
		}

		if cfg.requireDirection && cg.loc == "" {
			pos = idx[0] + 1
			continue
		}
		if cfg.onlyLoc != LocNone && coord.Loc != cfg.onlyLoc {
			pos = idx[1]
			continue
		}

		start, end := trimMatchSpan(s, idx[0], idx[1])
		if start >= end {
			pos = idx[1]
			continue
		}

		out = append(out, Match{
			Start: start,
			End:   end,
			Text:  s[start:end],
			Coord: coord,
		})
		pos = idx[1]
	}

	return out
}

// acceptLetterGlue reports whether a candidate's loc letter is followed by
// bytes that should make us drop it. Returns true to keep the candidate.
//
// The check is in two parts:
//
//   - Unit suffix (".M.", " MILE", "KM" etc.): always drops, since the
//     direction letter is part of a distance token, not a coordinate.
//   - Plain letter glue (e.g. "1NORTH", "3 NM"): drops minimal candidates
//     (no min/sec) where the trailing word could be anything, but accepts
//     fully-specified DMS candidates whose trailing letters can't form a
//     coord on their own (NAVTEX terminator "NNN", message words like
//     "END"/"AREA"). See TestFindCoordsGluedToTerminator.
func acceptLetterGlue(s string, locEnd int, cg *coordGroups) bool {
	if cg.loc == "" || locEnd < 0 || locEnd >= len(s) {
		return true
	}
	rest := s[locEnd:]
	if unitSuffixRE.MatchString(rest) {
		return false
	}
	if !isASCIILetter(rest[0]) {
		return true
	}
	if isMonthGlue(s, locEnd) {
		return false
	}
	// OCR-glue: a single non-direction letter wedged between two adjacent
	// coordinates ("46-20.0NR142-2.0E"). The next regex pass will pick up
	// the second coord; here we just accept the current one.
	if !isDirectionLetter(rest[0]) && isOCRGluePrefix(rest) {
		return true
	}
	// Trailing letter glue. Keep the candidate only if it has full DMS
	// structure (deg+min+sec) AND the trailing alphanumeric run contains
	// no digits — i.e. it's a word, not another glued coord.
	if cg.sec == "" {
		return false
	}
	for i := 0; i < len(rest); i++ {
		c := rest[i]
		if c >= '0' && c <= '9' {
			return false
		}
		if !isASCIILetter(c) {
			break
		}
	}
	return true
}

func isMonthGlue(s string, locEnd int) bool {
	if locEnd <= 0 {
		return false
	}
	start := locEnd - 1
	end := locEnd
	for end < len(s) && isASCIILetter(s[end]) {
		end++
	}
	if end-start != 3 {
		return false
	}
	month := upperASCII(s[start:end])
	switch month {
	case "JAN", "FEB", "MAR", "APR", "MAY", "JUN", "JUL", "AUG", "SEP", "OCT", "NOV", "DEC":
		return true
	}
	return false
}

func upperASCII(s string) string {
	var b [3]byte
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'a' && c <= 'z' {
			c -= 'a' - 'A'
		}
		b[i] = c
	}
	return string(b[:len(s)])
}

func isASCIILetter(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z')
}

func isDirectionLetter(b byte) bool {
	switch b {
	case 'N', 'S', 'E', 'W', 'n', 's', 'e', 'w':
		return true
	}
	return false
}

// isOCRGluePrefix reports whether rest looks like an OCR-induced single
// letter wedged between two coordinates: one ASCII letter, optionally
// followed by horizontal whitespace, then an ASCII digit. Two consecutive
// letters disqualify it (that's a word, not a glue artifact).
func isOCRGluePrefix(rest string) bool {
	if len(rest) < 2 || !isASCIILetter(rest[0]) {
		return false
	}
	if isASCIILetter(rest[1]) {
		return false
	}
	i := 1
	for i < len(rest) && (rest[i] == ' ' || rest[i] == '\t') {
		i++
	}
	return i < len(rest) && rest[i] >= '0' && rest[i] <= '9'
}

func isLatLonHyphenSeparator(s string, hyphen int) bool {
	if hyphen <= 0 || hyphen+1 >= len(s) || s[hyphen] != '-' {
		return false
	}
	if s[hyphen+1] < '0' || s[hyphen+1] > '9' {
		return false
	}
	prev := hyphen - 1
	for prev >= 0 && (s[prev] == ' ' || s[prev] == '\t') {
		prev--
	}
	return prev >= 0 && (s[prev] == 'N' || s[prev] == 'S')
}

// firstSpacePast returns the byte position right after the first ASCII
// whitespace inside s[start:end], or -1 if none. Used to recover from
// candidates whose numeric prefix made the whole span fail range checks
// ("235 43-18.0N" → restart at "43-18.0N", not "35 43-18.0N").
func firstSpacePast(s string, start, end int) int {
	for i := start; i < end; i++ {
		if isASCIISpace(s[i]) {
			return i + 1
		}
	}
	return -1
}

// trimMatchSpan returns [start, end] adjusted to exclude ASCII whitespace
// that the regex may have consumed via its outer (\s*) groups or via inner
// separators when no further coord component followed. Matches RE2's \s
// class: space, tab, newline, carriage return, form feed, vertical tab.
func trimMatchSpan(s string, start, end int) (int, int) {
	for start < end && isASCIISpace(s[start]) {
		start++
	}
	for end > start && isASCIISpace(s[end-1]) {
		end--
	}
	return start, end
}

func isASCIISpace(b byte) bool {
	switch b {
	case ' ', '\t', '\n', '\r', '\f', '\v':
		return true
	}
	return false
}
