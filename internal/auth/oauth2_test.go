package auth_test

import (
	"fmt"
	"testing"

	"github.com/andriihomiak/wallabago/internal/auth"
)

func TestNewScope(t *testing.T) {
	cases := []struct {
		scopeNames    []auth.ScopeName
		shouldSucceed bool
		expectedValue string
	}{
		{
			scopeNames:    []auth.ScopeName{auth.ScopeName("entries")},
			shouldSucceed: true,
			expectedValue: "entries",
		},
		{
			scopeNames:    []auth.ScopeName{},
			shouldSucceed: true,
			expectedValue: "",
		},
		{
			scopeNames:    []auth.ScopeName{auth.ScopeName("bad")},
			shouldSucceed: false,
			expectedValue: "",
		},
	}
	for i, testCase := range cases {
		t.Run(fmt.Sprintf("TestNewScope_%d_%#v", i, testCase.scopeNames), func(t *testing.T) {
			result, err := auth.NewScope(testCase.scopeNames...)
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
				if *result != auth.Scope(testCase.expectedValue) {
					t.Fatalf("Expected %#v but got %#v", testCase.expectedValue, *result)
				}
			}
		})
	}
}
