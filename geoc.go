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
}

var coordRegExp = regexp.MustCompile(
	`(\s*)` +
		`(?P<sgn>[-+])?` +
		`(?:(?P<deg>\d+(?:[\.,]\d+)?)(\s*[-°]?\s*))` +
		`(?:(?P<min>\d+(?:[\.,]\d+)?)(\s*[-']?\s*))?` +
		`(?:(?P<sec>\d+(?:[\.,]\d+)?)(\s*[-"]?\s*))?` +
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
			}
			totalLen += len(value)
		}
	}

	if totalLen != len(cs) {
		return nil, errors.New("invalid coordinate string")
	}
	return &cg, nil
}

func (cg coordGroups) checkLocation() error {
	if cg.loc == "" || cg.loc == "N" || cg.loc == "S" || cg.loc == "E" || cg.loc == "W" {
		return nil
	}
	return fmt.Errorf("invalid location sign %q", cg.loc)
}

func (cg coordGroups) checkSign() error {
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

func (cg coordGroups) getDegrees() (float64, error) {
	if cg.deg == "" {
		return 0, errors.New("missing degrees")
	}

	// Check float degrees & exists minutes/seconds
	idx := strings.IndexAny(cg.deg, ".,")
	if idx != -1 && (cg.min != "" || cg.sec != "") {
		return 0, errors.New("invalid combination of degrees and minutes")
	}
	if idx != -1 {
		cg.deg = cg.deg[:idx] + "." + cg.deg[idx+1:]
	}

	if degrees, err := strconv.ParseFloat(cg.deg, 64); err == nil {
		limit := 180.0
		if cg.loc == "S" || cg.loc == "N" {
			limit = 90.0
		}
		return checkLimits(degrees, limit, "degrees")
	}
	return 0, errors.New("unable to convert degrees to float")
}

func (cg coordGroups) getMinutes() (float64, error) {
	if cg.min == "" {
		return 0, nil
	}

	idx := strings.IndexAny(cg.min, ".,")
	if idx != -1 && cg.sec != "" {
		return 0, errors.New("invalid combination of minutes and seconds")
	}
	if idx != -1 {
		cg.min = cg.min[:idx] + "." + cg.min[idx+1:]
	}

	if minutes, err := strconv.ParseFloat(cg.min, 64); err == nil {
		return checkLimits(minutes, 60, "minutes")
	}
	return 0, errors.New("unable to convert minutes to float")
}

func (cg coordGroups) getSeconds() (float64, error) {
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
	return 0, errors.New("unable to convert seconds to float")
}

func (cg coordGroups) getCoord() (float64, error) {
	if err := cg.checkSign(); err != nil {
		return 0, err
	}
	if err := cg.checkLocation(); err != nil {
		return 0, err
	}
	deg, err := cg.getDegrees()
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

	coord := deg + min/60 + sec/3600
	if cg.sgn == "-" || cg.loc == "S" || cg.loc == "W" {
		coord = -coord
	}
	return coord, nil
}

// StringToCoord converts geographic coordinate to string.
// Returns float64 representation of coordinate or error
// if coordinate string is invalid.
func StringToCoord(cs string) (float64, error) {
	gc, err := newCoordGroups(cs)
	if err != nil {
		return 0, fmt.Errorf("%q in string %q", err, cs)
	}

	coord, err := gc.getCoord()
	if err != nil {
		return 0, fmt.Errorf("%q in string %q", err, cs)
	}

	return coord, nil
}
