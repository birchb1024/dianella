package dianella

import (
	"fmt"
	"testing"
)

func TestIntMin(t *testing.T) {
	//t.Parallel()
	testTable := []struct{ a, b, expected int }{
		{0, 0, 0},
		{1, 2, 1},
		{2, 1, 1},
		{-1, 1, -1},
		{2, -1, -1},
		{20, 21, 20},
		{20, -21, -21},
	}
	for _, item := range testTable {
		t.Run(fmt.Sprintf("%+v", item), func(t *testing.T) {
			actual := intMin(item.a, item.b)
			if item.expected != actual {
				t.Logf("failed intMin(%d, %d) expected %d but got %d", item.a, item.b, item.expected, actual)
				t.Fail()
			}
		})
	}

}

func TestStringTruncate(t *testing.T) {
	t.Parallel()
	testTable := []struct {
		txt      string
		width    uint
		expected string
	}{
		{"", 0, ""},
		{"x", 0, "..."},
		{"xyz", 0, "..."},
		{"a", 1, "a"},
		{"a", 2, "a"},
		{"abc", 1, "a..."},
		{"abc", 2, "ab..."},
	}
	for _, item := range testTable {
		t.Run(fmt.Sprintf("%#v", item), func(t *testing.T) {
			actual := stringTruncate(item.txt, item.width)
			if item.expected != actual {
				t.Logf("failed stringTruncate(%s, %d) expected '%s' but got '%s'", item.txt, item.width, item.expected, actual)
				t.Fail()
			}
		})
	}
}
