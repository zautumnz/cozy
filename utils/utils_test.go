package utils

import "testing"

func TestSetReplOrRun(t *testing.T) {
	if IsRepl {
		t.Errorf("IsRepl initialized at true!")
	}

	SetReplOrRun(true)

	if !IsRepl {
		t.Errorf("IsRepl not set to true!")
	}

	SetReplOrRun(false)

	if IsRepl {
		t.Errorf("IsRepl not set to false!")
	}
}
