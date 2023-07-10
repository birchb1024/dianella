package dianella

import (
	"flag"
	"fmt"
	"log"

	"os"
)

type Stepper interface {
	AND(string) Stepper
	After()
	Bash(command string) Stepper
	Before(...any)
	CONTINUE(string) Stepper
	ContinueOnError(bool) Stepper
	Call(func(Stepper) Stepper) Stepper
	END() Stepper
	Expand(template string, outputFileName string) Stepper
	Fail(msg string) Stepper
	FailErr(e error)
	GetArg() []string
	GetDescription() string
	GetErr() error
	GetFlag() map[string]any
	GetStatus() int
	GetStringVar(name string) (string, Stepper)
	GetVar() map[string]any
	Init(Stepper, string)
	IsFailed() bool
	ReadCSV(filename string) (Stepper, RowsOfFields)
	Sbash(cmd string) (string, Stepper)
	Set(variableName string, value any) Stepper
	SetLogger(l *log.Logger)
	Sexpand(cmd string) (string, Stepper)
}

// Step - Struct to hold status of execution steps and variables passed between steps.
type Step struct {
	Arg            []string
	Flag           map[string]any
	Var            map[string]any
	description    string
	err            error
	Self           Stepper
	status         int
	logg           *log.Logger
	continueOnFail bool
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

func (s *Step) SetLogger(l *log.Logger) { s.logg = l }
func (s *Step) After()                  {}
func (s *Step) Before(info ...any) {
	v, ok := s.Var["trace"]
	if !ok {
		return
	}
	if v != true {
		return
	}
	longMessage := fmt.Sprintf("INFO: %-16v", info)
	width, _ := GetIntBinding(s.GetVar(), "trace_length", 80)
	s.logg.Printf(stringTruncate(longMessage, uint(width)))
}

func (s *Step) Init(st Stepper, desc string) {
	s.Self = st
	s.logg = log.Default()
	s.description = desc
	s.Var = map[string]any{}
	s.Flag = map[string]any{}
	s.Arg = flag.Args()
	s.Var["trace"] = true
	flag.VisitAll(func(f *flag.Flag) { s.Flag[f.Name] = f.Value })

}

func (s *Step) ContinueOnError(x bool) Stepper {
	s.continueOnFail = x
	return s
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

// GetIntBinding - look up integer variable, return the value or the alt if not found or not int
func GetIntBinding(symbols map[string]any, name string, alt int) (int, error) {
	va, ok := symbols[name]
	if !ok {
		return alt, fmt.Errorf("missing '%s' variable", name)
	}
	result, ok := va.(int)
	if !ok {
		return alt, fmt.Errorf("'%s' variable not integer: '%v'", name, va)
	}
	return result, nil
}

func (s *Step) GetStringVar(name string) (string, Stepper) {
	dd, ok := s.GetVar()[name]
	if !ok {
		s.FailErr(fmt.Errorf("missing '%s' variable", name))
		return "", s
	}
	return fmt.Sprintf("%s", dd), s
}

func (s *Step) FailErr(e error) {
	s.Self.Before("FailErr", e)
	defer s.Self.After()
	s.err = e
	s.status = 1
	if !s.continueOnFail {
		log.Fatalf("When %s FailErr: %#v", s.description, e)
	}
}
func (s *Step) Fail(msg string) Stepper {
	s.Self.Before("Fail", msg)
	defer s.Self.After()
	s.status = 1
	s.err = fmt.Errorf(msg)
	if !s.continueOnFail {
		log.Fatalf("When %s Fail: %s", s.description, msg)
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
		logg:        log.Default(),
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
		s.logg.Printf("INFO: CONTINUE ignoring '%s' failure with status %d, %s", s.Self.GetDescription(), s.Self.GetStatus(), s.Self.GetErr())
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
		s.logg.Printf("ERROR: END '%s' failed with status %d, %s", s.Self.GetDescription(), s.Self.GetStatus(), s.Self.GetErr())
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
