package main

import (
	"fmt"
	. "github.com/birchb1024/dianella"
	"strings"
	"time"
)

type myStep struct {
	Step
	timestamp time.Time
}

func MyBEGIN(desc string) *myStep {
	m := myStep{}
	m.Init(&m, desc)
	return &m
}
func (m *myStep) After() {
	fmt.Printf("%s %s\n", m.GetDescription(), time.Now().Sub(m.timestamp))
}
func (m *myStep) Before() { m.timestamp = time.Now() }

func main() {
	var s Stepper
	s = MyBEGIN("Start example1").
		AND("Set a variable to the current date")
	s.Set("date", time.Now().String()).
		AND("Call a function").
		Call(func(s Stepper) Stepper {
			fmt.Printf("%v\n", s.GetVar())
			return s
		}).
		AND("us bash to print the date").
		Bash("date").
		AND("Use bash with a template interpolation").
		Bash("echo {{.Var.date}}").
		AND("Create a temporary file name with Sbash")
	tmpFile, s := s.Sbash("mktemp")
	tmpFile = strings.TrimSpace(tmpFile)
	s.Set("tmpFile", tmpFile).
		AND("Expand a template to the file").
		Expand("Date: {{.Var.date}}\n", tmpFile).
		AND("Use bash to send the file to stdout and then remove it").
		Bash("cat {{.Var.tmpFile}}").
		Bash("rm -f {{.Var.tmpFile}}")
	s.END()
	s = s
}