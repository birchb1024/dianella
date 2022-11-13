package main

import (
	"flag"
	"fmt"
	. "github.com/birchb1024/dianella"
	"strings"
	"time"
)

type myStep struct {
	Step
	timestamp time.Time
	details   any
	dbUrl     string
}

func MyBEGIN(desc, url string) *myStep {
	m := myStep{dbUrl: url}
	m.Init(&m, desc)
	return &m
}
func (m *myStep) After() {
	fmt.Printf("%#v %s\n", m.details, time.Now().Sub(m.timestamp))
}
func (m *myStep) Before(info ...any) {
	// uncomment to get parental function also: m.Step.Before(info)
	m.timestamp = time.Now()
	m.details = info
}

func (m *myStep) PostgreSQL(query string) Stepper {
	// Mock
	if m.Self.IsFailed() {
		return m
	}
	m.Self.Before("PosgreSQL", query)
	defer m.Self.After()
	data := [][]string{{"Name", "Runs"}, {"Hales", "7"}, {"Butler", "54"}}
	fmt.Printf("%#v", data)
	return m
}

func main() {
	flag.Parse()
	var s *myStep
	s = MyBEGIN("Start example1", "postgres://localhost:5234/mydatabase")

	s.AND("Set a variable to the current date").
		Set("date", time.Now().String()).
		AND("Call a function").
		Call(func(s Stepper) Stepper {
			time.Sleep(5 * time.Second)
			fmt.Printf("%v\n", s.GetVar())
			return s
		}).
		AND("us bash to print the date").
		Bash("date").
		AND("Use bash with a template interpolation").
		Bash("echo {{.Var.date}}").
		AND("Create a temporary file name with Sbash")
	tmpFile, _ := s.Sbash("mktemp")
	tmpFile = strings.TrimSpace(tmpFile)
	s.Set("tmpFile", tmpFile).
		AND("Expand a template to the file").
		Expand("Date: {{.Var.date}}\n", tmpFile).
		AND("Use bash to send the file to stdout and then remove it").
		Bash("cat {{.Var.tmpFile}}").
		Bash("rm -f {{.Var.tmpFile}}")
	s.AND("Datbase query")
	s.PostgreSQL("select * from batters").
		END()
	s = s
}
