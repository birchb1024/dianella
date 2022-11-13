package main

import (
	"flag"
	"fmt"
	. "github.com/birchb1024/dianella"
	"strings"
	"time"
)

func printAllVariables(s Stepper) Stepper {
	fmt.Printf("%#v\n", s.GetVar())
	return s
}
func main() {
	flag.Parse()
	var s Stepper = BEGIN("Start example1").
		// Set("trace", false).
		Set("date", time.Now().String()).
		Call(printAllVariables).
		Bash("date").
		Bash("echo {{.Var.date}}")
	tmpFile, s := s.Sbash("mktemp")
	tmpFile = strings.TrimSpace(tmpFile)
	s.Set("tmpFile", tmpFile).
		Expand("tmpFile - Date: {{.Var.date}}\n", tmpFile).
		Bash("cat {{.Var.tmpFile}}").
		Bash("rm -f {{.Var.tmpFile}}").
		END()
	s = s
}
