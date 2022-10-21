package promise_test

import (
	"fmt"
	"testing"
)

type (
	inputs struct {
		val string
		err error
	}
)

var (
	testCases = map[string]inputs{
		"zeroVal":    {},
		"valNoError": {"test", nil},
		"noValError": {"", fmt.Errorf("some error")},
		"valError":   {"some value", fmt.Errorf("some error")},
	}

	noErrorTestCases = map[string]string{
		"zeroVal": "",
		"val":     "test",
	}
)

func expect(t *testing.T, expected, actual any) {
	if expected != actual {
		t.Errorf("expected %v: got %v", expected, actual)
	}
}

// actual test cases are located in promiseme_test and promiseyou_test to keep file sizes down
