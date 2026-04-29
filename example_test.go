package geoc

import "fmt"

func ExampleParseCoord() {
	c, err := ParseCoord("48-33-27N")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Printf("%.4f %v\n", c.Value, c.Loc)
	// Output:
	// 48.5575 Lat
}

func ExampleParsePoint() {
	p, err := ParsePoint("48-33-27N; 120-57-49E")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Printf("%.4f %.6f\n", p.Lat.Value, p.Lon.Value)
	// Output:
	// 48.5575 120.963611
}

func ExampleCoord_Format() {
	c := Coord{Value: -48.5575, Loc: LocLat}
	s, err := c.Format(`48°33'27"N`)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(s)
	// Output:
	// 48°33'27"S
}

func ExampleCoord_String() {
	fmt.Println(Coord{Value: 48.5575, Loc: LocLat}.String())
	// Output:
	// 48-33.4N
}

func ExamplePoint_Format() {
	p, err := ParsePoint("48-33-27N; 120-57-49E")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	s, err := p.Format(`48°33'27"N`, `120°57'49"E`, "; ")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(s)
	// Output:
	// 48°33'27"N; 120°57'49"E
}

func ExampleFindCoords() {
	text := "WARNING 175/22 - mines reported near 39 27 25.55N - 009 39 25E, " +
		"stay clear within 3 NM."
	for _, m := range FindCoords(text, RequireDirection()) {
		fmt.Printf("%d-%d %s %q\n", m.Start, m.End, m.Coord.Loc, m.Text)
	}
	// Output:
	// 37-49 Lat "39 27 25.55N"
	// 52-62 Lon "009 39 25E"
}

func ExampleCoordRegexSource() {
	src := CoordRegexSource()
	fmt.Println(len(src) > 0)
	// Output:
	// true
}

func ExamplePoint_String() {
	p := Point{
		Lat: Coord{Value: 48.5575, Loc: LocLat},
		Lon: Coord{Value: 120.963611, Loc: LocLon},
	}
	fmt.Println(p.String())
	// Output:
	// 48-33.4N 120-57.8E
}
