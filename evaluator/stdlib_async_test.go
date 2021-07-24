package evaluator

import (
	"testing"
	"time"
)

func TestAsyncAwat(t *testing.T) {
	v := Async(func() interface{} {
		return func() string {
			time.Sleep(1 * time.Second)
			return "foo"
		}()
	})

	x := v.Await()
	expected := "foo"
	if x != expected {
		t.Errorf("async test failed! got %q, wanted %q", x, expected)
	}
}
