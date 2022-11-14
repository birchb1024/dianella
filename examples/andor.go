package main

import (
	"flag"
	. "github.com/birchb1024/dianella"
	"time"
)

func main() {
	flag.Parse()
	var s Stepper = BEGIN("Get the star date, Cassian").
		Set("date", time.Now().Unix()).
		AND("Cassian, fail to Print the date").
		Bash("ZZZZecho {{.Var.date}}").
		CONTINUE("Bix resets the error and prints the date").
		Bash("echo {{.Var.date}}").
		END()
	s = s
}
