package dianella

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/bitfield/script"
	"log"
	"os"
	"text/template"
)

func (s *Step) Expand(temp string, filename string) *Step {
	if s.ExitStatus() != 0 {
		return s
	}
	expanded := expando(temp, s)
	err := os.WriteFile(filename, []byte(expanded), 0644)
	if err != nil {
		panic(err)
	}
	return s
}

func (s *Step) Set(name string, value any) *Step {
	if s.ExitStatus() != 0 {
		return s
	}
	sv, ok := value.(string)
	if ok {
		s.Var[name] = expando(sv, s)
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
	flag.VisitAll(func(f *flag.Flag) { s.Flag[f.Name] = f.Value })

	log.Printf("INFO: begin %s", desc)
	return &s
}
func (s *Step) END() *Step {
	if s.status != 0 || s.err != nil {
		log.Printf("ERROR: END '%s' failed with status %d, %s", s.description, s.status, s.err)
		os.Exit(1)
	}
	log.Printf("INFO: END '%s'", s.description)
	return s
}
func (s *Step) AND(desc string) *Step {
	if s.status != 0 || s.err != nil {
		log.Printf("ERROR: AND '%s' failed with status %d, %s", s.description, s.status, s.err)
		os.Exit(1)
	}
	s.description = desc
	log.Printf("INFO: AND '%s'", s.description)
	return s
}

func (s *Step) Bash(cmd string) *Step {
	log.Printf("INFO: Bash '%s'", cmd)
	if s.ExitStatus() != 0 {
		return s
	}
	bc := fmt.Sprintf("bash -c '%s'", expando(cmd, s))
	log.Printf("DEBUG: %s", bc)
	_, err := script.Exec(bc).Stdout()

	if err != nil {
		s.status = 1
	}
	s.err = err
	return s
}

func expando(cmd string, e any) string {
	temple := template.New("expando")
	temp, err := temple.Parse(cmd)
	if err != nil {
		panic(err)
	}
	var buf bytes.Buffer
	err = temp.Execute(&buf, e)
	if err != nil {
		panic(err)
	}
	return buf.String()
}

type Step struct {
	description string
	status      int
	err         error
	Var         map[string]any
	Flag        map[string]any
	Arg         []string
}

func (s *Step) ExitStatus() int  { return s.status }
func (s *Step) Error() error     { return s.err }
func (s *Step) Describe() string { return s.description }
func (s *Step) Noop() *Step {
	return s
}

func (s *Step) Fail(msg string) *Step {
	s.status = 1
	s.err = fmt.Errorf(msg)
	return s
}

func (s *Step) Call(f func(s *Step)) *Step {
	if s.status == 1 {
		return s
	}
	f(s)
	return s
}
