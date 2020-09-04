package jsonp

import (
	"reflect"
	"testing"
)

var testCases = []struct {
	jsp  string
	path []string
}{
	{"", []string{}},
	{"/", []string{""}},
	{"/a", []string{"a"}},
	{"/~0", []string{"~"}},
	{"/~1", []string{"/"}},
	{"/~01", []string{"~1"}},
	{"/a/b", []string{"a", "b"}},
	{"////", []string{"", "", "", ""}},
}

func TestParse(t *testing.T) {
	for i, test := range testCases {
		p, err := Parse(test.jsp)
		if !reflect.DeepEqual(p, test.path) || err != nil {
			t.Errorf("Test %v failed: %q != %q", i, p, test.path)
		}
	}
}

func TestParseInvalid(t *testing.T) {
	if _, err := Parse("a"); err != ErrMissingTokenPrefix {
		t.Error("Expected missing token prefix error")
	}
}

func TestFormat(t *testing.T) {
	for i, test := range testCases {
		if f := Format(test.path); f != test.jsp {
			t.Errorf("Test %v failed: %q != %q", i, f, test.jsp)
		}
	}
}
