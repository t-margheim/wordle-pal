package main

import (
	"flag"
	"strings"
)

var (
	debug   bool
	rawPath string
	path    []string
	target  string
)

func init() {
	flag.BoolVar(&debug, "d", false, "enable debug logging")
	flag.StringVar(&rawPath, "p", "", "comma separated list of guesses to get to the solution")
	flag.StringVar(&target, "t", "", "target word for the wordle puzzle")
	flag.Parse()
	rawPath, target = strings.ToLower(rawPath), strings.ToLower(target)
	path = strings.Split(rawPath, ",")
}
