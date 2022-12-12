package dianella

import (
	"flag"
	"fmt"
	"log"
	"os"
)

type Stepper interface {
	After()
	AND(string) Stepper
	Before(...any)
	Call(func(Stepper) Stepper) Stepper
	CONTINUE(string) Stepper
	END() Stepper
	Expand(template string, outputFileName string) Stepper
	Fail(msg string) Stepper
	FailErr(e error)
	Init(Stepper, string)
	IsFailed() bool
	Set(variableName string, value any) Stepper
	Sexpand(cmd string) (string, Stepper)

	Bash(command string) Stepper
	Sbash(cmd string) (string, Stepper)

	ReadCSV(filename string) (*Step, RowsOfFields)

	GetDescription() string
	GetErr() error
	GetArg() []string
	GetFlag() map[string]any
	GetVar() map[string]any
	GetStringVar(name string) (string, Stepper)
	GetStatus() int
}

// Step - Struct to hold status of execution steps and variables passed between steps.
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

// GetDescription - return the current step description
func (s *Step) GetDescription() string { return s.description }
func (s *Step) GetErr() error          { return s.err }
func (s *Step) GetStatus() int         { return s.status }
func (s *Step) IsFailed() bool {
	return s.Self.GetStatus() != 0 || s.Self.GetErr() != nil
}

func (s *Step) After() {}
func (s *Step) Before(info ...any) {
	v, ok := s.Var["trace"]
	if !ok {
		return
	}
	if v != true {
		return
	}
	longMessage := fmt.Sprintf("INFO: %-16v", info)
	log.Printf(stringTruncate(longMessage, 80))
}

func (s *Step) Init(st Stepper, desc string) {
	s.Self = st
	s.description = desc
	s.Var = map[string]any{}
	s.Flag = map[string]any{}
	s.Arg = flag.Args()
	s.Var["trace"] = true
	flag.VisitAll(func(f *flag.Flag) { s.Flag[f.Name] = f.Value })

}

func (s *Step) Set(name string, value any) Stepper {
	if s.Self.IsFailed() {
		return s
	}
	s.Self.Before("Set", name, value)
	defer s.Self.After()
	s.Var[name] = value
	if sv, ok := value.(string); ok {
		ex, err := Expando(sv, s)
		if err != nil {
			s.FailErr(err)
			return s
		}
		s.Var[name] = ex
	}
	return s
}

func (s *Step) GetStringVar(name string) (string, Stepper) {
	dd, ok := s.GetVar()[name]
	if !ok {
		s.Fail("missing dataDir variable")
		return "", s
	}
	return fmt.Sprintf("%s", dd), s
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
func (s *Step) AND(desc string) Stepper {
	if s.Self.IsFailed() {
		return s
	}
	s.Self.Before("AND", desc)
	defer s.Self.After()
	s.description = desc
	return s
}
func (s *Step) CONTINUE(desc string) Stepper {
	if s.Self.IsFailed() {
		log.Printf("INFO: CONTINUE ignoring '%s' failure with status %d, %s", s.Self.GetDescription(), s.Self.GetStatus(), s.Self.GetErr())
	}
	s.Self.Before("CONTINUE", desc)
	defer s.Self.After()
	s.description = desc
	s.status = 0
	s.err = nil
	return s
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

func (s *Step) Call(f func(s Stepper) Stepper) Stepper {
	if s.Self.IsFailed() {
		return s
	}
	s.Self.Before("Call")
	defer s.Self.After()
	f(s.Self)
	return s
}
