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
	Path            string
	Source          string
	Platform        string
	IsMSIX          bool   // true when installed via Microsoft Store (WindowsApps)
	MSIXFamilyName  string // e.g. "TradingView.Desktop_n534cwy3pjxzj", set when IsMSIX=true
	MSIXAppID       string // e.g. "TradingView.Desktop", the Application Id from AppxManifest
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
	if r := findWindowsStore(); r != nil {
		return r, nil
	}
	return nil, fmt.Errorf("TradingView Desktop not found on Windows; set TRADINGVIEW_PATH or install from tradingview.com")
}

// findWindowsStore queries PowerShell for the Microsoft Store package location.
// Returns nil on any error (Access Denied, not installed, etc.).
func findWindowsStore() *Result {
	out, err := exec.Command(
		"powershell", "-NoProfile", "-NonInteractive",
		"-Command",
		`$pkg = Get-AppxPackage -Name 'TradingView.Desktop' -ErrorAction SilentlyContinue; if ($pkg) { $pkg.InstallLocation + '|' + $pkg.PackageFamilyName }`,
	).Output()
	if err != nil {
		return nil
	}
	line := strings.TrimSpace(string(out))
	if line == "" {
		return nil
	}
	parts := strings.SplitN(line, "|", 2)
	loc := parts[0]
	family := ""
	if len(parts) == 2 {
		family = parts[1]
	}
	for _, rel := range []string{"TradingView.exe", filepath.Join("app", "TradingView.exe")} {
		p := filepath.Join(loc, rel)
		if fileExists(p) {
			return &Result{
				Path:           p,
				Source:         "Microsoft Store (WindowsApps)",
				Platform:       "windows",
				IsMSIX:         true,
				MSIXFamilyName: family,
				MSIXAppID:      "TradingView.Desktop",
			}
		}
	}
	return nil
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
