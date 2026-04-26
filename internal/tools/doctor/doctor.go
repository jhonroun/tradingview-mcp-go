// Package doctor provides actionable system diagnostics for TradingView CDP setup.
package doctor

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// Report is the full diagnostic output returned by Run.
type Report struct {
	Port      PortCheck    `json:"port"`
	Process   ProcessCheck `json:"process"`
	Install   InstallCheck `json:"install"`
	LaunchCmd string       `json:"launch_cmd,omitempty"`
	Hints     []string     `json:"hints"`
}

// PortCheck describes the state of localhost:9222.
type PortCheck struct {
	Reachable bool   `json:"reachable"`
	CDP       bool   `json:"cdp"`
	Owner     string `json:"owner,omitempty"` // process name if not TradingView
	Error     string `json:"error,omitempty"`
}

// ProcessCheck describes the running TradingView process (Windows-only; empty on other OSes).
type ProcessCheck struct {
	Running    bool   `json:"running"`
	PID        int    `json:"pid,omitempty"`
	HasCDPFlag bool   `json:"has_cdp_flag"`
	CmdLine    string `json:"cmdline,omitempty"`
}

// InstallCheck describes what was found on disk.
type InstallCheck struct {
	Found        bool   `json:"found"`
	Path         string `json:"path,omitempty"`
	Source       string `json:"source,omitempty"`
	IsMSIX       bool   `json:"is_msix,omitempty"`
	LocalAppData string `json:"local_appdata_dir,omitempty"` // %LOCALAPPDATA%\TradingView if exists
	AppDataDir   string `json:"appdata_dir,omitempty"`       // %APPDATA%\TradingView if exists
}

// Run performs all diagnostics and returns a Report with actionable hints.
func Run() *Report {
	r := &Report{}
	r.Port = checkPort(9222)
	r.Process = checkProcess()
	r.Install = checkInstall()
	r.LaunchCmd = buildLaunchCmd(r)
	r.Hints = buildHints(r)
	return r
}

// checkPort probes localhost:<port>/json/list and identifies the port owner.
func checkPort(port int) PortCheck {
	addr := fmt.Sprintf("http://localhost:%d/json/list", port)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, "GET", addr, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		ch := PortCheck{Reachable: false}
		owner := portOwner(port)
		if owner != "" {
			ch.Owner = owner
			ch.Error = fmt.Sprintf("port in use by %q but not CDP", owner)
		} else {
			ch.Error = "connection refused — port not listening"
		}
		return ch
	}
	defer resp.Body.Close()
	return PortCheck{
		Reachable: true,
		CDP:       resp.StatusCode == 200,
		Owner:     portOwner(port),
	}
}

// portOwner returns the name of the process listening on the given TCP port.
// Windows-only; returns "" on other platforms or on error.
func portOwner(port int) string {
	if runtime.GOOS != "windows" {
		return ""
	}
	out, err := exec.Command("netstat", "-ano").Output()
	if err != nil {
		return ""
	}
	portSuffix := fmt.Sprintf(":%d", port)
	var pid string
	for _, line := range strings.Split(string(out), "\n") {
		fields := strings.Fields(line)
		// TCP  local_addr  remote_addr  state  pid
		if len(fields) < 5 {
			continue
		}
		if fields[0] != "TCP" {
			continue
		}
		if !strings.HasSuffix(fields[1], portSuffix) {
			continue
		}
		if fields[3] == "LISTENING" {
			pid = fields[4]
			break
		}
	}
	if pid == "" || pid == "0" {
		return ""
	}
	nameOut, err := exec.Command(
		"tasklist", "/FI", "PID eq "+pid, "/FO", "CSV", "/NH",
	).Output()
	if err != nil {
		return ""
	}
	line := strings.TrimSpace(string(nameOut))
	if line == "" || strings.HasPrefix(strings.ToUpper(line), "INFO:") {
		return ""
	}
	// CSV: "TradingView.exe","1234","Console","1","42,736 K"
	parts := strings.SplitN(line, ",", 2)
	return strings.Trim(parts[0], `"`)
}

// checkProcess finds TradingView.exe and checks its command-line for the CDP flag.
// Windows-only; returns empty ProcessCheck on other platforms.
func checkProcess() ProcessCheck {
	if runtime.GOOS != "windows" {
		return ProcessCheck{}
	}
	out, err := exec.Command(
		"tasklist", "/FI", "IMAGENAME eq TradingView.exe", "/FO", "CSV", "/NH",
	).Output()
	if err != nil {
		return ProcessCheck{}
	}
	line := strings.TrimSpace(string(out))
	if line == "" || strings.HasPrefix(strings.ToUpper(line), "INFO:") {
		return ProcessCheck{Running: false}
	}
	// "TradingView.exe","12345","Console","1","..."
	parts := strings.Split(line, ",")
	pc := ProcessCheck{Running: true}
	if len(parts) >= 2 {
		pc.PID, _ = strconv.Atoi(strings.Trim(parts[1], `"`))
	}
	if pc.PID > 0 {
		pc.CmdLine = getCommandLine(pc.PID)
		pc.HasCDPFlag = strings.Contains(pc.CmdLine, "--remote-debugging-port")
	}
	return pc
}

