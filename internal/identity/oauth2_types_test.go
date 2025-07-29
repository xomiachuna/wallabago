package identity_test

import (
	"fmt"
	"testing"

	"github.com/andriihomiak/wallabago/internal/identity"
)

func TestNewScope(t *testing.T) {
	cases := []struct {
		scopeNames    []identity.ScopeName
		shouldSucceed bool
		expectedValue string
	}{
		{
			scopeNames:    []identity.ScopeName{identity.ScopeName("entries")},
			shouldSucceed: true,
			expectedValue: "entries",
		},
		{
			scopeNames:    []identity.ScopeName{},
			shouldSucceed: true,
			expectedValue: "",
		},
		{
			scopeNames:    []identity.ScopeName{identity.ScopeName("bad")},
			shouldSucceed: false,
			expectedValue: "",
		},
	}
	for i, testCase := range cases {
		t.Run(fmt.Sprintf("TestNewScope_%d_%#v", i, testCase.scopeNames), func(t *testing.T) {
			result, err := identity.NewScope(testCase.scopeNames...)
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
				if *result != identity.Scope(testCase.expectedValue) {
					t.Fatalf("Expected %#v but got %#v", testCase.expectedValue, *result)
				}
			}
		})
	}
}
