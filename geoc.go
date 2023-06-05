// Package geoc provides geographic coordinate converter from string to native float64.
package geoc

import (
	"errors"
	"fmt"
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

var coordRegExp = regexp.MustCompile(
	`(\s*)` +
		`(?P<sgn>[-+])?` +
		`(?:(?P<deg>\d+(?:[\.,]\d+)?)(?P<dsr>\s*[-°]?\s*)?)` +
		`(?:(?P<min>\d+(?:[\.,]\d+)?)(?P<msr>\s*[-']?\s*)?)?` +
		`(?:(?P<sec>\d+(?:[\.,]\d+)?)(?P<ssr>\s*[ "]?\s*)?)?` +
		`(?P<loc>[NSEW])?(\s*)`,
)

func newCoordGroups(cs string) (*coordGroups, error) {
	m := coordRegExp.FindAllStringSubmatch(cs, -1)
	if m == nil {
		return nil, errors.New("unable to match coords pattern")
	}
	if len(m[0]) == 0 {
		return nil, errors.New("invalid results of coords matching")
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
	cg := coordGroups{}
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
		return nil, errors.New("invalid coordinate format")
	}
	return &cg, nil
}

type Location int

const (
	None Location = iota
	Lat
	Lon
)

func (cg *coordGroups) checkLocation(loc Location) error {
	if cg.loc == "" {
		return nil
	}

	if cg.loc != "N" && cg.loc != "S" && cg.loc != "E" && cg.loc != "W" {
		return fmt.Errorf("invalid location sign %q", cg.loc)
	}

	if loc == None {
		return nil
	}

	if loc == Lat && (cg.loc != "S" && cg.loc != "N") {
		return fmt.Errorf("invalid latitude location sign %q", cg.loc)
	}

	if loc == Lon && (cg.loc != "W" && cg.loc != "E") {
		return fmt.Errorf("invalid longtitude location sign %q", cg.loc)
	}

	return nil
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
		cg.fmt += "f"
	} else {
		cg.fmt += "i"
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
		cg.fmt += "f"
	} else { // 48-3327N format
		if len(cg.min) == 4 && cg.sec == "" && cg.loc != "" {
			cg.sec = cg.min[2:]
			cg.min = cg.min[:2]
		}
		cg.fmt += "i"
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
		cg.fmt += "f"
	} else {
		cg.fmt += "i"
	}
	cg.fmt += cg.sep.sec

	if seconds, err := strconv.ParseFloat(cg.sec, 64); err == nil {
		return checkLimits(seconds, 60, "seconds")
	}
	return 0, errors.New("unable to convert seconds to float")
}

func (cg *coordGroups) getCoord(loc Location) (float64, error) {
	if err := cg.checkSign(); err != nil {
		return 0, err
	}
	if err := cg.checkLocation(loc); err != nil {
		return 0, err
	}

	deg, err := cg.getDegrees(loc)
	if err != nil {
		return 0, err
	}
	min, err := cg.getMinutes()
	if err != nil {
		return 0, err
	}
	sec, err := cg.getSeconds()
	if err != nil {
		return 0, err
	}
	if cg.loc != "" {
		cg.fmt += "l"
	}

	coord := deg + min/60 + sec/3600
	if cg.sgn == "-" || cg.loc == "S" || cg.loc == "W" {
		coord = -coord
	}
	return coord, nil
}

type Point struct {
	Lat float64
	Lon float64
}

func (p *Point) String() string {
	return fmt.Sprintf(
		"[%s, %s]",
		strings.TrimRight(strconv.FormatFloat(p.Lat, 'f', 6, 64), "0"),
		strings.TrimRight(strconv.FormatFloat(p.Lon, 'f', 6, 64), "0"),
	)
}

// StringToCoord converts string presentation
// of geographic coordinate to native float number.
// Returns float64 value of coordinate or error
// if coordinate string is invalid.
func StringToCoord(cs string) (float64, error) {
	gc, err := newCoordGroups(cs)
	if err != nil {
		return 0, fmt.Errorf("%v in string %q", err, cs)
	}

	coord, err := gc.getCoord(None)
	if err != nil {
		return 0, fmt.Errorf("%v in string %q", err, cs)
	}

	return coord, nil
}

// StringToPoint converts a pair of geographic coordinates string to Point.
// Returns float64 representation of coordinates or error
// if coordinate string is invalid.
func StringToPoint(lat string, lon string) (Point, error) {
	retErr := func(err error, str string) (Point, error) {
		return Point{}, fmt.Errorf("%v in string %q", err, str)
	}

	gt, err := newCoordGroups(lat)
	if err != nil {
		return retErr(err, lat)
	}
	pt, err := gt.getCoord(Lat)
	if err != nil {
		return retErr(err, lat)
	}

	gn, err := newCoordGroups(lon)
	if err != nil {
		return retErr(err, lon)
	}
	pn, err := gn.getCoord(Lon)
	if err != nil {
		return retErr(err, lon)
	}

	if gt.fmt != gn.fmt {
		return Point{}, fmt.Errorf("formats of lat (%q) and lon (%q) strings are not identical", lat, lon)
	}

	return Point{pt, pn}, nil
}
