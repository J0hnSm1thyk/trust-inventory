package main

import "testing"

func TestAdvisoryRules(t *testing.T) {
	tests := []struct {
		name, version, want string
	}{
		{"pip", "24.0", "update-required"},
		{"pip", "26.2.2", "safe-by-version"},
		{"uv", "0.11.25", "update-required"},
		{"go", "1.27.0", "update-required"},
		{"go", "1.27.3", "safe-by-version"},
		{"go", "1.26.10", "safe-by-version"},
		{"node", "24.19.0", "update-required"},
		{"node", "24.22.1", "safe-by-version"},
	}
	for _, test := range tests {
		component := Component{Name: test.name, Version: test.version, Status: "inventory-only"}
		assessComponent(&component)
		if component.Status != test.want {
			t.Errorf("%s %s: got %s, want %s", test.name, test.version, component.Status, test.want)
		}
	}
}
