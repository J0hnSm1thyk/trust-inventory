package main

import "testing"

func TestExtractVersion(t *testing.T) {
	tests := []struct {
		name, input, want string
	}{
		{"go", "go version go1.27.3 windows/amd64", "1.27.3"},
		{"node", "v24.22.1", "24.22.1"},
		{"pip", "pip 26.2.2 from C:\\Python\\Lib", "26.2.2"},
		{"uv", "uv 0.13.4 (7456b60 2026-10-20)", "0.13.4"},
		{"java", `java version "1.8.0_511"`, "8u511"},
	}
	for _, test := range tests {
		if got := extractVersion(test.name, test.input); got != test.want {
			t.Errorf("extractVersion(%q, %q) = %q, want %q", test.name, test.input, got, test.want)
		}
	}
}

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		left, right string
		want        int
	}{
		{"1.27.3", "1.27.3", 0},
		{"1.27.2", "1.27.3", -1},
		{"26.2.10", "26.2.2", 1},
		{"2026.10.19", "2026.10.17", 1},
	}
	for _, test := range tests {
		if got := compareVersions(test.left, test.right); got != test.want {
			t.Errorf("compareVersions(%q, %q) = %d, want %d", test.left, test.right, got, test.want)
		}
	}
}
