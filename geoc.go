// Package geoc parses and formats geographic coordinates and points.
// It supports conversion between string representations and native values.
package geoc

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

type Location int

const (
	LocNone Location = iota // LocNone means location type is not specified.
	LocLat                  // LocLat means coordinate is latitude.
	LocLon                  // LocLon means coordinate is longitude.
)

func (l Location) String() string {
	switch l {
	case LocLat:
		return "Lat"
	case LocLon:
		return "Lon"
	default:
		return "None"
	}
}

// Coord represents a geographic coordinate with its location type.
type Coord struct {
	Value float64
	Loc   Location
}

// Point represents a geographic point with latitude and longitude.
type Point struct {
	Lat Coord
	Lon Coord
}

// ParseError represents a coordinate parsing or formatting error.
type ParseError string

func (e ParseError) Error() string { return string(e) }

const (
	// ErrInvalidString indicates that input string cannot be parsed.
	ErrInvalidString = ParseError("unable to parse coordinates string")
	// ErrInvalidCoord indicates that coordinate components are inconsistent.
	ErrInvalidCoord = ParseError("invalid coordinate")
	// ErrOutOfRange indicates that coordinate value is outside allowed limits.
	ErrOutOfRange = ParseError("out of range")
)

// ParseCoord parses a coordinate string and returns a Coord.
// Location type (Lat/Lon) is auto-detected from the location letter (N/S/E/W).
func ParseCoord(s string) (Coord, error) {
	coord := Coord{}
	cg, err := newCoordGroups(s)
	if err != nil {
		return coord, fmt.Errorf("%w in string %q", err, s)
	}

	coord, err = cg.getCoord()
	if err != nil {
		return coord, fmt.Errorf("%w in string %q", err, s)
	}

	return coord, nil
}

// Format converts coordinate to string representation
// using provided example format string.
func (c Coord) Format(example string) (string, error) {
	cg, err := newCoordGroups(example)
	if err != nil {
		return "", fmt.Errorf("%w: invalid example format", ErrInvalidString)
	}

	// Use Coord's Loc if set, otherwise derive from example
	loc := c.Loc
	if loc == LocNone {
		switch cg.loc {
		case "N", "S":
			loc = LocLat
		case "E", "W":
			loc = LocLon
		}
	}

	// Validate coord bounds
	absCoord := math.Abs(c.Value)
	if loc == LocLat && absCoord > 90 {
		return "", fmt.Errorf("%w: latitude %f", ErrOutOfRange, c.Value)
	}
	if loc == LocLon && absCoord > 180 {
		return "", fmt.Errorf("%w: longitude %f", ErrOutOfRange, c.Value)
	}
	if loc == LocNone && absCoord > 180 {
		return "", fmt.Errorf("%w: coordinate %f", ErrOutOfRange, c.Value)
	}

	negative := c.Value < 0

	hasSec := cg.sec != ""
	hasMin := cg.min != ""

	// Detect decimal separator and precision from the most specific component
	decSep := "."
	precision := 0
	detectDecimal := func(s string) {
		if idx := strings.IndexAny(s, ".,"); idx != -1 {
			decSep = string(s[idx])
			precision = len(s) - idx - 1
		}
	}
	switch {
	case hasSec && !cg.compact:
		detectDecimal(cg.sec)
	case hasMin && !cg.compact:
		detectDecimal(cg.min)
	case !hasMin:
		detectDecimal(cg.deg)
	}

	// Detect component widths for zero-padding.
	componentIntWidth := func(s string) int {
		if idx := strings.IndexAny(s, ".,"); idx != -1 {
			return idx
		}
		return len(s)
	}
	formatFixed := func(value float64, intWidth, fracPrecision int) string {
		var out string
		if fracPrecision > 0 {
			totalWidth := intWidth + 1 + fracPrecision
			out = fmt.Sprintf("%0*.*f", totalWidth, fracPrecision, value)
		} else {
			out = fmt.Sprintf("%0*.0f", intWidth, value)
		}
		if decSep == "," {
			out = strings.Replace(out, ".", ",", 1)
		}
		return out
	}
	roundWithPrecision := func(value float64, fracPrecision int) float64 {
		scale := math.Pow(10, float64(fracPrecision))
		return math.Round(value*scale) / scale
	}

	degWidth := componentIntWidth(cg.deg)
	minWidth := componentIntWidth(cg.min)
	secWidth := componentIntWidth(cg.sec)

	// Determine output location letter
	locLetter := ""
	if cg.loc != "" {
		switch {
		case loc == LocLat && negative:
			locLetter = "S"
		case loc == LocLat:
			locLetter = "N"
		case negative:
			locLetter = "W"
		default:
			locLetter = "E"
		}
	}
	applySign := func(body string) string {
		if cg.loc != "" {
			return body
		}
		if negative {
			return "-" + body
		}
		if cg.sgn == "+" {
			return "+" + body
		}
		return body
	}

	// DegDec format
	if !hasMin {
		degStr := formatFixed(absCoord, degWidth, precision)
		return applySign(degStr) + cg.sep.deg + locLetter, nil
	}

	deg := math.Floor(absCoord)

	// MinDec format
	if !hasSec {
		minutes := roundWithPrecision((absCoord-deg)*60, precision)
		if minutes >= 60 {
			minutes = 0
			deg++
		}
		degStr := fmt.Sprintf("%0*.0f", degWidth, deg)
		minStr := formatFixed(minutes, minWidth, precision)
		return applySign(degStr+cg.sep.deg+minStr+cg.sep.min) + locLetter, nil
	}

	// DMS format
	totalSec := roundWithPrecision((absCoord-deg)*3600, precision)
	if totalSec >= 3600 {
		totalSec = 0
		deg++
	}
	minutes := math.Floor(totalSec / 60)
	sec := totalSec - minutes*60
	degStr := fmt.Sprintf("%0*.0f", degWidth, deg)

	if cg.compact {
		minStr := formatFixed(minutes, minWidth, 0)
		secStr := formatFixed(sec, secWidth, 0)
		return applySign(degStr+cg.sep.deg+minStr+secStr) + locLetter, nil
	}

	minStr := formatFixed(minutes, minWidth, 0)
	secStr := formatFixed(sec, secWidth, precision)
	return applySign(degStr+cg.sep.deg+minStr+cg.sep.min+secStr+cg.sep.sec) + locLetter, nil
}

