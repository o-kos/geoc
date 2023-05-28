# geoc - Golang geographic coordinates converter

Provide conversion from popular string representation of
geographic coordinates to golang-native float64 format.

## Supported formats

`geoc` support coordinate in the three basic forms:

- Degrees (integer), minutes (integer), and seconds (integer, or real number) (DMS).
  - `48°33'29"N`,
  - `48-33-29 N`,
  - `48-33-29.4768N`, etc.
- Degrees (integer) and minutes (real number) (MinDec).
  - `48°33.49128"N`,
  - `48-33N`,
  - `48-33.49128N`, etc.
- Degrees (real number) (DegDec).
  - `48.558188`,
  - `+48.558188`,
  - `-39.298358`, etc.

### Installation

Once you have [installed Go][golang-install], run this command
to install the `geoc` package:

    go get github.com/o-kos/geoc

[golang-install]: http://golang.org/doc/install.html
