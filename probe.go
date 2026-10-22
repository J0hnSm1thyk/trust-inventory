package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type probeTarget struct {
	Name string
	URL  string
}

var emudhraTargets = []probeTarget{
	{Name: "direct", URL: "https://dv-g3b-emudhra.we-have-a-problem.com/"},
	{Name: "cross-signed", URL: "https://dv-g3b-aaa-emudhra.we-have-a-problem.com/"},
}

func runEMudhraProbes(ctx context.Context, components []Component, timeout time.Duration) []Probe {
	probes := probeBuiltinGo(ctx, timeout)
	for _, component := range components {
		switch component.Name {
		case "curl":
			for _, target := range emudhraTargets {
				output, err := runCommand(ctx, timeout, component.Path,
					"--silent", "--show-error", "--head", "--connect-timeout", "5", "--max-time", "8", target.URL)
				probes = append(probes, classifyProbe("curl", component.Path, target.Name, output, err))
			}
		case "node":
			script := `const https=require("https");const u=process.argv[1];const r=https.get(u,x=>{console.log("CONNECTED "+x.statusCode);x.resume()});r.setTimeout(5000,()=>r.destroy(new Error("timeout")));r.on("error",e=>{console.error("REJECTED "+(e.code||"")+" "+e.message);process.exitCode=2})`
			for _, target := range emudhraTargets {
				output, err := runCommand(ctx, timeout, component.Path, "-e", script, target.URL)
				probes = append(probes, classifyProbe("node", component.Path, target.Name, output, err))
			}
		case "python":
			script := `import sys,urllib.request; r=urllib.request.urlopen(sys.argv[1],timeout=5); print("CONNECTED",r.status)`
			for _, target := range emudhraTargets {
				output, err := runCommand(ctx, timeout, component.Path, "-c", script, target.URL)
				probes = append(probes, classifyProbe("python-stdlib", component.Path, target.Name, output, err))
			}
		}
	}
	return probes
}

func probeBuiltinGo(parent context.Context, timeout time.Duration) []Probe {
	transport := &http.Transport{TLSClientConfig: &tls.Config{MinVersion: tls.VersionTLS12}}
	client := &http.Client{Timeout: timeout, Transport: transport}
	var result []Probe
	for _, target := range emudhraTargets {
		request, err := http.NewRequestWithContext(parent, http.MethodHead, target.URL, nil)
		if err != nil {
			result = append(result, Probe{Client: "scanner-go", Target: target.Name, Status: "unknown", Detail: err.Error()})
			continue
		}
		response, err := client.Do(request)
		if err == nil {
			_ = response.Body.Close()
			result = append(result, Probe{Client: "scanner-go", Target: target.Name, Status: "vulnerable", Detail: fmt.Sprintf("connection accepted (HTTP %d)", response.StatusCode)})
			continue
		}
		result = append(result, classifyProbe("scanner-go", "", target.Name, err.Error(), err))
	}
	return result
}

func classifyProbe(client, path, target, output string, err error) Probe {
	detail := oneLine(output, 500)
	if err == nil {
		return Probe{Client: client, Path: path, Target: target, Status: "vulnerable", Detail: detail}
	}
	lower := strings.ToLower(output + " " + err.Error())
	certificateIndicators := []string{
		"certificate", "cert_", "crypt_e_revoked", "unknownissuer", "unknown issuer",
		"unable_to_get_issuer", "unable to get local issuer", "self signed", "schannel",
	}
	for _, indicator := range certificateIndicators {
		if strings.Contains(lower, indicator) {
			return Probe{Client: client, Path: path, Target: target, Status: "safe", Detail: detail}
		}
	}
	return Probe{Client: client, Path: path, Target: target, Status: "unknown", Detail: detail}
}
