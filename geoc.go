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
		if cg.loc == "N" || cg.loc == "S" {
			loc = LocLat
		} else if cg.loc == "E" || cg.loc == "W" {
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
	if hasSec && !cg.compact {
		detectDecimal(cg.sec)
	} else if hasMin && !cg.compact {
		detectDecimal(cg.min)
	} else if !hasMin {
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
		if loc == LocLat {
			locLetter = "N"
			if negative {
				locLetter = "S"
			}
		} else {
			locLetter = "E"
			if negative {
				locLetter = "W"
			}
		}
	}

	// DegDec format
	if !hasMin {
		degStr := formatFixed(absCoord, degWidth, precision)
		if negative && cg.loc == "" {
			degStr = "-" + degStr
		} else if cg.sgn == "+" {
			degStr = "+" + degStr
		}
		return degStr + cg.sep.deg + locLetter, nil
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
		return degStr + cg.sep.deg + minStr + cg.sep.min + locLetter, nil
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
		return degStr + cg.sep.deg + minStr + secStr + locLetter, nil
	}

	minStr := formatFixed(minutes, minWidth, 0)
	secStr := formatFixed(sec, secWidth, precision)
	return degStr + cg.sep.deg + minStr + cg.sep.min + secStr + cg.sep.sec + locLetter, nil
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

// normalizeCompact splits compact MMSS minutes (e.g., "5749")
// into separate min ("57") and sec ("49") fields.
func (cg *coordGroups) normalizeCompact() {
	if len(cg.min) == 4 && cg.sec == "" && cg.loc != "" &&
		!strings.ContainsAny(cg.min, ".,") {
		cg.compact = true
		cg.sec = cg.min[2:]
		cg.min = cg.min[:2]
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

var coordRegExp = regexp.MustCompile(
	`(\s*)` +
		`(?P<sgn>[-+])?` +
		`(?:(?P<deg>\d+(?:[\.,]\d+)?)(?P<dsr>\s*[-°]?\s*)?)` +
		`(?:(?P<min>\d+(?:[\.,]\d+)?)(?P<msr>\s*[-']?\s*)?)?` +
		`(?:(?P<sec>\d+(?:[\.,]\d+)?)(?P<ssr>\s*[ "]?\s*)?)?` +
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
	if cg.loc == "N" || cg.loc == "S" {
		return LocLat, nil
	}
	if cg.loc == "E" || cg.loc == "W" {
		return LocLon, nil
	}
	if cg.loc == "" {
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
