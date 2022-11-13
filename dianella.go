package dianella

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"text/template"
)

type Stepper interface {
	After()
	AND(string) Stepper
	Bash(command string) Stepper
	Before(...any)
	Call(func(Stepper) Stepper) Stepper
	END() Stepper
	Expand(template string, outputFileName string) Stepper
	Fail(msg string) Stepper
	FailErr(e error)
	Init(Stepper, string)
	IsFailed() bool
	Sbash(cmd string) (string, Stepper)
	Set(variableName string, value any) Stepper
	Sexpand(cmd string) string

	GetDescription() string
	GetErr() error
	GetArg() []string
	GetFlag() map[string]any
	GetVar() map[string]any
	GetStatus() int
}

type Step struct {
	Arg         []string
	Flag        map[string]any
	Var         map[string]any
	description string
	err         error
	Self        Stepper
	status      int
}

func (s *Step) GetArg() []string        { return s.Arg }
func (s *Step) GetVar() map[string]any  { return s.Var }
func (s *Step) GetFlag() map[string]any { return s.Flag }
func (s *Step) GetDescription() string  { return s.description }
func (s *Step) GetErr() error           { return s.err }
func (s *Step) GetStatus() int          { return s.status }

func (s *Step) After() {}
func (s *Step) Before(info ...any) {
	v, ok := s.Var["trace"]
	if !ok {
		return
	}
	if v != true {
		return
	}
	log.Printf("INFO: %v", info)
}
func (s *Step) Init(st Stepper, desc string) {
	s.Self = st
	s.description = desc
	s.Var = map[string]any{}
	s.Flag = map[string]any{}
	s.Arg = flag.Args()
	s.Var["trace"] = true
}

func (s *Step) FailErr(e error) {
	s.Self.Before("FailErr", e)
	defer s.Self.After()
	s.err = e
	s.status = 1
}
func (s *Step) Fail(msg string) Stepper {
	s.Self.Before("Fail", msg)
	defer s.Self.After()
	s.status = 1
	s.err = fmt.Errorf(msg)
	return s
}

func (s *Step) Set(name string, value any) Stepper {
	if s.Self.IsFailed() {
		return s
	}
	s.Self.Before("Set", name, value)
	defer s.Self.After()
	sv, ok := value.(string)
	if ok {
		s.Var[name] = Expando(sv, s)
	} else {
		s.Var[name] = value
	}
	return s
}
func BEGIN(desc string) *Step {
	s := Step{
		description: desc,
		status:      0,
		err:         nil,
		Var:         map[string]any{"trace": true},
		Flag:        map[string]any{},
		Arg:         flag.Args(),
	}
	s.Self = &s
	flag.VisitAll(func(f *flag.Flag) { s.Flag[f.Name] = f.Value })

	return &s
}
func (s *Step) IsFailed() bool {
	return s.Self.GetStatus() != 0 || s.Self.GetErr() != nil
}
func (s *Step) END() Stepper {
	if s.Self.IsFailed() {
		log.Printf("ERROR: END '%s' failed with status %d, %s", s.Self.GetDescription(), s.Self.GetStatus(), s.Self.GetErr())
		os.Exit(1)
	}
	s.Self.Before("End")
	defer s.Self.After()
	return s
}

func (s *Step) dieIfFailed(name string) {
	if s.Self.IsFailed() {
		log.Printf("ERROR: %s '%s' failed with status %d, %s", name, s.Self.GetDescription(), s.Self.GetStatus(), s.Self.GetErr())
		os.Exit(1)
	}
}
func (s *Step) AND(desc string) Stepper {
	if s.Self.IsFailed() {
		return s
	}
	s.Self.Before("AND", desc)
	defer s.Self.After()
	s.description = desc
	return s
}

func (s *Step) Bash(cmd string) Stepper {
	if s.Self.IsFailed() {
		return s
	}
	s.Self.Before("Bash", cmd)
	defer s.Self.After()
	c := exec.Command("/bin/bash", "-c", Expando(cmd, s))
	err := c.Run()
	if err != nil {
		s.Self.FailErr(err)
	}
	return s
}

func (s *Step) Sbash(cmd string) (result string, rs Stepper) {
	s.dieIfFailed("Sbash")
	s.Self.Before("Sbash", cmd)
	defer s.Self.After()
	c := exec.Command("/bin/bash", "-c", Expando(cmd, s))
	c.Stderr = os.Stderr
	stdoutBytes, err := c.Output()
	if err != nil {
		s.Self.FailErr(err)
	}
	return string(stdoutBytes), s
}

// Expando - Use Go template module to interpolate expansions in a string
// using data from the environmnent (the SICP sense of environment)
func intMin(x, y int) int {
	if x > y {
		return y
	}
	return x
}
func Expando(templateSource string, environment any) string {
	temp, err := template.New("Expando").Parse(templateSource)
	if err != nil {
		log.Fatalf("ERROR: %s", err)
	}
	var buf bytes.Buffer
	err = temp.Execute(&buf, environment)
	if err != nil {
		log.Fatalf("ERROR: %s", err)
	}
	return buf.String()
}
func (s *Step) Expand(temp string, filename string) Stepper {
	if s.Self.IsFailed() {
		return s
	}
	s.Self.Before("Expand", temp[:intMin(len(temp)-1, 20)], filename)
	defer s.Self.After()
	expanded := Expando(temp, s)
	err := os.WriteFile(filename, []byte(expanded), 0644)
	if err != nil {
		log.Fatalf("ERROR: ", err)
	}
	return s
}
func (s *Step) Sexpand(template string) string {
	s.dieIfFailed("Sexpand")
	s.Self.Before("Sexpand", template[:intMin(len(template)-1, 20)])
	defer s.Self.After()
	return Expando(template, s)
}

func (s *Step) Call(f func(s Stepper) Stepper) Stepper {
	if s.Self.IsFailed() {
		return s
	}
	s.Self.Before("Call")
	defer s.Self.After()
	f(s.Self)
	return s
}
