// Package geoc parses and formats geographic coordinates and points.
// It supports conversion between string representations and native values.
package geoc

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

type Location int

const (
	None Location = iota // None means location type is not specified.
	Lat                  // Lat means coordinate is latitude.
	Lon                  // Lon means coordinate is longitude.
)

func (l Location) String() string {
	switch l {
	case Lat:
		return "Lat"
	case Lon:
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
	if loc == None {
		if cg.loc == "N" || cg.loc == "S" {
			loc = Lat
		} else if cg.loc == "E" || cg.loc == "W" {
			loc = Lon
		}
	}

	// Validate coord bounds
	absCoord := math.Abs(c.Value)
	if loc == Lat && absCoord > 90 {
		return "", fmt.Errorf("%w: latitude %f", ErrOutOfRange, c.Value)
	}
	if loc == Lon && absCoord > 180 {
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

	// Detect degree width for zero-padding (e.g., "048" → width 3)
	degWidth := len(cg.deg)
	if idx := strings.IndexAny(cg.deg, ".,"); idx != -1 {
		degWidth = idx
	}

	// Determine output location letter
	locLetter := ""
	if cg.loc != "" {
		if loc == Lat {
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
		var degStr string
		if precision > 0 {
			totalWidth := degWidth + 1 + precision
			degStr = fmt.Sprintf("%0*.*f", totalWidth, precision, absCoord)
		} else {
			degStr = fmt.Sprintf("%0*.0f", degWidth, absCoord)
		}
		if decSep == "," {
			degStr = strings.Replace(degStr, ".", ",", 1)
		}
		if negative && cg.loc == "" {
			degStr = "-" + degStr
		} else if cg.sgn == "+" {
			degStr = "+" + degStr
		}
		return degStr + cg.sep.deg + locLetter, nil
	}

	deg := math.Floor(absCoord)
	degStr := fmt.Sprintf("%0*.0f", degWidth, deg)

	// MinDec format
	if !hasSec {
		minutes := (absCoord - deg) * 60
		minStr := fmt.Sprintf("%.*f", precision, minutes)
		if decSep == "," {
			minStr = strings.Replace(minStr, ".", ",", 1)
		}
		return degStr + cg.sep.deg + minStr + cg.sep.min + locLetter, nil
	}

	// DMS format
	totalMin := (absCoord - deg) * 60
	minutes := math.Floor(totalMin)
	sec := (totalMin - minutes) * 60

	if cg.compact {
		return degStr + cg.sep.deg + fmt.Sprintf("%02.0f", minutes) + fmt.Sprintf("%02.0f", sec) + locLetter, nil
	}

	minStr := fmt.Sprintf("%.0f", minutes)
	secStr := fmt.Sprintf("%.*f", precision, sec)
	if decSep == "," {
		secStr = strings.Replace(secStr, ".", ",", 1)
	}
	return degStr + cg.sep.deg + minStr + cg.sep.min + secStr + cg.sep.sec + locLetter, nil
}

// String returns default string representation of the coordinate.
// Latitude uses MinDec format (48-33.0N), longitude uses MinDec
// with 3-digit degrees (048-33.0E), unspecified uses decimal degrees.
func (c Coord) String() string {
	var example string
	switch c.Loc {
	case Lat:
		example = "48-33.0N"
	case Lon:
		example = "048-33.0E"
	default:
		example = "48.557489"
	}
	s, err := c.Format(example)
	if err != nil {
		return strconv.FormatFloat(c.Value, 'f', -1, 64)
	}
	return s
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

	if lat.Loc != Lat {
		return p, fmt.Errorf("%w: bad latitude location in string %q", ErrInvalidString, s)
	}

	lon, err := cgLon.getCoord()
	if err != nil {
		return p, fmt.Errorf("%w in string %q", err, s)
	}

	if lon.Loc != Lon {
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
		return "", fmt.Errorf("%w: latitude %q", err, lat)
	}
	lon, err := p.Lon.Format(lonFmt)
	if err != nil {
		return "", fmt.Errorf("%w: longitude %q", err, lon)
	}
	return lat + separator + lon, nil
}

// String returns default string representation of the point.
// Default format is "48-33.0N 048-33.0E".
// If formatting fails, empty string is returned.
func (p Point) String() string {
	s, err := p.Format("48-33.0N", "048-33.0E", " ")
	if err != nil {
		return ""
	}
	return s
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
		strings.IndexAny(cg.min, ".,") == -1 {
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
	m := coordRegExp.FindAllStringSubmatch(cs, 3)
	if len(m) == 0 {
		return cgLat, cgLon, fmt.Errorf("%w: coords not found", ErrInvalidString)
	}
	if len(m) == 1 {
		return cgLat, cgLon, fmt.Errorf("%w: too few coords found", ErrInvalidString)
	}
	if len(m) > 2 {
		return cgLat, cgLon, fmt.Errorf("%w: too many coords found", ErrInvalidString)
	}

	sen := coordRegExp.SubexpNames()
	cgLat, _ = coordGroupsFromMatch(m[0], sen)
	cgLon, _ = coordGroupsFromMatch(m[1], sen)

	cgLat.normalizeCompact()
	cgLon.normalizeCompact()
	return cgLat, cgLon, nil
}

func (cg *coordGroups) getLocation() (Location, error) {
	if cg.loc == "N" || cg.loc == "S" {
		return Lat, nil
	}
	if cg.loc == "E" || cg.loc == "W" {
		return Lon, nil
	}
	if cg.loc == "" {
		return None, nil
	}

	return None, fmt.Errorf("%w: bad location sign %q", ErrInvalidCoord, cg.loc)
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
	if idx != -1 {
		cg.deg = cg.deg[:idx] + "." + cg.deg[idx+1:]
	}
	if degrees, err := strconv.ParseFloat(cg.deg, 64); err == nil {
		limit := 180.0
		if cg.loc == "S" || cg.loc == "N" || loc == Lat {
			limit = 90.0
		}
		if degrees > limit {
			return 0, fmt.Errorf("%w: degrees", ErrOutOfRange)
		}
		return degrees, nil
	}

	return 0, fmt.Errorf("%w: bad degrees %q", ErrInvalidCoord, cg.deg)
}

func (cg *coordGroups) getMinutes() (float64, error) {
	if cg.min == "" {
		return 0, nil
	}

	idx := strings.IndexAny(cg.min, ".,")
	if idx != -1 && cg.sec != "" {
		return 0, fmt.Errorf("%w: minutes with decimal and seconds", ErrInvalidCoord)
	}
	if idx != -1 {
		cg.min = cg.min[:idx] + "." + cg.min[idx+1:]
	}

	if minutes, err := strconv.ParseFloat(cg.min, 64); err == nil {
		return checkLimits(minutes, 60, "minutes")
	}

	return 0, fmt.Errorf("%w: bad minutes %q", ErrInvalidCoord, cg.min)
}

func (cg *coordGroups) getSeconds() (float64, error) {
	if cg.sec == "" {
		return 0, nil
	}

	idx := strings.IndexAny(cg.sec, ".,")
	if idx != -1 {
		cg.sec = cg.sec[:idx] + "." + cg.sec[idx+1:]
	}

	if seconds, err := strconv.ParseFloat(cg.sec, 64); err == nil {
		return checkLimits(seconds, 60, "seconds")
	}

	return 0, fmt.Errorf("%w: bad seconds %q", ErrInvalidCoord, cg.sec)
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
	coord.Value = deg + minutes/60 + sec/3600
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
