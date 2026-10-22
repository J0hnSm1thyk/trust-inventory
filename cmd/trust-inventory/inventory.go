package main

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"
)

type toolSpec struct {
	Name        string
	Category    string
	Executables []string
	Args        []string
}

var toolSpecs = []toolSpec{
	{Name: "winget", Category: "os-package", Executables: []string{"winget"}, Args: []string{"--version"}},
	{Name: "chocolatey", Category: "os-package", Executables: []string{"choco"}, Args: []string{"--version"}},
	{Name: "scoop", Category: "os-package", Executables: []string{"scoop"}, Args: []string{"--version"}},
	{Name: "homebrew", Category: "os-package", Executables: []string{"brew"}, Args: []string{"--version"}},
	{Name: "macports", Category: "os-package", Executables: []string{"port"}, Args: []string{"version"}},
	{Name: "apt", Category: "os-package", Executables: []string{"apt", "apt-get"}, Args: []string{"--version"}},
	{Name: "dnf", Category: "os-package", Executables: []string{"dnf"}, Args: []string{"--version"}},
	{Name: "yum", Category: "os-package", Executables: []string{"yum"}, Args: []string{"--version"}},
	{Name: "pacman", Category: "os-package", Executables: []string{"pacman"}, Args: []string{"--version"}},
	{Name: "apk", Category: "os-package", Executables: []string{"apk"}, Args: []string{"--version"}},
	{Name: "zypper", Category: "os-package", Executables: []string{"zypper"}, Args: []string{"--version"}},
	{Name: "flatpak", Category: "os-package", Executables: []string{"flatpak"}, Args: []string{"--version"}},
	{Name: "snap", Category: "os-package", Executables: []string{"snap"}, Args: []string{"--version"}},
	{Name: "nix", Category: "os-package", Executables: []string{"nix"}, Args: []string{"--version"}},
	{Name: "guix", Category: "os-package", Executables: []string{"guix"}, Args: []string{"--version"}},

	{Name: "pip", Category: "python-package", Executables: []string{"pip", "pip3"}, Args: []string{"--version"}},
	{Name: "pipx", Category: "python-package", Executables: []string{"pipx"}, Args: []string{"--version"}},
	{Name: "uv", Category: "python-package", Executables: []string{"uv"}, Args: []string{"--version"}},
	{Name: "poetry", Category: "python-package", Executables: []string{"poetry"}, Args: []string{"--version"}},
	{Name: "pipenv", Category: "python-package", Executables: []string{"pipenv"}, Args: []string{"--version"}},
	{Name: "pdm", Category: "python-package", Executables: []string{"pdm"}, Args: []string{"--version"}},
	{Name: "hatch", Category: "python-package", Executables: []string{"hatch"}, Args: []string{"--version"}},
	{Name: "conda", Category: "python-package", Executables: []string{"conda"}, Args: []string{"--version"}},
	{Name: "mamba", Category: "python-package", Executables: []string{"mamba"}, Args: []string{"--version"}},
	{Name: "micromamba", Category: "python-package", Executables: []string{"micromamba"}, Args: []string{"--version"}},
	{Name: "pixi", Category: "python-package", Executables: []string{"pixi"}, Args: []string{"--version"}},
	{Name: "rye", Category: "python-package", Executables: []string{"rye"}, Args: []string{"--version"}},

	{Name: "node", Category: "javascript", Executables: []string{"node"}, Args: []string{"--version"}},
	{Name: "npm", Category: "javascript-package", Executables: []string{"npm"}, Args: []string{"--version"}},
	{Name: "yarn", Category: "javascript-package", Executables: []string{"yarn"}, Args: []string{"--version"}},
	{Name: "pnpm", Category: "javascript-package", Executables: []string{"pnpm"}, Args: []string{"--version"}},
	{Name: "corepack", Category: "javascript-package", Executables: []string{"corepack"}, Args: []string{"--version"}},
	{Name: "bun", Category: "javascript", Executables: []string{"bun"}, Args: []string{"--version"}},
	{Name: "deno", Category: "javascript", Executables: []string{"deno"}, Args: []string{"--version"}},

	{Name: "java", Category: "java", Executables: []string{"java"}, Args: []string{"-version"}},
	{Name: "maven", Category: "java-package", Executables: []string{"mvn"}, Args: []string{"--version"}},
	{Name: "gradle", Category: "java-package", Executables: []string{"gradle"}, Args: []string{"--version"}},
	{Name: "ant", Category: "java-package", Executables: []string{"ant"}, Args: []string{"-version"}},

	{Name: "dotnet", Category: "dotnet", Executables: []string{"dotnet"}, Args: []string{"--version"}},
	{Name: "nuget", Category: "dotnet-package", Executables: []string{"nuget"}, Args: []string{"help"}},
	{Name: "go", Category: "go", Executables: []string{"go"}, Args: []string{"version"}},
	{Name: "rustc", Category: "rust", Executables: []string{"rustc"}, Args: []string{"--version"}},
	{Name: "cargo", Category: "rust-package", Executables: []string{"cargo"}, Args: []string{"--version"}},
	{Name: "rustup", Category: "rust-package", Executables: []string{"rustup"}, Args: []string{"--version"}},

	{Name: "php", Category: "php", Executables: []string{"php"}, Args: []string{"--version"}},
	{Name: "composer", Category: "php-package", Executables: []string{"composer"}, Args: []string{"--version"}},
	{Name: "ruby", Category: "ruby", Executables: []string{"ruby"}, Args: []string{"--version"}},
	{Name: "rubygems", Category: "ruby-package", Executables: []string{"gem"}, Args: []string{"--version"}},
	{Name: "bundler", Category: "ruby-package", Executables: []string{"bundle"}, Args: []string{"--version"}},
	{Name: "perl", Category: "perl", Executables: []string{"perl"}, Args: []string{"-v"}},
	{Name: "cpan", Category: "perl-package", Executables: []string{"cpan"}, Args: []string{"-v"}},
	{Name: "cpanm", Category: "perl-package", Executables: []string{"cpanm"}, Args: []string{"--version"}},
	{Name: "erlang", Category: "erlang", Executables: []string{"erl"}, Args: []string{"-version"}},
	{Name: "elixir", Category: "elixir", Executables: []string{"elixir"}, Args: []string{"--version"}},
	{Name: "mix", Category: "elixir-package", Executables: []string{"mix"}, Args: []string{"--version"}},
	{Name: "rebar3", Category: "erlang-package", Executables: []string{"rebar3"}, Args: []string{"version"}},

	{Name: "curl", Category: "tls-client", Executables: []string{"curl"}, Args: []string{"--version"}},
	{Name: "wget", Category: "tls-client", Executables: []string{"wget"}, Args: []string{"--version"}},
	{Name: "git", Category: "developer-tool", Executables: []string{"git"}, Args: []string{"--version"}},
	{Name: "openssl", Category: "tls-client", Executables: []string{"openssl"}, Args: []string{"version"}},
	{Name: "mise", Category: "version-manager", Executables: []string{"mise"}, Args: []string{"--version"}},
	{Name: "asdf", Category: "version-manager", Executables: []string{"asdf"}, Args: []string{"--version"}},
	{Name: "claude", Category: "developer-tool", Executables: []string{"claude"}, Args: []string{"--version"}},
}

