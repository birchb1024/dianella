package main

import (
	"flag"
	. "github.com/birchb1024/dianella"
)

func main() {
	flag.Parse()
	s := BEGIN("example").
		Bash("ls -l").
		END()
	s = s
}
