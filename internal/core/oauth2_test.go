package core_test

import (
	"fmt"
	"testing"

	"github.com/andriihomiak/wallabago/internal/core"
)

func TestNewScope(t *testing.T) {
	cases := []struct {
		scopeNames    []core.ScopeName
		shouldSucceed bool
		expectedValue string
	}{
		{
			scopeNames:    []core.ScopeName{core.ScopeName("entries")},
			shouldSucceed: true,
			expectedValue: "entries",
		},
		{
			scopeNames:    []core.ScopeName{},
			shouldSucceed: true,
			expectedValue: "",
		},
		{
			scopeNames:    []core.ScopeName{core.ScopeName("bad")},
			shouldSucceed: false,
			expectedValue: "",
		},
	}
	for i, testCase := range cases {
		t.Run(fmt.Sprintf("TestNewScope_%d_%#v", i, testCase.scopeNames), func(t *testing.T) {
			result, err := core.NewScope(testCase.scopeNames...)
			if err != nil {
				if testCase.shouldSucceed {
					t.Fatalf("Should succeed without error")
				}
				if result != nil {
					t.Fatalf("Scope should be nil in case of an error")
				}
			} else {
				if result == nil {
					t.Fatalf("Scope should not be nil in case of no error")
				}
				if *result != core.Scope(testCase.expectedValue) {
					t.Fatalf("Expected %#v but got %#v", testCase.expectedValue, *result)
				}
			}
		})
	}
}