// getCommandLine fetches the full command line of a process via PowerShell.
func getCommandLine(pid int) string {
	out, err := exec.Command(
		"powershell", "-NoProfile", "-NonInteractive", "-Command",
		fmt.Sprintf(
			`(Get-CimInstance Win32_Process -Filter "ProcessId=%d" -ErrorAction SilentlyContinue).CommandLine`,
			pid,
		),
	).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// checkInstall locates the TradingView executable and probes AppData directories.
func checkInstall() InstallCheck {
	ic := InstallCheck{}

	// Probe data directories regardless of whether the exe is found.
	if localAppData := os.Getenv("LOCALAPPDATA"); localAppData != "" {
		dir := filepath.Join(localAppData, "TradingView")
		if _, err := os.Stat(dir); err == nil {
			ic.LocalAppData = dir
		}
	}
	if appData := os.Getenv("APPDATA"); appData != "" {
		dir := filepath.Join(appData, "TradingView")
		if _, err := os.Stat(dir); err == nil {
			ic.AppDataDir = dir
		}
	}

	path, source, isMSIX := findExecutable()
	if path != "" {
		ic.Found = true
		ic.Path = path
		ic.Source = source
		ic.IsMSIX = isMSIX
	}
	return ic
}

// findExecutable searches for TradingView.exe using the same priority as the
// discovery package but inlined so doctor has no import cycle risk.
func findExecutable() (path, source string, isMSIX bool) {
	if p := os.Getenv("TRADINGVIEW_PATH"); p != "" {
		if _, err := os.Stat(p); err == nil {
			return p, "TRADINGVIEW_PATH env", false
		}
	}
	if runtime.GOOS != "windows" {
		return "", "", false
	}

	localAppData := os.Getenv("LOCALAPPDATA")
	programFiles := os.Getenv("PROGRAMFILES")
	programFiles86 := os.Getenv("PROGRAMFILES(X86)")
	candidates := []struct{ p, s string }{
		{filepath.Join(localAppData, "TradingView", "TradingView.exe"), "LOCALAPPDATA"},
		{filepath.Join(programFiles, "TradingView", "TradingView.exe"), "PROGRAMFILES"},
		{filepath.Join(programFiles86, "TradingView", "TradingView.exe"), "PROGRAMFILES(X86)"},
	}
	for _, c := range candidates {
		if _, err := os.Stat(c.p); err == nil {
			return c.p, c.s, false
		}
	}

	// Probe MSIX via Get-AppxPackage *TradingView* (broad wildcard per Phase 3 spec).
	out, err := exec.Command(
		"powershell", "-NoProfile", "-NonInteractive", "-Command",
		`$pkg = Get-AppxPackage -Name '*TradingView*' -ErrorAction SilentlyContinue | Select-Object -First 1; `+
			`if ($pkg) { $pkg.InstallLocation + '|' + $pkg.PackageFamilyName }`,
	).Output()
	if err == nil {
		line := strings.TrimSpace(string(out))
		if line != "" {
			parts := strings.SplitN(line, "|", 2)
			loc := parts[0]
			for _, rel := range []string{"TradingView.exe", filepath.Join("app", "TradingView.exe")} {
				p := filepath.Join(loc, rel)
				if _, err := os.Stat(p); err == nil {
					return p, "Microsoft Store (WindowsApps)", true
				}
			}
		}
	}
	return "", "", false
}

// buildLaunchCmd returns the exact shell command the user should run.
func buildLaunchCmd(r *Report) string {
	if r.Install.Found {
		dir := filepath.Dir(r.Install.Path)
		return fmt.Sprintf(`cd "%s" && "%s" --remote-debugging-port=9222`, dir, r.Install.Path)
	}
	return "tv launch"
}

// buildHints generates ordered, actionable hint strings from the report state.
func buildHints(r *Report) []string {
	var hints []string

	// Port conflict with a non-TradingView process.
	owner := r.Port.Owner
	if !r.Port.Reachable && owner != "" && !isTradingViewProc(owner) {
		hints = append(hints,
			fmt.Sprintf("Port 9222 is in use by %q. Close it or choose a different port.", owner),
		)
	}

	// TradingView running but without CDP flag.
	if r.Process.Running && !r.Process.HasCDPFlag {
		hints = append(hints,
			"TradingView.exe is running but --remote-debugging-port is not set. Restart it: tv launch --kill",
		)
		if r.LaunchCmd != "" && r.LaunchCmd != "tv launch" {
			hints = append(hints,
				"Or restart manually: "+r.LaunchCmd,
			)
		}
	}

	// Not running at all.
	if !r.Process.Running && !r.Port.Reachable {
		if r.Install.Found {
			hints = append(hints, "TradingView is not running. Start it: tv launch")
		} else {
			hints = append(hints,
				"TradingView Desktop not found. Install from tradingview.com or set TRADINGVIEW_PATH.",
			)
		}
	}

	// CDP reachable but the /json/list endpoint returned non-200.
	if r.Port.Reachable && !r.Port.CDP {
		hints = append(hints,
			"Port 9222 is open but did not return a CDP target list. Another service may be using it.",
		)
	}

	// Installation missing entirely.
	if !r.Install.Found {
		hints = append(hints,
			"TradingView Desktop executable not found. Install from tradingview.com or set TRADINGVIEW_PATH.",
		)
	}

	// All green.
	if r.Port.Reachable && r.Port.CDP && (r.Process.HasCDPFlag || !r.Process.Running) && r.Install.Found && len(hints) == 0 {
		hints = append(hints, "CDP is available. Run: tv status")
	}

	return hints
}

func isTradingViewProc(name string) bool {
	return strings.Contains(strings.ToLower(name), "tradingview")
}
