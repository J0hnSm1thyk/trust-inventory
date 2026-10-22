package main

import (
	"regexp"
	"strconv"
	"strings"
)

var (
	goVersionPattern      = regexp.MustCompile(`(?i)\bgo(\d+(?:\.\d+){1,3})\b`)
	genericVersionPattern = regexp.MustCompile(`(?i)\bv?(\d+(?:\.\d+){1,4}(?:[-+][0-9a-z.-]+)?)\b`)
	java8VersionPattern   = regexp.MustCompile(`(?i)\b(?:1\.)?8(?:\.0)?[_u](\d+)\b`)
	numberPattern         = regexp.MustCompile(`\d+`)
)

func extractVersion(name, output string) string {
	if strings.EqualFold(name, "go") {
		if match := goVersionPattern.FindStringSubmatch(output); len(match) == 2 {
			return match[1]
		}
	}
	if strings.EqualFold(name, "java") {
		if match := java8VersionPattern.FindStringSubmatch(output); len(match) == 2 {
			return "8u" + match[1]
		}
	}
	if match := genericVersionPattern.FindStringSubmatch(output); len(match) == 2 {
		return strings.TrimSuffix(match[1], ".")
	}
	return ""
}

func compareVersions(left, right string) int {
	a := versionNumbers(left)
	b := versionNumbers(right)
	length := len(a)
	if len(b) > length {
		length = len(b)
	}
	for i := 0; i < length; i++ {
		av, bv := 0, 0
		if i < len(a) {
			av = a[i]
		}
		if i < len(b) {
			bv = b[i]
		}
		if av < bv {
			return -1
		}
		if av > bv {
			return 1
		}
	}
	return 0
}

func versionNumbers(value string) []int {
	matches := numberPattern.FindAllString(value, -1)
	result := make([]int, 0, len(matches))
	for _, match := range matches {
		number, err := strconv.Atoi(match)
		if err == nil {
			result = append(result, number)
		}
	}
	return result
}

func majorMinor(value string) (int, int, bool) {
	numbers := versionNumbers(value)
	if len(numbers) < 2 {
		return 0, 0, false
	}
	return numbers[0], numbers[1], true
}
