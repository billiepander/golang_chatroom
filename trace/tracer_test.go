package trace

import (
	"testing"
	"bytes"
)

//The Go tools will treat any function that starts with Test (taking a single *testing.T argument) as a unit test
func TestNew(t *testing.T) {
	var buf bytes.Buffer
	tracer := New(&buf)
	if tracer == nil {
		t.Error("Return from New should not be nil")
	} else {
		tracer.Trace("Hello trace package.")
		if buf.String() != "Hello trace package.\n" {
			t.Errorf("Trace should not write '%s'.", buf.String())
		}
	}
}

//a trace.Off() method that will return an object that satisfies the Tracer interface but will not do anything when the Trace method is called.
func TestOff(t *testing.T) {
	var silentTracer Tracer = Off()
	silentTracer.Trace("something")
}