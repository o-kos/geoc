// Package geoc provides geographic coordinate converter from string to native float64.
package geoc

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

type coordGroups struct {
	sgn string
	deg string
	min string
	sec string
	loc string
	sep struct {
		deg string
		min string
		sec string
	}
	fmt string
}

func (cg *coordGroups) formatClass() string {
	if cg.min == "" {
		return "degdec"
	}
	if cg.sec == "" {
		return "mindec"
	}
	return "dms"
}

var coordRegExp = regexp.MustCompile(
	`(\s*)` +
		`(?P<sgn>[-+])?` +
		`(?:(?P<deg>\d+(?:[\.,]\d+)?)(?P<dsr>\s*[-°]?\s*)?)` +
		`(?:(?P<min>\d+(?:[\.,]\d+)?)(?P<msr>\s*[-']?\s*)?)?` +
		`(?:(?P<sec>\d+(?:[\.,]\d+)?)(?P<ssr>\s*[ "]?\s*)?)?` +
		`(?P<loc>[NSEW])?(\s*)`,
)

func newCoordGroups(cs string) (coordGroups, error) {
	cg := coordGroups{}
	m := coordRegExp.FindAllStringSubmatch(cs, -1)
	if m == nil {
		return cg, errors.New("unable to match coords pattern")
	}
	if len(m[0]) == 0 {
		return cg, errors.New("invalid results of coords matching")
	}

	makeSep := func(sep string) string {
		if ret := strings.TrimSpace(sep); ret != "" {
			return ret
		}
		if sep != "" {
			return " "
		}
		return ""
	}

	totalLen := 0
	for i, name := range coordRegExp.SubexpNames() {
		value := m[0][i]
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

	if totalLen != len(cs) {
		return cg, errors.New("invalid coordinate format")
	}
	return cg, nil
}

func newPointGroups(cs string) (coordGroups, coordGroups, error) {
	var cgLat coordGroups
	var cgLon coordGroups
	m := coordRegExp.FindAllStringSubmatch(cs, -1)
	if m == nil {
		return cgLat, cgLon, errors.New("unable to match coords pattern")
	}
	if len(m) != 2 {
		return cgLat, cgLon, errors.New("invalid coords count")
	}
	if len(m[0]) == 0 {
		return cgLat, cgLon, errors.New("invalid results of lat-coords matching")
	}
	if len(m[1]) == 0 {
		return cgLat, cgLon, errors.New("invalid results of lon-coords matching")
	}

	//	fillCoordGroups(cs)

	return cgLat, cgLon, nil
}

type Location int

const (
	None Location = iota
	Lat
	Lon
)

// Coord represents a geographic coordinate with its location type.
type Coord struct {
	Value float64
	Loc   Location
}

// CoordError wraps coordinate parsing errors with the original input string.
type CoordError struct {
	Coord string
	Err   error
}

func (e CoordError) Error() string {
	if e.Err == nil {
		return "invalid coordinate"
	}
	return e.Err.Error() + " in string " + strconv.Quote(e.Coord)
}

func (e CoordError) Unwrap() error {
	return e.Err
}

// Format converts coordinate to string representation
// using provided example format string.
func (c Coord) Format(example string) (string, error) {
	cg, err := newCoordGroups(example)
	if err != nil {
		return "", fmt.Errorf("invalid example format: %v", err)
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
	if loc == Lat && absCoord >= 90 {
		return "", fmt.Errorf("coordinate %f out of range for latitude", c.Value)
	}
	if loc == Lon && absCoord >= 180 {
		return "", fmt.Errorf("coordinate %f out of range for longitude", c.Value)
	}

	negative := c.Value < 0

	// Detect compact format (e.g., 120-5749E): 4 integer digits in minutes
	isCompact := len(cg.min) == 4 && cg.sec == "" && cg.loc != "" && strings.IndexAny(cg.min, ".,") == -1
	hasSec := cg.sec != "" || isCompact
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
	if hasSec && !isCompact {
		detectDecimal(cg.sec)
	} else if hasMin && !isCompact {
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

	if isCompact {
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

	return None, fmt.Errorf("invalid location sign %q", cg.loc)
}

func (cg *coordGroups) checkSign() error {
	if cg.sgn == "" {
		return nil
	}
	if (cg.sgn == "+" || cg.sgn == "-") && cg.loc != "" {
		return errors.New("invalid combination sign & location symbols")
	}
	return nil
}

func checkLimits(value float64, limit float64, kind string) (float64, error) {
	if value < limit {
		return value, nil
	}
	return 0, fmt.Errorf("%s out of range", kind)
}

func (cg *coordGroups) getDegrees(loc Location) (float64, error) {
	if cg.deg == "" {
		return 0, errors.New("missing degrees")
	}

	cg.fmt = "d"
	// Check float degrees & exists minutes/seconds
	idx := strings.IndexAny(cg.deg, ".,")
	if idx != -1 && (cg.min != "" || cg.sec != "") {
		return 0, errors.New("invalid combination of degrees and minutes")
	}
	if idx != -1 {
		cg.deg = cg.deg[:idx] + "." + cg.deg[idx+1:]
	}
	cg.fmt += cg.sep.deg

	if degrees, err := strconv.ParseFloat(cg.deg, 64); err == nil {
		limit := 180.0
		if cg.loc == "S" || cg.loc == "N" || loc == Lat {
			limit = 90.0
		}
		return checkLimits(degrees, limit, "degrees")
	}
	return 0, errors.New("unable to convert degrees to float")
}

func (cg *coordGroups) getMinutes() (float64, error) {
	if cg.min == "" {
		return 0, nil
	}

	cg.fmt += "m"
	idx := strings.IndexAny(cg.min, ".,")
	if idx != -1 && cg.sec != "" {
		return 0, errors.New("invalid combination of minutes and seconds")
	}
	if idx != -1 {
		cg.min = cg.min[:idx] + "." + cg.min[idx+1:]
	} else { // 48-3327N format
		if len(cg.min) == 4 && cg.sec == "" && cg.loc != "" {
			cg.sec = cg.min[2:]
			cg.min = cg.min[:2]
		}
	}
	cg.fmt += cg.sep.min

	if minutes, err := strconv.ParseFloat(cg.min, 64); err == nil {
		return checkLimits(minutes, 60, "minutes")
	}
	return 0, errors.New("unable to convert minutes to float")
}

func (cg *coordGroups) getSeconds() (float64, error) {
	if cg.sec == "" {
		return 0, nil
	}

	cg.fmt += "s"
	idx := strings.IndexAny(cg.sec, ".,")
	if idx != -1 {
		cg.sec = cg.sec[:idx] + "." + cg.sec[idx+1:]
	}
	cg.fmt += cg.sep.sec

	if seconds, err := strconv.ParseFloat(cg.sec, 64); err == nil {
		return checkLimits(seconds, 60, "seconds")
	}
	return 0, errors.New("unable to convert seconds to float")
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
	if cg.loc != "" {
		cg.fmt += "l"
	}

	coord.Value = deg + minutes/60 + sec/3600
	if cg.sgn == "-" || cg.loc == "S" || cg.loc == "W" {
		coord.Value = -coord.Value
	}
	return coord, nil
}

// Point represents a geographic point with latitude and longitude.
type Point struct {
	Lat Coord
	Lon Coord
}

// Format converts Point to string representation using provided
// format examples for latitude and longitude coordinates.
func (p Point) Format(latFmt, lonFmt string) (string, error) {
	lat, err := p.Lat.Format(latFmt)
	if err != nil {
		return "", fmt.Errorf("latitude: %v", err)
	}
	lon, err := p.Lon.Format(lonFmt)
	if err != nil {
		return "", fmt.Errorf("longitude: %v", err)
	}
	return lat + "; " + lon, nil
}

func (p Point) String() string {
	s, err := p.Format("48-33.0N", "048-33.0E")
	if err != nil {
		return fmt.Sprintf(
			"[%s, %s]",
			strings.TrimRight(strconv.FormatFloat(p.Lat.Value, 'f', 6, 64), "0"),
			strings.TrimRight(strconv.FormatFloat(p.Lon.Value, 'f', 6, 64), "0"),
		)
	}
	return s
}

// parseCoordGroups parses a coordinate string, determines its location type,
// and returns the parsed groups, coordinate value, and location.
// func parseCoordGroups(s string, loc Location) (coordGroups, float64, error) {
// 	cg, err := newCoordGroups(s)
// 	if err != nil {
// 		return nil, 0, fmt.Errorf("%v in string %q", err, s)
// 	}

// 	if loc == None {
// 		if cg.loc == "N" || cg.loc == "S" {
// 			loc = Lat
// 		} else if cg.loc == "E" || cg.loc == "W" {
// 			loc = Lon
// 		}
// 	}

// 	value, err := cg.getCoord(loc)
// 	if err != nil {
// 		return nil, 0, fmt.Errorf("%v in string %q", err, s)
// 	}

// 	return cg, value, nil
// }

// ParseCoord parses a coordinate string and returns a Coord.
// Location type (Lat/Lon) is auto-detected from the location letter (N/S/E/W).
func ParseCoord(s string) (Coord, error) {
	coord := Coord{}
	cg, err := newCoordGroups(s)
	if err != nil {
		return coord, fmt.Errorf("%v in string %q", err, s)
	}

	coord, err = cg.getCoord()
	if err != nil {
		return coord, fmt.Errorf("%v in string %q", err, s)
	}

	return coord, nil
}

// ParsePoint parses a string containing latitude and longitude.
// Latitude is parsed from the beginning of the string; longitude is then
// searched to the right starting from the next symbol after latitude.
func ParsePoint(s string) (Point, error) {
	p := Point{}
	cgLat, cgLon, err := newPointGroups(s)
	if err != nil {
		return p, CoordError{Coord: s, Err: err}
	}
	if cgLat.formatClass() != cgLon.formatClass() {
		return p, fmt.Errorf("formats of lat and lon parts of string (%q) are incompatible", s)
	}

	lat, err := cgLat.getCoord()
	if err != nil {
		return p, CoordError{Coord: s, Err: err}
	}

	lon, err := cgLon.getCoord()
	if err != nil {
		return p, CoordError{Coord: s, Err: err}
	}

	return Point{lat, lon}, nil
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