func formatMinDec(value float64, degWidth int, pos, neg byte) string {
	absVal := math.Abs(value)
	deg := math.Floor(absVal)
	minutes := math.Round((absVal-deg)*60*10) / 10
	if minutes >= 60 {
		minutes = 0
		deg++
	}
	letter := pos
	if value < 0 {
		letter = neg
	}
	return fmt.Sprintf("%0*.0f-%04.1f%c", degWidth, deg, minutes, letter)
}

// String returns default string representation of the coordinate.
// Latitude uses MinDec format (48-33.0N), longitude uses MinDec
// with 3-digit degrees (048-33.0E), unspecified uses decimal degrees.
func (c Coord) String() string {
	switch c.Loc {
	case LocLat:
		return formatMinDec(c.Value, 2, 'N', 'S')
	case LocLon:
		return formatMinDec(c.Value, 3, 'E', 'W')
	default:
		return strconv.FormatFloat(c.Value, 'f', -1, 64)
	}
}

// ParsePoint parses a string containing latitude and longitude.
// Latitude is parsed from the beginning of the string; longitude is then
// searched to the right starting from the next symbol after latitude.
// Latitude and longitude must use compatible format classes (degDec/minDec/dms).
func ParsePoint(s string) (Point, error) {
	p := Point{}
	cgLat, cgLon, err := newPointGroups(s)
	if err != nil {
		return p, fmt.Errorf("%w in string %q", err, s)
	}

	lat, err := cgLat.getCoord()
	if err != nil {
		return p, fmt.Errorf("%w in string %q", err, s)
	}

	if lat.Loc != LocLat {
		return p, fmt.Errorf("%w: bad latitude location in string %q", ErrInvalidString, s)
	}

	lon, err := cgLon.getCoord()
	if err != nil {
		return p, fmt.Errorf("%w in string %q", err, s)
	}

	if lon.Loc != LocLon {
		return p, fmt.Errorf("%w: bad longitude location in string %q", ErrInvalidString, s)
	}

	if cgLat.getFormatClass() != cgLon.getFormatClass() {
		return p, fmt.Errorf("%w: incompatible lat/lon formats in string %q", ErrInvalidString, s)
	}

	return Point{lat, lon}, nil
}

