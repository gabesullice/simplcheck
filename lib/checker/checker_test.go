package checker

import (
	"fmt"
	"net/http"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	conf := Config{
		Settings:     Settings{"5s"},
		Applications: []string{},
	}

	checker := Checker{}
	checker.LoadConfig(conf)
}

func TestCheck(t *testing.T) {
	checks := []struct {
		url          string
		returnCode   int
		applications []string
		expectState  string
		expectChecks uint
		expectError  bool
	}{
		{"google.com", 200, []string{"google.com"}, "passing", 1, false},
		{"google.com", 500, []string{"google.com"}, "failing", 1, false},
		{"google.com", 200, []string{"aten.io"}, "passing", 1, true},
	}

	for _, check := range checks {
		getter := MockGetter{http.Response{StatusCode: check.returnCode}, nil}

		checker := NewChecker(UseClient(getter))
		checker.LoadConfig(Config{Applications: check.applications})

		status, err := checker.Check(check.url)
		if err != nil {
			if !check.expectError {
				t.Error("Unexpected error")
			} else {
				msg := fmt.Sprintf("Cannot check %s. No associated configuration.", check.url)
				if err.Error() != msg {
					t.Error("Incorrect error format")
				}
			}
		} else {
			if status.state != check.expectState {
				t.Errorf("Expected %s state, got: %s", check.expectState, status.state)
			}
			if status.checks != check.expectChecks {
				t.Errorf("Expected %d checks, got: %d", check.expectChecks, status.checks)
			}
		}

	}

}

type MockGetter struct {
	resp http.Response
	err  error
}

func (m MockGetter) Get(string) (resp *http.Response, err error) {
	if m.err != nil {
		return nil, m.err
	}
	return &m.resp, nil
}
