package dianella

import (
	"bytes"
	"os"
	"text/template"
)

// Expando - Use Go template module to interpolate expansions in a string
// using data from the environment (the SICP sense of environment)
func Expando(templateSource string, environment any) (string, error) {
	temp, err := template.New("Expando").Parse(templateSource)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	err = temp.Execute(&buf, environment)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

// Expand - Using variables in the Step struct, expand the template and output the result to the
// filename provided.
func (s *Step) Expand(template string, filename string) Stepper {
	if s.Self.IsFailed() {
		return s
	}
	s.Self.Before("Expand", template, filename)
	defer s.Self.After()
	expanded, err := Expando(template, s)
	if err != nil {
		s.FailErr(err)
		return s
	}
	err = os.WriteFile(filename, []byte(expanded), 0644)
	if err != nil {
		s.FailErr(err)
	}
	return s
}

// Sexpand - Same as Expand, but return the result as a string
func (s *Step) Sexpand(template string) (string, Stepper) {
	if s.Self.IsFailed() {
		return "", s
	}
	s.Self.Before("Sexpand", template, 20)
	defer s.Self.After()
	ex, err := Expando(template, s)
	if err != nil {
		s.FailErr(err)
	}
	return ex, s
}
