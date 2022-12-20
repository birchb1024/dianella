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
	var s Stepper = BEGIN("Basic example").
		Set("trace_length", 2000).
		Set("date", time.Now().String()).
		Call(printAllVariables).
		Bash("date").
		Bash("echo {{.Var.date}}")
	tmpFile, s := s.Sbash("mktemp")
	s.Set("tmpFile", strings.TrimSpace(tmpFile)).
		Expand("tmpFile: {{.Var.tmpFile}} - Date: {{.Var.date}}\n", strings.TrimSpace(tmpFile)).
		Bash("cat {{.Var.tmpFile}}").
		Bash("rm -f {{.Var.tmpFile}}").
		END()
	s = s
}
