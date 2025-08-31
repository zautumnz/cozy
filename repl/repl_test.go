package repl

import (
	"os"
	"testing"
)

func TestGetHistorySize(t *testing.T) {
	v1 := getHistorySize()
	if v1 != 1000 {
		t.Fatalf("expected default history size to be 1000")
	}
	os.Setenv("KEAI_HISTSIZE", "500")
	v2 := getHistorySize()
	if v2 != 500 {
		t.Fatalf("excpted history size to be 500")
	}
}