func collectInventory(ctx context.Context, timeout time.Duration) []Component {
	var components []Component
	seen := map[string]bool{}
	for _, spec := range toolSpecs {
		for _, executable := range spec.Executables {
			for _, path := range findExecutablePaths(executable) {
				key := strings.ToLower(spec.Name + "\x00" + cleanPath(path))
				if seen[key] {
					continue
				}
				seen[key] = true
				output, err := runCommand(ctx, timeout, path, spec.Args...)
				component := Component{
					Name:     spec.Name,
					Category: spec.Category,
					Path:     cleanPath(path),
					Source:   "PATH",
					Status:   "inventory-only",
					Evidence: oneLine(output, 400),
				}
				component.Version = extractVersion(spec.Name, output)
				if err != nil {
					component.Error = err.Error()
					if component.Version == "" {
						component.Status = "unknown"
					}
				}
				components = append(components, component)
			}
		}
	}
	components = append(components, inspectPythonInstallations(ctx, timeout)...)
	return deduplicateComponents(components)
}

func runCommand(parent context.Context, timeout time.Duration, path string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, path, args...)
	cmd.Env = os.Environ()
	output, err := cmd.CombinedOutput()
	text := strings.TrimSpace(string(output))
	if ctx.Err() == context.DeadlineExceeded {
		return text, errors.New("command timed out")
	}
	return text, err
}

