package dianella

import (
	"flag"
	"strings"
	"testing"
	"time"
)

func TestBasicsPass(t *testing.T) {

	mock := func(s Stepper) Stepper {
		t.Logf("description %v\n", s.GetDescription())
		t.Logf("arg %v\n", s.GetArg())
		t.Logf("flag %v\n", s.GetFlag())
		t.Logf("var %v\n", s.GetVar())
		return s
	}

	t.Parallel()
	flag.Parse()
	given := struct{ time string }{time.Now().String()}
	var s Stepper = BEGIN("Basic example")
	s.Init(s, "Init")
	s.Set("trace", false).
		Set("date", given.time).
		Call(mock).
		AND("bash date").
		Bash("date").
		Bash("echo {{.Var.date}}")
	tmpFile, s := s.Sbash("mktemp")
	s.Set("tmpFile", strings.TrimSpace(tmpFile)).
		Expand("tmpFile: {{.Var.tmpFile}} - Date: {{.Var.date}}\n", strings.TrimSpace(tmpFile)).
		Bash("cat {{.Var.tmpFile}}").
		Bash("rm -f {{.Var.tmpFile}}").
		Bash("false").
		CONTINUE("ignore failure which is expected").
		END()
	if s.IsFailed() {
		t.Log(s.GetErr())
		t.Fail()
	}
	actual, s := s.GetStringVar("date")
	if actual != given.time {
		t.Errorf("date variable discrepency, expected: '%v', got: '%v'", given.time, actual)
	}

}

func TestBasicsEarlyFail(t *testing.T) {
	t.Parallel()
	flag.Parse()
	given := struct{ time string }{time.Now().String()}
	var s Stepper = BEGIN("Basic example")
	s.Init(s, "Init")
	s.Set("trace", true).
		Set("date", given.time).
		Set("dummy", "template failure {{").
		Call(func(stepper Stepper) Stepper { return s }).
		AND("bash date").
		Bash("date").
		Bash("echo {{.Var.date}}")
	tmpFile, s := s.Sbash("mktemp")
	s, _ = s.ReadCSV("foo")
	_, s = s.Sexpand("{{")
	s.Set("tmpFile", strings.TrimSpace(tmpFile)).
		Expand("tmpFile: {{.Var.tmpFile}} - Date: {{.Var.date}}\n", strings.TrimSpace(tmpFile)).
		Bash("cat {{.Var.tmpFile}}").
		Bash("rm -f {{.Var.tmpFile}}").
		Bash("false")

	if !s.IsFailed() {
		t.Logf("should have failed")
		t.Fail()
	}
	_, s = s.GetStringVar("ZZZZ")
	if !strings.Contains(s.GetErr().Error(), "missing 'ZZZZ' variable") {
		t.Errorf("expected error with missing ZZZZ variable: got: '%s'", s.GetErr().Error())
	}

}

func TestSexpandFails(t *testing.T) {
	t.Parallel()
	flag.Parse()
	given := struct{ time string }{time.Now().String()}
	s := BEGIN("Bad template").
		Set("date", given.time)
	_, s = s.Sexpand("{{")
	t.Log(s.GetErr())
	if !s.IsFailed() {
		t.Logf("should have failed")
		t.Fail()
	}
	actual, s := s.GetStringVar("date")
	if actual != given.time {
		t.Errorf("date variable discrepency, expected: '%v', got: '%v'", given.time, actual)
	}

}

func TestExpandFails(t *testing.T) {
	t.Parallel()
	flag.Parse()
	given := struct{ time string }{time.Now().String()}
	s := BEGIN("Bad template").
		Set("date", given.time).
		Expand("{{", "/tmp/foo")
	t.Log(s.GetErr())
	if !s.IsFailed() {
		t.Logf("should have failed")
		t.Fail()
	}
	actual, s := s.GetStringVar("date")
	if actual != given.time {
		t.Errorf("date variable discrepency, expected: '%v', got: '%v'", given.time, actual)
	}

}