// Format converts Point to string representation using provided format
// examples for latitude and longitude coordinates and joins them with separator.
func (p Point) Format(latFmt, lonFmt, separator string) (string, error) {
	lat, err := p.Lat.Format(latFmt)
	if err != nil {
		return "", fmt.Errorf("%w: latitude format %q", err, latFmt)
	}
	lon, err := p.Lon.Format(lonFmt)
	if err != nil {
		return "", fmt.Errorf("%w: longitude format %q", err, lonFmt)
	}
	return lat + separator + lon, nil
}

// String returns default string representation of the point.
// Default format is "48-33.0N 048-33.0E".
func (p Point) String() string {
	return formatMinDec(p.Lat.Value, 2, 'N', 'S') + " " + formatMinDec(p.Lon.Value, 3, 'E', 'W')
}

type coordGroups struct {
	sgn     string
	deg     string
	min     string
	sec     string
	loc     string
	compact bool
	sep     struct {
		deg string
		min string
		sec string
	}
}

// compactMMSSRegExp matches a 4-digit no-decimal field used by
// normalizeCompact to split MMSS minutes into separate min/sec fields.
var compactMMSSRegExp = regexp.MustCompile(`^(\d{2})(\d{2})$`)

// normalizeCompact splits compact MMSS minutes (e.g., "5749")
// into separate min ("57") and sec ("49") fields.
func (cg *coordGroups) normalizeCompact() {
	if cg.sec != "" || cg.loc == "" {
		return
	}
	m := compactMMSSRegExp.FindStringSubmatch(cg.min)
	if m == nil {
		return
	}
	cg.compact = true
	cg.min = m[1]
	cg.sec = m[2]
}

// italianMinFracRegExp matches a 3-digit no-decimal trailer used by
// normalizeItalianMinFrac as the fractional-part fingerprint.
var italianMinFracRegExp = regexp.MustCompile(`^\d{3}$`)

// normalizeItalianMinFrac handles space-separated forms where the fractional
// part of minutes is written as 3 digits without a decimal point — common in
// italian NAVTEX OCR output ("44 33 367N" meaning "44° 33.367' N"). The shape
// is `<deg> <min> <minFrac3>[NSEW]`, with both separators being plain
// whitespace and the would-be sec field consisting of exactly 3 ASCII digits.
// Classical DMS uses "-"/"°"/":" or has a decimal in the sec field, so a
// pure-space separator with a 3-digit no-decimal trailer is a reliable
// fingerprint. Without this normalization the candidate is interpreted as
// DMS and then either rejected (sec ≥ 60, e.g. "367") or — worse — silently
// produces wrong coordinates (sec < 60, e.g. "048" parsed as 48 seconds).
func (cg *coordGroups) normalizeItalianMinFrac() {
	if cg.loc == "" || cg.min == "" {
		return
	}
	if cg.sep.deg != " " || cg.sep.min != " " {
		return
	}
	if !italianMinFracRegExp.MatchString(cg.sec) {
		return
	}
	if strings.ContainsAny(cg.min, ".,") {
		return
	}
	cg.min = cg.min + "." + cg.sec
	cg.sec = ""
	cg.sep.min = ""
}

// compactMinDecLatRegExp / compactMinDecLonRegExp match concatenated MinDec
// integer parts: 2-digit deg + 2-digit min for N/S, 3-digit deg + 2-digit min
// for E/W. The deg group is range-checked (≤90 / ≤180) and the min group is
// constrained to `[0-5]\d` so the trailing minutes stay < 60.
var (
	compactMinDecLatRegExp = regexp.MustCompile(`^(\d{2})([0-5]\d)\.(\d+)$`)
	compactMinDecLonRegExp = regexp.MustCompile(`^(\d{3})([0-5]\d)\.(\d+)$`)
)