func findExecutablePaths(name string) []string {
	var candidates []string
	extensions := []string{""}
	if runtime.GOOS == "windows" && filepath.Ext(name) == "" {
		extensions = []string{".exe", ".com", ".bat", ".cmd", ""}
	}
	for _, dir := range filepath.SplitList(os.Getenv("PATH")) {
		dir = strings.Trim(strings.TrimSpace(dir), `"`)
		if dir == "" {
			continue
		}
		for _, extension := range extensions {
			path := filepath.Join(dir, name+extension)
			info, err := os.Stat(path)
			if err == nil && !info.IsDir() {
				candidates = append(candidates, cleanPath(path))
			}
		}
	}
	if path, err := exec.LookPath(name); err == nil {
		candidates = append(candidates, cleanPath(path))
	}
	seen := map[string]bool{}
	var result []string
	for _, path := range candidates {
		key := path
		if runtime.GOOS == "windows" {
			key = strings.ToLower(path)
		}
		if !seen[key] {
			seen[key] = true
			result = append(result, path)
		}
	}
	sort.Strings(result)
	return result
}

func cleanPath(path string) string {
	abs, err := filepath.Abs(path)
	if err == nil {
		path = abs
	}
	if evaluated, err := filepath.EvalSymlinks(path); err == nil {
		path = evaluated
	}
	return filepath.Clean(path)
}

func pathIdentity(path string) string {
	if runtime.GOOS == "windows" {
		return strings.ToLower(path)
	}
	return path
}

func deduplicateComponents(components []Component) []Component {
	seen := map[string]bool{}
	result := make([]Component, 0, len(components))
	for _, component := range components {
		key := strings.ToLower(component.Name + "\x00" + component.Path + "\x00" + component.Version)
		if seen[key] {
			continue
		}
		seen[key] = true
		result = append(result, component)
	}
	return result
}

func redactHome(path string) string {
	if path == "" {
		return ""
	}
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return path
	}
	if runtime.GOOS == "windows" {
		if strings.HasPrefix(strings.ToLower(path), strings.ToLower(home)) {
			return "~" + path[len(home):]
		}
		return path
	}
	if strings.HasPrefix(path, home) {
		return "~" + path[len(home):]
	}
	return path
}

func redactText(value string) string {
	if value == "" {
		return ""
	}
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return value
	}
	value = replaceFold(value, home, "~")
	if runtime.GOOS == "windows" {
		value = replaceFold(value, strings.ReplaceAll(home, `\\`, "/"), "~")
	}
	return value
}

func replaceFold(value, old, replacement string) string {
	if old == "" {
		return value
	}
	for {
		index := strings.Index(strings.ToLower(value), strings.ToLower(old))
		if index < 0 {
			return value
		}
		value = value[:index] + replacement + value[index+len(old):]
	}
}

func oneLine(value string, max int) string {
	value = strings.Join(strings.Fields(value), " ")
	if len(value) > max {
		return value[:max-1] + "…"
	}
	return value
}
