package main

import "fmt"

type fixedRule struct {
	Version string
	Source  string
}

var fixedRules = map[string]fixedRule{
	"pip":     {Version: "26.2.2", Source: "pip 26.2.2 bundles the corrected CA data"},
	"certifi": {Version: "2026.10.19", Source: "certifi 2026.10.19 removes the affected trust anchors"},
	"uv":      {Version: "0.13.4", Source: "uv 0.13.4 updates bundled Mozilla roots"},
	"pixi":    {Version: "0.82.1", Source: "first listed patched pixi release"},
	"rustup":  {Version: "1.29.2", Source: "first listed patched rustup release"},
	"deno":    {Version: "2.9.8", Source: "first listed patched Deno release"},
	"mise":    {Version: "2026.10.17", Source: "first listed patched mise release"},
	"claude":  {Version: "2.1.303", Source: "first listed patched Claude Code release"},
}

func assessComponent(component *Component) {
	if component.Version == "" {
		if component.Status == "" {
			component.Status = "unknown"
		}
		return
	}
	if rule, ok := fixedRules[component.Name]; ok {
		component.FixedVersion = rule.Version
		component.Evidence = rule.Source
		if compareVersions(component.Version, rule.Version) >= 0 {
			component.Status = "safe-by-version"
		} else {
			component.Status = "update-required"
		}
		return
	}
	switch component.Name {
	case "go":
		assessGo(component)
	case "node":
		assessNode(component)
	case "java":
		assessJava(component)
	case "bun":
		component.Status = "unknown"
		component.Evidence = "the incident page still lists Bun remediation as pending"
	}
}

func assessGo(component *Component) {
	major, minor, ok := majorMinor(component.Version)
	component.FixedVersion = "1.26.10 or 1.27.3"
	component.Evidence = "CVE-2026-111943; Go security release recommends immediate upgrade"
	if !ok || major != 1 {
		component.Status = "unknown"
		return
	}
	switch {
	case minor > 27:
		component.Status = "safe-by-version"
	case minor == 27:
		component.Status = statusAtLeast(component.Version, "1.27.3")
	case minor == 26:
		component.Status = statusAtLeast(component.Version, "1.26.10")
	default:
		component.Status = "update-required"
	}
}

func assessNode(component *Component) {
	numbers := versionNumbers(component.Version)
	component.FixedVersion = "22.23.4 / 24.22.1 / 26.12.1"
	component.Evidence = "Node.js October 2026 security releases update root certificates"
	if len(numbers) == 0 {
		component.Status = "unknown"
		return
	}
	switch numbers[0] {
	case 22:
		component.Status = statusAtLeast(component.Version, "22.23.4")
	case 24:
		component.Status = statusAtLeast(component.Version, "24.22.1")
	case 26:
		component.Status = statusAtLeast(component.Version, "26.12.1")
	default:
		if numbers[0] > 26 {
			component.Status = "unknown"
		} else {
			component.Status = "update-required"
		}
	}
}

func assessJava(component *Component) {
	numbers := versionNumbers(component.Version)
	if len(numbers) == 0 {
		component.Status = "unknown"
		return
	}
	major := numbers[0]
	if major == 1 && len(numbers) > 1 {
		major = numbers[1]
	}
	thresholds := map[int]string{
		27: "27.0.1.0.2",
		25: "25.0.5.0.2",
		21: "21.0.13.0.2",
		17: "17.0.21.0.2",
		11: "11.0.33.0.2",
		8:  "8u511",
	}
	fixed, ok := thresholds[major]
	if !ok {
		component.Status = "unknown"
		component.Evidence = "no patched threshold is recorded for this Java release line"
		return
	}
	component.FixedVersion = fixed
	component.Evidence = fmt.Sprintf("patched threshold for the Java %d release line", major)
	if major == 8 {
		match := java8VersionPattern.FindStringSubmatch(component.Version)
		if len(match) == 2 && compareVersions(match[1], "511") >= 0 {
			component.Status = "safe-by-version"
		} else {
			component.Status = "update-required"
		}
		return
	}
	component.Status = statusAtLeast(component.Version, fixed)
}

func statusAtLeast(current, fixed string) string {
	if compareVersions(current, fixed) >= 0 {
		return "safe-by-version"
	}
	return "update-required"
}
