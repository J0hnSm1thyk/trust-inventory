package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"runtime"
	"sort"
	"strings"
	"time"
)

const scannerVersion = "0.1.0"

type Report struct {
	SchemaVersion  string      `json:"schema_version"`
	ScannerVersion string      `json:"scanner_version"`
	GeneratedAt    string      `json:"generated_at"`
	OS             string      `json:"os"`
	Architecture   string      `json:"architecture"`
	Hostname       string      `json:"hostname,omitempty"`
	Advisory       Advisory    `json:"advisory"`
	Components     []Component `json:"components"`
	Probes         []Probe     `json:"probes,omitempty"`
}

type Advisory struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Published string `json:"published"`
	URL       string `json:"url"`
}

type Component struct {
	Name         string `json:"name"`
	Category     string `json:"category"`
	Version      string `json:"version,omitempty"`
	Path         string `json:"path,omitempty"`
	Source       string `json:"source,omitempty"`
	Status       string `json:"status"`
	FixedVersion string `json:"fixed_version,omitempty"`
	Evidence     string `json:"evidence,omitempty"`
	Error        string `json:"error,omitempty"`
}

type Probe struct {
	Client string `json:"client"`
	Path   string `json:"path,omitempty"`
	Target string `json:"target"`
	Status string `json:"status"`
	Detail string `json:"detail,omitempty"`
}

func main() {
	format := flag.String("format", "table", "output format: table or json")
	output := flag.String("output", "", "write the report to a file instead of stdout")
	includeHostname := flag.Bool("include-hostname", false, "include the machine hostname in the report")
	probeEMudhra := flag.Bool("probe-emudhra", false, "contact the public eMudhra test hosts (off by default)")
	failOnVulnerable := flag.Bool("fail-on-vulnerable", false, "exit with status 2 when an update-required or vulnerable result is found")
	timeout := flag.Duration("timeout", 6*time.Second, "timeout for each local command or network probe")
	flag.Parse()

	if *format != "table" && *format != "json" {
		fatalf("unsupported format %q; use table or json", *format)
	}

	ctx := context.Background()
	components := collectInventory(ctx, *timeout)
	for i := range components {
		assessComponent(&components[i])
	}
	var probes []Probe
	if *probeEMudhra {
		// Probe before redacting executable paths; redaction is an output concern.
		probes = runEMudhraProbes(ctx, components, *timeout)
	}
	for i := range components {
		components[i].Path = redactHome(components[i].Path)
		components[i].Evidence = redactText(components[i].Evidence)
		components[i].Error = redactText(components[i].Error)
	}
	sort.Slice(components, func(i, j int) bool {
		if components[i].Category != components[j].Category {
			return components[i].Category < components[j].Category
		}
		if components[i].Name != components[j].Name {
			return components[i].Name < components[j].Name
		}
		return components[i].Path < components[j].Path
	})

	report := Report{
		SchemaVersion:  "1",
		ScannerVersion: scannerVersion,
		GeneratedAt:    time.Now().UTC().Format(time.RFC3339),
		OS:             runtime.GOOS,
		Architecture:   runtime.GOARCH,
		Advisory: Advisory{
			ID:        "CVE-2026-111943",
			Title:     "eMudhra intermediate CA compromise",
			Published: "2026-10-20",
			URL:       "https://emudhra.we-have-a-problem.com/",
		},
		Components: components,
	}
	if *includeHostname {
		report.Hostname, _ = os.Hostname()
	}
	if *probeEMudhra {
		report.Probes = probes
		for i := range report.Probes {
			report.Probes[i].Path = redactHome(report.Probes[i].Path)
			report.Probes[i].Detail = redactText(report.Probes[i].Detail)
		}
	}

	var data []byte
	var err error
	if *format == "json" {
		data, err = json.MarshalIndent(report, "", "  ")
		if err == nil {
			data = append(data, '\n')
		}
	} else {
		data = []byte(formatTable(report))
	}
	if err != nil {
		fatalf("format report: %v", err)
	}
	if *output != "" {
		if err := os.WriteFile(*output, data, 0o600); err != nil {
			fatalf("write report: %v", err)
		}
		fmt.Fprintf(os.Stderr, "report written to %s\n", *output)
	} else {
		_, _ = os.Stdout.Write(data)
	}

	if *failOnVulnerable && hasRisk(report) {
		os.Exit(2)
	}
}

func hasRisk(report Report) bool {
	for _, component := range report.Components {
		if component.Status == "vulnerable" || component.Status == "update-required" {
			return true
		}
	}
	for _, probe := range report.Probes {
		if probe.Status == "vulnerable" {
			return true
		}
	}
	return false
}

func formatTable(report Report) string {
	var b strings.Builder
	fmt.Fprintf(&b, "trust-inventory %s  %s/%s\n", report.ScannerVersion, report.OS, report.Architecture)
	fmt.Fprintf(&b, "Advisory: %s (%s)\n\n", report.Advisory.Title, report.Advisory.ID)
	fmt.Fprintf(&b, "%-18s %-17s %-18s %-18s %s\n", "COMPONENT", "CATEGORY", "VERSION", "STATUS", "PATH")
	fmt.Fprintf(&b, "%s\n", strings.Repeat("-", 110))
	for _, component := range report.Components {
		fmt.Fprintf(&b, "%-18s %-17s %-18s %-18s %s\n",
			truncate(component.Name, 18), truncate(component.Category, 17), truncate(component.Version, 18),
			truncate(component.Status, 18), component.Path)
		if component.FixedVersion != "" {
			fmt.Fprintf(&b, "  fixed: %s", component.FixedVersion)
			if component.Evidence != "" {
				fmt.Fprintf(&b, "  (%s)", component.Evidence)
			}
			b.WriteByte('\n')
		}
		if component.Error != "" {
			fmt.Fprintf(&b, "  note: %s\n", oneLine(component.Error, 180))
		}
	}
	if len(report.Probes) > 0 {
		b.WriteString("\nNETWORK PROBES (explicitly enabled)\n")
		for _, probe := range report.Probes {
			fmt.Fprintf(&b, "%-18s %-12s %-14s %s\n", probe.Client, probe.Target, probe.Status, oneLine(probe.Detail, 120))
		}
	}
	return b.String()
}

func truncate(value string, width int) string {
	if len(value) <= width {
		return value
	}
	if width < 2 {
		return value[:width]
	}
	return value[:width-1] + "…"
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "trust-inventory: "+format+"\n", args...)
	os.Exit(1)
}
