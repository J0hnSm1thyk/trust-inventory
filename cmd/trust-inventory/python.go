package main

import (
	"context"
	"regexp"
	"runtime"
	"strings"
	"time"
)

var windowsPythonPath = regexp.MustCompile(`(?i)([A-Z]:\\[^\r\n]*?python(?:3)?\.exe)\s*$`)

func inspectPythonInstallations(ctx context.Context, timeout time.Duration) []Component {
	paths := map[string]string{}
	for _, name := range []string{"python", "python3"} {
		for _, path := range findExecutablePaths(name) {
			paths[pathIdentity(cleanPath(path))] = "PATH"
		}
	}
	if runtime.GOOS == "windows" {
		for _, launcher := range findExecutablePaths("py") {
			output, _ := runCommand(ctx, timeout, launcher, "-0p")
			for _, line := range strings.Split(output, "\n") {
				match := windowsPythonPath.FindStringSubmatch(strings.TrimSpace(line))
				if len(match) == 2 {
					paths[pathIdentity(cleanPath(match[1]))] = "py launcher"
				}
			}
		}
	}

	var result []Component
	for path, source := range paths {
		versionOutput, err := runCommand(ctx, timeout, path, "--version")
		if err != nil {
			continue // Ignore Microsoft Store aliases and broken shims.
		}
		result = append(result, Component{
			Name: "python", Category: "python", Path: cleanPath(path), Source: source,
			Version: extractVersion("python", versionOutput), Status: "inventory-only", Evidence: oneLine(versionOutput, 300),
		})

		pipOutput, pipErr := runCommand(ctx, timeout, path, "-m", "pip", "--version")
		if pipErr == nil || strings.Contains(strings.ToLower(pipOutput), "pip ") {
			result = append(result, Component{
				Name: "pip", Category: "python-package", Path: cleanPath(path), Source: source + " (-m pip)",
				Version: extractVersion("pip", pipOutput), Status: "inventory-only", Evidence: oneLine(pipOutput, 300),
			})
		}

		certifiScript := `import importlib.util; s=importlib.util.find_spec("certifi"); print("not-installed" if s is None else __import__("certifi").__version__)`
		certifiOutput, certifiErr := runCommand(ctx, timeout, path, "-c", certifiScript)
		if certifiErr == nil && strings.TrimSpace(certifiOutput) != "not-installed" {
			result = append(result, Component{
				Name: "certifi", Category: "python-package", Path: cleanPath(path), Source: source + " (import certifi)",
				Version: extractVersion("certifi", certifiOutput), Status: "inventory-only", Evidence: oneLine(certifiOutput, 300),
			})
		}
	}
	return result
}
