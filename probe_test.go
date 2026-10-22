package main

import (
	"errors"
	"testing"
)

func TestClassifyProbe(t *testing.T) {
	tests := []struct {
		name, output string
		err          error
		want         string
	}{
		{name: "accepted", output: "CONNECTED 200", want: "vulnerable"},
		{name: "certificate rejected", output: "certificate verify failed", err: errors.New("exit 1"), want: "safe"},
		{name: "traceback source is not success", output: `print("CONNECTED", r.status) certificate verify failed`, err: errors.New("exit 1"), want: "safe"},
		{name: "network failure", output: "connection timed out", err: errors.New("exit 1"), want: "unknown"},
	}
	for _, test := range tests {
		probe := classifyProbe("test", "", "target", test.output, test.err)
		if probe.Status != test.want {
			t.Errorf("%s: got %s, want %s", test.name, probe.Status, test.want)
		}
	}
}
