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

func newCoordGroups(names []string, values []string) coordGroups {
	cg := coordGroups{}
	for i, name := range names {
		if i != 0 && values[i] != "" {
			switch name {
			case "sgn":
				cg.sgn = values[i]
			case "deg":
				cg.deg = values[i]
			case "min":
				cg.min = values[i]
			case "sec":
				cg.sec = values[i]
			case "loc":
				cg.loc = values[i]
			}
		}
	}
	return cg
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
	idx := strings.Index(cg.deg, ".")

	// Check float degrees & exists minutes/seconds
	if idx != -1 && (cg.min != "" || cg.sec != "") {
		return 0, errors.New("invalid combination of degrees and minutes")
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

	idx := strings.Index(cg.min, ".")
	if idx != -1 && cg.sec != "" {
		return 0, errors.New("invalid combination of minutes and seconds")
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

var coordRegExp = regexp.MustCompile(
	`(?P<sgn>[-+])?` +
		`(?:(?P<deg>\d+(?:\.\d+)?)\s*[-°]?\s*)` +
		`(?:(?P<min>\d+(?:\.\d+)?)\s*[-']?\s*)?` +
		`(?:(?P<sec>\d+(?:\.\d+)?)\s*[-"]?\s*)?` +
		`(?P<loc>[NSEW])?`,
)

// StringToCoord converts geographic coordinate to string.
// Returns float64 representation of coordinate or error
// if coordinate string is invalid.
func StringToCoord(cs string) (float64, error) {
	makeErr := func(msg string) (float64, error) {
		return 0, fmt.Errorf("%s in string %q", msg, cs)
	}

	m := coordRegExp.FindAllStringSubmatch(cs, -1)
	if m == nil {
		return makeErr("unable to match coords pattern")
	}
	if len(m[0]) == 0 {
		return makeErr("invalid results of coords matching")
	}

	gc := newCoordGroups(coordRegExp.SubexpNames(), m[0])
	coord, err := gc.getCoord()
	if err != nil {
		return makeErr(err.Error())
	}

	return coord, nil
}