// normalizeCompactMinDec handles compact MinDec inputs where degrees and
// decimal minutes are written as a single concatenated number with no
// separator, e.g. "3630.055N" meaning 36° 30.055' N, or "01202.598E" meaning
// 12° 02.598' E. This is the MinDec analogue of the compact DMS form
// (`DDMMSS`) that normalizeCompact already handles. The regex parses such
// inputs as a single deg group with a decimal fraction and an empty min/sec,
// which the degDec branch then rejects as out-of-range (>90 / >180).
//
// Trigger is intentionally narrow:
//   - direction letter is present;
//   - no deg-min separator was consumed (degDec parse path);
//   - both min and sec are empty;
//   - deg matches `^\d{4,5}\.\d+$` — 4 integer digits for N/S or 5 for E/W;
//   - splitting deg into a leading deg part (2 digits for N/S, 3 for E/W)
//     and a trailing `\d{2}\.\d+` minutes part keeps the deg part in axis
//     bounds (≤90 lat, ≤180 lon) and the minutes part < 60.
func (cg *coordGroups) normalizeCompactMinDec() {
	if cg.loc == "" || cg.min != "" || cg.sec != "" || cg.sep.deg != "" {
		return
	}
	var re *regexp.Regexp
	var degMax int
	switch cg.loc {
	case "N", "S":
		re, degMax = compactMinDecLatRegExp, 90
	case "E", "W":
		re, degMax = compactMinDecLonRegExp, 180
	default:
		return
	}
	m := re.FindStringSubmatch(cg.deg)
	if m == nil {
		return
	}
	degVal, err := strconv.Atoi(m[1])
	if err != nil || degVal > degMax {
		return
	}
	cg.deg = m[1]
	cg.min = m[2] + "." + m[3]
}

// dotDMSDegRegExp matches a `deg.min` integer split used by normalizeDotDMS
// to reshape inputs like "70.19.4N" into regular DMS groups.
var dotDMSDegRegExp = regexp.MustCompile(`^(\d+)\.(\d+)$`)

// normalizeDotDMS handles forms like "70.19.4N" and "018.07.5E".
// Regex can initially parse them as deg=70.19, min=4. This method
// converts such shape into regular DMS groups: deg=70, min=19, sec=4.
func (cg *coordGroups) normalizeDotDMS() {
	if cg.loc == "" || cg.min == "" || cg.sec != "" || cg.sep.deg != "." {
		return
	}
	if strings.ContainsAny(cg.min, ".,") {
		return
	}
	m := dotDMSDegRegExp.FindStringSubmatch(cg.deg)
	if m == nil {
		return
	}
	cg.deg = m[1]
	cg.sec = cg.min
	cg.min = m[2]
	if cg.sep.min == "" {
		cg.sep.min = "."
	}
}

type formatClass int

const (
	degDec formatClass = iota
	minDec
	dms
)

func (cg *coordGroups) getFormatClass() formatClass {
	if cg.min == "" {
		return degDec
	}
	if cg.sec == "" {
		return minDec
	}
	return dms
}

// Shared building blocks for the coordinate regex. coordRegexCore (the
// non-capturing source returned by CoordRegexSource) and coordRegExp (the
// named-group version used by ParseCoord/ParsePoint) reuse these so the two
// stay structurally in sync.
const (
	rxNum = `\d+(?:[\.,]\d+)?`
	// rxDegSep allows a newline only when bridged by an explicit "-"/"°"/"."
	// separator (intra-coord linebreak like "38-\n34.2N"). Without such a
	// separator, only horizontal whitespace is allowed — otherwise FindCoords
	// would greedily eat "\n" between an unrelated number and the next line's
	// real coordinate (e.g. "BR-117\n55-54.0N").
	rxDegSep = `[^\S\n]*(?:[-°\.][^\S\n]*\n?[^\S\n]*)?`
	rxMinSep = `\s*[-'\.]?\s*`
	rxSecSep = `\s*[ "]?\s*`
)

// coordRegexCore matches a single coordinate without capture groups and
// without surrounding whitespace. Suitable for embedding into user patterns.
const coordRegexCore = `[-+]?` +
	`(?:` + rxNum + `(?:` + rxDegSep + `)?)` +
	`(?:` + rxNum + `(?:` + rxMinSep + `)?)?` +
	`(?:` + rxNum + `(?:` + rxSecSep + `)?)?` +
	`[NSEW]?`

var coordRegExp = regexp.MustCompile(
	`(\s*)` +
		`(?P<sgn>[-+])?` +
		`(?:(?P<deg>` + rxNum + `)(?P<dsr>` + rxDegSep + `)?)` +
		`(?:(?P<min>` + rxNum + `)(?P<msr>` + rxMinSep + `)?)?` +
		`(?:(?P<sec>` + rxNum + `)(?P<ssr>` + rxSecSep + `)?)?` +
		`(?P<loc>[NSEW])?(\s*)`,
)

