package dianella

import (
	"testing"
)

func TestBashFailures(t *testing.T) {
	t.Parallel()

	testTable := map[string]string{
		"@prior-step-failed":    "",
		"missing command":       "DOESNOTEXIST",
		"syntax error":          "fi",
		"file not found":        "cat /does/not/exist",
		"template syntax error": "{{ . ",
		"template missing var":  "{{ .missing }}",
		"process failure":       "false",
		"process kill signal":   "kill $$",
	}

	runBash(t, testTable, false)
}

func TestBashPasses(t *testing.T) {
	t.Parallel()

	testTable := map[string]string{
		"command":      "true",
		"file found":   "echo foo >/dev/null",
		"template var": "echo {{ .Var.testName }}",
	}

	runBash(t, testTable, true)
}

func TestSbashPasses(t *testing.T) {
	t.Parallel()

	testTable := map[string]struct{ cmd, expected string }{
		"echo":                {"echo -n Hello,World", "Hello,World"},
		"command":             {"true", ""},
		"command with output": {"ls -1d /", "/\n"},
		"file found":          {"echo foo >/dev/null", ""},
		"template variable":   {"echo -n {{ .Var.testName }}", "template variable"},
	}

	runSbash(t, testTable, true)
}

func TestSbashFailures(t *testing.T) {
	t.Parallel()

	testTable := map[string]struct{ cmd, expected string }{
		"@prior-step-failed":    {"", ""},
		"missing command":       {"DOESNOTEXIST", ""},
		"syntax error":          {"fi", ""},
		"file not found":        {"cat /does/not/exist", ""},
		"template syntax error": {"{{ . ", ""},
		"template missing var":  {"{{ .missing }}", ""},
		"process failure":       {"false", ""},
		"process kill signal":   {"kill $$", ""},
	}

	runSbash(t, testTable, false)
}

func runSbash(t *testing.T, testTable map[string]struct{ cmd, expected string }, pass bool) {
	t.Helper()
	for name, scenario := range testTable {
		t.Run(name, func(t *testing.T) {
			s := BEGIN(name).ContinueOnError(true).
				Set("testName", name)
			if name == "@prior-step-failed" {
				s.Fail(name)
			}
			actual, s := s.Sbash(scenario.cmd)
			if s.IsFailed() == pass {
				t.Logf("expected isFailed() %v did not occur on '%s'", pass, scenario)
				t.Fail()
			}
			if actual != scenario.expected {
				t.Logf("from '%s' expected '%s' but got '%s'", scenario.cmd, scenario.expected, actual)
				t.Fail()
			}
		})
	}
}

func runBash(t *testing.T, testTable map[string]string, pass bool) {
	t.Helper()
	for name, cmd := range testTable {
		t.Run(name, func(t *testing.T) {
			s := BEGIN(name).ContinueOnError(true).
				Set("testName", name)
			if name == "@prior-step-failed" {
				s.Fail(name)
			}
			s.Bash(cmd)
			if s.IsFailed() == pass {
				t.Logf("expected isFailed() %v did not occur on '%s'", pass, cmd)
				t.Fail()
			}
		})
	}
}
