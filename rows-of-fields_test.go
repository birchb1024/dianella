package dianella

import (
	"fmt"
	"sort"
	"strings"
	"testing"
)

type mapped map[string]map[string]string

func TestRowsOfFields_Rows2Maps(t *testing.T) {
	t.Parallel()

	testTable := map[string]struct {
		keys     []string
		given    RowsOfFields
		expected mapped
		errors   string
	}{
		"not enough rows": {
			keys:   []string{"one"},
			given:  RowsOfFields{{"one"}},
			errors: "expected header and one row",
		},
		"no row": {
			keys:   []string{"one"},
			given:  RowsOfFields{{"one"}, {}},
			errors: "row too short",
		},
		"missing column": {
			keys:   []string{"ZZZ"},
			given:  RowsOfFields{{"one"}, {"A"}},
			errors: "missing key column",
		},
		"one column": {
			keys:     []string{"one"},
			given:    RowsOfFields{{"one"}, {"A"}, {"B"}},
			expected: mapped{"A": {"one": "A"}, "B": {"one": "B"}},
		},
		"table": {
			keys: []string{"one", "three"},
			given: RowsOfFields{
				{"one", "two", "three"},
				{"1", "2", "3"},
				{"4", "5", "6"},
				{"7", "8", "9"},
			},
			expected: mapped{
				"13": {"one": "1",
					"two":   "2",
					"three": "3"},
				"46": {"one": "4",
					"two":   "5",
					"three": "6"},
				"79": {"one": "7",
					"two":   "8",
					"three": "9"},
			},
		},
	}

	for name, plot := range testTable {
		t.Run(name, func(t *testing.T) {
			actual, err := plot.given.Rows2Maps(plot.keys)
			if plot.errors != "" && err == nil {
				t.Fail()
			}
			if err != nil {
				if !strings.Contains(err.Error(), plot.errors) {
					t.Logf("given %v, %v , expected '%v', but got '%v'", plot.keys, plot.given, plot.errors, err.Error())
					t.Fail()
					return
				}
			}
			if fmt.Sprintf("%v", actual) != fmt.Sprintf("%v", plot.expected) {
				t.Logf("given %v, %v , expected '%v', but got '%v'", plot.keys, plot.given, plot.expected, actual)
				t.Fail()
			}
		})
	}
}

func TestRowsOfFields_SelectColumnDistinctValues(t *testing.T) {
	t.Parallel()

	testTable := map[string]struct {
		key      string
		given    RowsOfFields
		expected []string
		errors   string
	}{
		"not enough rows": {
			key:    "one",
			given:  RowsOfFields{{"one"}},
			errors: "expected header and one row",
		},
		"no row": {
			key:      "one",
			given:    RowsOfFields{{"one"}, {}},
			expected: []string{},
		},
		"missing column": {
			key:    "ZZZ",
			given:  RowsOfFields{{"one"}, {"A"}},
			errors: "could not find",
		},
		"one column": {
			key:      "one",
			given:    RowsOfFields{{"one"}, {"A"}, {"B"}},
			expected: []string{"A", "B"},
		},
		"table": {
			key: "one",
			given: RowsOfFields{
				{"one", "two", "three"},
				{"1", "2", "3"},
				{"4", "5", "6"},
				{"7", "8", "9"},
			},
			expected: []string{"1", "4", "7"},
		},
	}

	for name, plot := range testTable {
		t.Run(name, func(t *testing.T) {
			actual, err := plot.given.SelectColumnDistinctValues(plot.key)
			sort.Strings(actual)
			sort.Strings(plot.expected)
			if plot.errors != "" && err == nil {
				t.Logf("given %v, %v , expected '%v', but got '%v'", plot.key, plot.given, plot.errors, actual)
				t.Fail()
				return
			}
			if err != nil {
				if !strings.Contains(err.Error(), plot.errors) {
					t.Logf("given %v, %v , expected '%v', but got '%v'", plot.key, plot.given, plot.errors, err.Error())
					t.Fail()
					return
				}
			}
			if fmt.Sprintf("%v", actual) != fmt.Sprintf("%v", plot.expected) {
				t.Logf("given %v, %v , expected '%v', but got '%v'", plot.key, plot.given, plot.expected, actual)
				t.Fail()
			}
		})
	}
}

func TestRowsOfFields_ReadCSV(t *testing.T) {
	t.Parallel()
	s := BEGIN("Test CSV file reader").ContinueOnError(true)
	_, actual := s.ReadCSV("test/fixture.csv")
	if s.IsFailed() {
		t.Error(s.GetErr())
	}
	expected := `[[one two three four] [1 2 3 4] [5 6 7 8] [9 10 11 12]]`
	if fmt.Sprintf("%v", actual) != expected {
		t.Logf("expected '%v', but got '%v'", expected, actual)
		t.Fail()
	}

}

func TestRowsOfFields_ReadCSV_404(t *testing.T) {
	t.Parallel()
	var s Stepper = BEGIN("Test CSV file reader").ContinueOnError(true)
	s, _ = s.ReadCSV("test/nosuchfile.csv")
	t.Log(s.GetErr())
	if !s.IsFailed() {
		t.Fail()
	}
}

func TestRowsOfFields_ReadCSV_Bad(t *testing.T) {
	t.Parallel()
	var s Stepper = BEGIN("Test CSV file reader").ContinueOnError(true)
	s, _ = s.ReadCSV("test/badfixture.csv")
	t.Log(s.GetErr())
	if !s.IsFailed() {
		t.Fail()
	}
}
