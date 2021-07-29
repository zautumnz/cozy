package utils

import "testing"

func TestSetReplOrRun(t *testing.T) {
	if isRepl {
		t.Errorf("isRepl initialized at true!")
	}

	SetReplOrRun(true)

	if !isRepl {
		t.Errorf("isRepl not set to true!")
	}

	SetReplOrRun(false)

	if isRepl {
		t.Errorf("isRepl not set to false!")
	}
}
