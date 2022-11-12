package dianella

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/bitfield/script"
	"log"
	"os"
	"os/exec"
	"text/template"
)

type Stepper interface {
	After()
	AND(string) Stepper
	Bash(command string) Stepper
	Before()
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
	self        Stepper
	status      int
}

func (s *Step) GetArg() []string        { return s.Arg }
func (s *Step) GetVar() map[string]any  { return s.Var }
func (s *Step) GetFlag() map[string]any { return s.Flag }
func (s *Step) GetDescription() string  { return s.description }
func (s *Step) GetErr() error           { return s.err }
func (s *Step) GetStatus() int          { return s.status }

func (s *Step) After()  {}
func (s *Step) Before() {}
func (s *Step) Init(st Stepper, desc string) {
	s.self = st
	s.description = desc
	s.Var = map[string]any{}
	s.Flag = map[string]any{}
	s.Arg = flag.Args()
}

func (s *Step) FailErr(e error) { s.err = e; s.status = 1 }

func (s *Step) Expand(temp string, filename string) Stepper {
	if s.self.GetStatus() != 0 {
		return s
	}
	expanded := Expando(temp, s)
	err := os.WriteFile(filename, []byte(expanded), 0644)
	if err != nil {
		panic(err)
	}
	return s
}

func (s *Step) Set(name string, value any) Stepper {
	s.self.Before()
	defer s.self.After()
	if s.GetStatus() != 0 {
		return s
	}
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
		Var:         map[string]any{},
		Flag:        map[string]any{},
		Arg:         flag.Args(),
	}
	s.self = &s
	flag.VisitAll(func(f *flag.Flag) { s.Flag[f.Name] = f.Value })

	log.Printf("INFO: begin %s", desc)
	return &s
}
func (s *Step) IsFailed() bool {
	return s.self.GetStatus() != 0 || s.self.GetErr() != nil
}
func (s *Step) END() Stepper {
	if s.self.IsFailed() {
		log.Printf("ERROR: END '%s' failed with status %d, %s", s.self.GetDescription(), s.self.GetStatus(), s.self.GetErr())
		os.Exit(1)
	}
	log.Printf("INFO: END '%s'", s.self.GetDescription())
	return s
}
func (s *Step) AND(desc string) Stepper {
	if s.self.IsFailed() {
		log.Printf("ERROR: AND '%s' failed with status %d, %s", s.self.GetDescription(), s.self.GetStatus(), s.self.GetErr())
		os.Exit(1)
	}
	s.description = desc
	log.Printf("INFO: AND '%s'", s.self.GetDescription())
	return s
}

func (s *Step) Bash(cmd string) Stepper {
	s.self.Before()
	defer s.self.After()
	log.Printf("INFO: Bash '%s'", cmd)
	if s.self.GetStatus() != 0 {
		return s
	}
	bc := fmt.Sprintf("bash -c '%s'", s.self.Sexpand(cmd))
	log.Printf("DEBUG: %s", bc)
	_, err := script.Exec(bc).Stdout()

	if err != nil {
		s.status = 1
	}
	s.err = err
	return s
}

func (s *Step) Sbash(cmd string) (result string, rs Stepper) {
	s.self.Before()
	defer s.self.After()
	log.Printf("INFO: Sbash '%s'", cmd)
	if s.self.GetStatus() != 0 {
		return "", s
	}
	bc := fmt.Sprintf("bash -c '%s'", s.self.Sexpand(cmd))
	log.Printf("DEBUG: %s", bc)

	c := exec.Command("/bin/bash", "-c", cmd)
	c.Stderr = os.Stderr
	stdoutBytes, err := c.Output()
	if err != nil {
		s.FailErr(err)
	}
	return string(stdoutBytes), s
}

// Expando - Use Go template module to interpolate expansions in a string
// using data from the environmnent (the SICP sense of environment)
func Expando(templateSource string, environment any) string {
	temp, err := template.New("Expando").Parse(templateSource)
	if err != nil {
		panic(err)
	}
	var buf bytes.Buffer
	err = temp.Execute(&buf, environment)
	if err != nil {
		panic(err)
	}
	return buf.String()
}

func (s *Step) Sexpand(template string) string {
	s.self.Before()
	defer s.self.After()
	return Expando(template, s)
}

func (s *Step) Fail(msg string) Stepper {
	s.status = 1
	s.err = fmt.Errorf(msg)
	return s
}

func (s *Step) Call(f func(s Stepper) Stepper) Stepper {
	s.self.Before()
	defer s.self.After()
	if s.self.GetStatus() != 0 {
		return s
	}
	f(s.self)
	return s
}
