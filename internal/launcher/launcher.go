// Package launcher starts TradingView Desktop with CDP remote debugging enabled.
package launcher

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/jhonroun/tradingview-mcp-go/internal/cdp"
	"github.com/jhonroun/tradingview-mcp-go/internal/discovery"
)

// Launch finds and starts TradingView with --remote-debugging-port=<port>.
// TradingView.exe must be run from its own installation directory; cmd.Dir is
// set to filepath.Dir(execPath) so the flag is accepted.
// If tvPath is non-empty it is used directly, skipping auto-discovery.
// If killExisting is true, any running TradingView processes are killed first.
func Launch(port int, killExisting bool, tvPath string) (map[string]interface{}, error) {
	if !killExisting {
		// Return immediately if TradingView is already running with CDP.
		ctx0, cancel0 := context.WithTimeout(context.Background(), 2*time.Second)
		targets, err := cdp.ListTargets(ctx0, "localhost", port)
		cancel0()
		if err == nil {
			if _, err := cdp.FindChartTarget(targets); err == nil {
				return map[string]interface{}{
					"success":         true,
					"connected":       true,
					"already_running": true,
					"port":            port,
				}, nil
			}
		}
	}

	if killExisting {
		killRunning()
		time.Sleep(2 * time.Second)
	}

	var found *discovery.Result
	if tvPath != "" {
		if _, err := os.Stat(tvPath); err != nil {
			return map[string]interface{}{
				"success": false,
				"error":   fmt.Sprintf("--tv-path not found: %v", err),
			}, nil
		}
		found = &discovery.Result{Path: tvPath, Source: "cli-flag", Platform: runtime.GOOS}
	} else {
		var err error
		found, err = discovery.Find()
		if err != nil {
			return map[string]interface{}{"success": false, "error": err.Error()}, nil
		}
	}

	pid, err := launchDirect(found.Path, port)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("failed to start TradingView: %v", err),
		}, nil
	}

	// Wait up to 30 s for CDP to become available.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	base := map[string]interface{}{
		"success":  true,
		"pid":      pid,
		"port":     port,
		"platform": runtime.GOOS,
		"path":     found.Path,
		"source":   found.Source,
	}
	if found.IsMSIX {
		base["msix"] = true
	}

	for {
		targets, err := cdp.ListTargets(ctx, "localhost", port)
		if err == nil {
			if _, err := cdp.FindChartTarget(targets); err == nil {
				return base, nil
			}
		}
		select {
		case <-ctx.Done():
			base["hint"] = "TradingView started but chart not yet detected; open a chart"
			return base, nil
		case <-time.After(500 * time.Millisecond):
		}
	}
}

// launchDirect starts TradingView.exe with --remote-debugging-port=<port>.
// cmd.Dir is set to the installation directory — required for the CDP flag
// to be accepted by TradingView. Run tv launch from an interactive terminal.
func launchDirect(execPath string, port int) (int, error) {
	cmd := exec.Command(execPath, fmt.Sprintf("--remote-debugging-port=%d", port))
	cmd.Dir = filepath.Dir(execPath)
	if err := cmd.Start(); err != nil {
		return 0, err
	}
	return cmd.Process.Pid, nil
}

func killRunning() {
	switch runtime.GOOS {
	case "windows":
		_ = exec.Command("taskkill", "/F", "/IM", "TradingView.exe").Run()
	case "darwin":
		_ = exec.Command("pkill", "-f", "TradingView").Run()
	default:
		_ = exec.Command("pkill", "-f", "tradingview").Run()
	}
}
