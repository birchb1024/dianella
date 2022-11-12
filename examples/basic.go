package main

import (
	"fmt"
	. "github.com/birchb1024/dianella"
	"strings"
	"time"
)

func main() {
	s := BEGIN("Start example1").
		Set("date", time.Now().String()).
		Call(func(s Stepper) Stepper {
			fmt.Printf("%v\n", s.GetVar())
			return s
		}).
		Bash("date").
		Bash("echo {{.Var.date}}")
	tmpFile, s := s.Sbash("mktemp")
	tmpFile = strings.TrimSpace(tmpFile)
	s.Set("tmpFile", tmpFile).
		Expand("Date: {{.Var.date}}\n", tmpFile).
		Bash("cat {{.Var.tmpFile}}").
		Bash("rm -f {{.Var.tmpFile}}")
	s.END()
	s = s
}