func coordGroupsFromMatch(matches []string, subNames []string) (coordGroups, int) {
	makeSep := func(sep string) string {
		if ret := strings.TrimSpace(sep); ret != "" {
			return ret
		}
		if sep != "" {
			return " "
		}
		return ""
	}

	cg := coordGroups{}
	totalLen := 0
	for i, name := range subNames {
		value := matches[i]
		if i != 0 && value != "" {
			switch name {
			case "sgn":
				cg.sgn = value
			case "deg":
				cg.deg = value
			case "min":
				cg.min = value
			case "sec":
				cg.sec = value
			case "loc":
				cg.loc = value
			case "dsr":
				cg.sep.deg = makeSep(value)
			case "msr":
				cg.sep.min = makeSep(value)
			case "ssr":
				cg.sep.sec = makeSep(value)
			}
			totalLen += len(value)
		}
	}

	return cg, totalLen
}

func newCoordGroups(cs string) (coordGroups, error) {
	cg := coordGroups{}
	// Request up to 2 matches to detect "too many coords" case
	m := coordRegExp.FindAllStringSubmatch(cs, 2)
	if len(m) == 0 {
		return cg, fmt.Errorf("%w: coords not found", ErrInvalidString)
	}
	if len(m) > 1 {
		return cg, fmt.Errorf("%w: too many coords found", ErrInvalidString)
	}

	cg, totalLen := coordGroupsFromMatch(m[0], coordRegExp.SubexpNames())
	if totalLen != len(cs) {
		return cg, fmt.Errorf("%w: extra characters detected", ErrInvalidString)
	}

	cg.normalizeItalianMinFrac()
	cg.normalizeDotDMS()
	cg.normalizeCompactMinDec()
	cg.normalizeCompact()
	return cg, nil
}

func newPointGroups(cs string) (coordGroups, coordGroups, error) {
	cgLat := coordGroups{}
	cgLon := coordGroups{}
	// Request up to 3 matches to detect "too many coords" case
	matchIdx := coordRegExp.FindAllStringSubmatchIndex(cs, 3)
	if len(matchIdx) == 0 {
		return cgLat, cgLon, fmt.Errorf("%w: coords not found", ErrInvalidString)
	}
	if len(matchIdx) == 1 {
		return cgLat, cgLon, fmt.Errorf("%w: too few coords found", ErrInvalidString)
	}
	if len(matchIdx) > 2 {
		return cgLat, cgLon, fmt.Errorf("%w: too many coords found", ErrInvalidString)
	}
	if matchIdx[0][0] != 0 || matchIdx[1][1] != len(cs) {
		return cgLat, cgLon, fmt.Errorf("%w: extra characters detected", ErrInvalidString)
	}
	if sep := cs[matchIdx[0][1]:matchIdx[1][0]]; !isPointSeparator(sep) {
		return cgLat, cgLon, fmt.Errorf("%w: invalid separator", ErrInvalidString)
	}

	m := coordRegExp.FindAllStringSubmatch(cs, 2)

	sen := coordRegExp.SubexpNames()
	cgLat, _ = coordGroupsFromMatch(m[0], sen)
	cgLon, _ = coordGroupsFromMatch(m[1], sen)

	// If coords are glued with '-' (e.g. "70.19.4N-018.07.5E"),
	// regex treats '-' as sign of second coord. When second coord has
	// location letter, sign is invalid and '-' should be interpreted
	// as point separator.
	if matchIdx[0][1] == matchIdx[1][0] && cgLon.sgn == "-" && cgLon.loc != "" {
		cgLon.sgn = ""
	}

	cgLat.normalizeItalianMinFrac()
	cgLon.normalizeItalianMinFrac()
	cgLat.normalizeDotDMS()
	cgLon.normalizeDotDMS()
	cgLat.normalizeCompactMinDec()
	cgLon.normalizeCompactMinDec()
	cgLat.normalizeCompact()
	cgLon.normalizeCompact()
	return cgLat, cgLon, nil
}

