package dianella

import (
	"bytes"
	"flag"
	"io"
	"log"
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
	var s Stepper = BEGIN("Basic example").ContinueOnError(true)
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
	var s Stepper = BEGIN("Basic example").ContinueOnError(true)
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
	//t.Parallel()
	flag.Parse()
	given := struct{ time string }{time.Now().String()}
	s := BEGIN("Bad template").ContinueOnError(true).
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
	s := BEGIN("Bad template").ContinueOnError(true).
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

func TestGetIntBinding(t *testing.T) {
	t.Parallel()

	for name, plot := range map[string]struct {
		given  any
		alt    int
		expect int
	}{
		"no variable": {given: nil, alt: 1, expect: 1},
		"not int":     {given: "ZZ", alt: 1, expect: 1},
		"normal":      {given: 42, alt: 0, expect: 42},
		"zero":        {given: 0, alt: 23, expect: 0},
		"one":         {given: 1, alt: 0, expect: 1},
		"big":         {given: 424242, alt: 0, expect: 424242},
	} {
		t.Run(name, func(t *testing.T) {
			s := BEGIN(t.Name()).ContinueOnError(true)
			if plot.given != nil {
				s.Set("trace-length", plot.given)
			}
			actual, _ := GetIntBinding(s.GetVar(), "trace-length", plot.alt)
			if actual != plot.expect {
				t.Logf("plot was %v, but got %d", plot, actual)
				t.Fail()
			}
		})
	}
}

func TestTrace(t *testing.T) {
	t.Parallel()

	for name, plot := range map[string]struct {
		given  any
		expect string
	}{
		"no variable": {given: nil, expect: "INFO: [Bash             echo '{{.Var.trace_length }}' 1 2 3 4 5 6 7 8 9 10 11 12"},
		"tiny":        {given: 3, expect: "\nINF...\n"},
		"short":       {given: 33, expect: "\nINFO: [Bash.cmd         echo '33'...\n"},
		"medium":      {given: 50, expect: "\nINFO: [Bash.cmd         echo '50' 1 2 3 4 5 6 7 8 ..."},
		"huge":        {given: 2000, expect: "\nINFO: [Bash.cmd         echo '2000' 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17 18 19 20 >/dev/null]\n"},
	} {
		t.Run(name, func(t *testing.T) {
			s := BEGIN(t.Name()).ContinueOnError(true)
			var b bytes.Buffer
			lg := log.New(io.Writer(&b), "", 0)
			s.SetLogger(lg)
			if plot.given != nil {
				s.Set("trace_length", plot.given)
			}

			s.Bash("echo '{{.Var.trace_length }}' 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17 18 19 20 >/dev/null")
			t.Logf(b.String())
			if !strings.Contains(b.String(), plot.expect) {
				t.Logf("plot was %v, but got '%s'", plot, b.String())
				t.Fail()
			}
		})
	}
}
