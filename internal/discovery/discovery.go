// Package discovery finds the TradingView Desktop executable on the local machine.
package discovery

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// Result holds the discovered TradingView executable path.
type Result struct {
	Path     string
	Source   string
	Platform string
}

// Find searches for TradingView Desktop. Checks TRADINGVIEW_PATH env first,
// then platform-specific standard locations.
func Find() (*Result, error) {
	if p := os.Getenv("TRADINGVIEW_PATH"); p != "" {
		if _, err := os.Stat(p); err == nil {
			return &Result{Path: p, Source: "TRADINGVIEW_PATH env", Platform: runtime.GOOS}, nil
		}
	}
	switch runtime.GOOS {
	case "windows":
		return findWindows()
	case "darwin":
		return findMacOS()
	default:
		return findLinux()
	}
}

func findWindows() (*Result, error) {
	localAppData := os.Getenv("LOCALAPPDATA")
	programFiles := os.Getenv("PROGRAMFILES")
	programFiles86 := os.Getenv("PROGRAMFILES(X86)")

	candidates := []struct{ path, source string }{
		{filepath.Join(localAppData, "TradingView", "TradingView.exe"), "LOCALAPPDATA"},
		{filepath.Join(programFiles, "TradingView", "TradingView.exe"), "PROGRAMFILES"},
		{filepath.Join(programFiles86, "TradingView", "TradingView.exe"), "PROGRAMFILES(X86)"},
	}
	for _, c := range candidates {
		if _, err := os.Stat(c.path); err == nil {
			return &Result{Path: c.path, Source: c.source, Platform: "windows"}, nil
		}
	}
	if p := findWindowsStore(); p != "" {
		return &Result{Path: p, Source: "Microsoft Store (WindowsApps)", Platform: "windows"}, nil
	}
	return nil, fmt.Errorf("TradingView Desktop not found on Windows; set TRADINGVIEW_PATH or install from tradingview.com")
}

// findWindowsStore queries PowerShell for the Microsoft Store package location.
// Returns empty string on any error (Access Denied, not installed, etc.).
func findWindowsStore() string {
	out, err := exec.Command(
		"powershell", "-NoProfile", "-NonInteractive",
		"-Command", `(Get-AppxPackage -Name 'TradingView.Desktop' -ErrorAction SilentlyContinue).InstallLocation`,
	).Output()
	if err != nil {
		return ""
	}
	loc := strings.TrimSpace(string(out))
	if loc == "" {
		return ""
	}
	for _, rel := range []string{"TradingView.exe", filepath.Join("app", "TradingView.exe")} {
		if p := filepath.Join(loc, rel); fileExists(p) {
			return p
		}
	}
	return ""
}

func findMacOS() (*Result, error) {
	home := os.Getenv("HOME")
	candidates := []string{
		"/Applications/TradingView.app/Contents/MacOS/TradingView",
		filepath.Join(home, "Applications", "TradingView.app", "Contents", "MacOS", "TradingView"),
	}
	for _, p := range candidates {
		if fileExists(p) {
			return &Result{Path: p, Source: "Applications", Platform: "darwin"}, nil
		}
	}
	return nil, fmt.Errorf("TradingView Desktop not found on macOS; set TRADINGVIEW_PATH or install from tradingview.com")
}

func findLinux() (*Result, error) {
	for _, name := range []string{"tradingview", "TradingView"} {
		if p, err := exec.LookPath(name); err == nil {
			return &Result{Path: p, Source: "PATH", Platform: "linux"}, nil
		}
	}
	home := os.Getenv("HOME")
	candidates := []string{
		"/opt/TradingView/tradingview",
		"/usr/local/bin/tradingview",
		filepath.Join(home, ".local", "share", "TradingView", "tradingview"),
	}
	for _, p := range candidates {
		if fileExists(p) {
			return &Result{Path: p, Source: "standard path", Platform: "linux"}, nil
		}
	}
	return nil, fmt.Errorf("TradingView Desktop not found on Linux; set TRADINGVIEW_PATH or install from tradingview.com")
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