func isPointSeparator(s string) bool {
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

func (cg *coordGroups) getLocation() (Location, error) {
	switch cg.loc {
	case "N", "S":
		return LocLat, nil
	case "E", "W":
		return LocLon, nil
	case "":
		return LocNone, nil
	}

	return LocNone, fmt.Errorf("%w: bad location sign %q", ErrInvalidCoord, cg.loc)
}

func (cg *coordGroups) checkSign() error {
	if cg.sgn == "" {
		return nil
	}
	if (cg.sgn == "+" || cg.sgn == "-") && cg.loc != "" {
		return fmt.Errorf("%w: sign & location symbols conflict", ErrInvalidCoord)
	}
	return nil
}

func checkLimits(value float64, limit float64, kind string) (float64, error) {
	if value < limit {
		return value, nil
	}
	return 0, fmt.Errorf("%w: %s", ErrOutOfRange, kind)
}

func (cg *coordGroups) getDegrees(loc Location) (float64, error) {
	if cg.deg == "" {
		return 0, fmt.Errorf("%w: missing degrees", ErrInvalidCoord)
	}

	// Check float degrees & exists minutes/seconds
	idx := strings.IndexAny(cg.deg, ".,")
	if idx != -1 && (cg.min != "" || cg.sec != "") {
		return 0, fmt.Errorf("%w: degrees with decimal and minutes", ErrInvalidCoord)
	}
	degStr := cg.deg
	if idx != -1 {
		degStr = cg.deg[:idx] + "." + cg.deg[idx+1:]
	}
	if degrees, err := strconv.ParseFloat(degStr, 64); err == nil {
		limit := 180.0
		if cg.loc == "S" || cg.loc == "N" || loc == LocLat {
			limit = 90.0
		}
		if degrees > limit {
			return 0, fmt.Errorf("%w: degrees", ErrOutOfRange)
		}
		return degrees, nil
	}

	return 0, fmt.Errorf("%w: bad degrees %q", ErrInvalidCoord, degStr)
}

func (cg *coordGroups) getMinutes() (float64, error) {
	if cg.min == "" {
		return 0, nil
	}

	idx := strings.IndexAny(cg.min, ".,")
	if idx != -1 && cg.sec != "" {
		return 0, fmt.Errorf("%w: minutes with decimal and seconds", ErrInvalidCoord)
	}
	minStr := cg.min
	if idx != -1 {
		minStr = cg.min[:idx] + "." + cg.min[idx+1:]
	}

	if minutes, err := strconv.ParseFloat(minStr, 64); err == nil {
		return checkLimits(minutes, 60, "minutes")
	}

	return 0, fmt.Errorf("%w: bad minutes %q", ErrInvalidCoord, minStr)
}

func (cg *coordGroups) getSeconds() (float64, error) {
	if cg.sec == "" {
		return 0, nil
	}

	idx := strings.IndexAny(cg.sec, ".,")
	secStr := cg.sec
	if idx != -1 {
		secStr = cg.sec[:idx] + "." + cg.sec[idx+1:]
	}

	if seconds, err := strconv.ParseFloat(secStr, 64); err == nil {
		return checkLimits(seconds, 60, "seconds")
	}

	return 0, fmt.Errorf("%w: bad seconds %q", ErrInvalidCoord, secStr)
}

func (cg *coordGroups) getCoord() (Coord, error) {
	var coord Coord
	if err := cg.checkSign(); err != nil {
		return coord, err
	}
	loc, err := cg.getLocation()
	if err != nil {
		return coord, err
	}
	deg, err := cg.getDegrees(loc)
	if err != nil {
		return coord, err
	}
	minutes, err := cg.getMinutes()
	if err != nil {
		return coord, err
	}
	sec, err := cg.getSeconds()
	if err != nil {
		return coord, err
	}
	unsigned := deg + minutes/60 + sec/3600
	limit := 180.0
	kind := "coordinate"
	if loc == LocLat {
		limit = 90.0
		kind = "latitude"
	} else if loc == LocLon {
		kind = "longitude"
	}
	if unsigned > limit {
		return coord, fmt.Errorf("%w: %s", ErrOutOfRange, kind)
	}

	coord.Value = unsigned
	if cg.sgn == "-" || cg.loc == "S" || cg.loc == "W" {
		coord.Value = -coord.Value
	}
	coord.Loc = loc

	return coord, nil
}

// StringToCoord converts string presentation of geographic coordinate to Coord.
// Deprecated: Use ParseCoord instead.
func StringToCoord(cs string) (Coord, error) {
	return ParseCoord(cs)
}

// StringToPoint converts a pair of geographic coordinates string to Point.
// Deprecated: Use ParsePoint instead.
func StringToPoint(lat string, lon string) (Point, error) {
	return ParsePoint(lat + "; " + lon)
}
