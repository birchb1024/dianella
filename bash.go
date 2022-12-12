package dianella

import (
	"os"
	"os/exec"
)

func (s *Step) Bash(cmd string) Stepper {
	if s.Self.IsFailed() {
		return s
	}
	s.Self.Before("Bash", cmd)
	defer s.Self.After()
	ex, err := Expando(cmd, s)
	if err != nil {
		s.Self.FailErr(err)
		return s
	}
	c := exec.Command("/bin/bash", "-c", ex)
	c.Stderr = os.Stderr
	c.Stdout = os.Stdout
	err = c.Run()
	if err != nil {
		s.Self.FailErr(err)
	}
	return s
}
func (s *Step) Sbash(cmd string) (result string, rs Stepper) {
	if s.Self.IsFailed() {
		return "", s
	}
	s.Self.Before("Sbash", cmd)
	defer s.Self.After()
	ex, err := Expando(cmd, s)
	if err != nil {
		s.Self.FailErr(err)
		return "", s
	}
	c := exec.Command("/bin/bash", "-c", ex)
	c.Stderr = os.Stderr
	stdoutBytes, err := c.Output()
	if err != nil {
		s.Self.FailErr(err)
	}
	return string(stdoutBytes), s
}
