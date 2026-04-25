// Package cli provides the command registry and dispatcher for the tv CLI.
package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// Handler executes a CLI command and returns a JSON-serialisable result.
type Handler func(args []string, opts map[string]string) (interface{}, error)

// Command describes one CLI command.
type Command struct {
	Name        string
	Description string
	Handler     Handler
}

var (
	registry = map[string]*Command{}
	order    []string
)

// Register adds a command to the global CLI registry.
func Register(cmd Command) {
	registry[cmd.Name] = &cmd
	order = append(order, cmd.Name)
}

// Dispatch parses os.Args[1:] and runs the matching command.
// Output is JSON on stdout; errors are JSON on stderr with exit code 1.
func Dispatch(args []string) {
	if len(args) == 0 || args[0] == "--help" || args[0] == "-h" || args[0] == "help" {
		printHelp()
		return
	}
	name := args[0]
	cmd, ok := registry[name]
	if !ok {
		writeError(fmt.Errorf("unknown command: %s", name))
		os.Exit(1)
	}
	opts, rest := parseFlags(args[1:])
	result, err := cmd.Handler(rest, opts)
	if err != nil {
		writeError(err)
		os.Exit(1)
	}
	out, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		writeError(fmt.Errorf("marshal error: %v", err))
		os.Exit(1)
	}
	fmt.Println(string(out))
}

func writeError(err error) {
	out, _ := json.MarshalIndent(map[string]interface{}{"success": false, "error": err.Error()}, "", "  ")
	fmt.Fprintln(os.Stderr, string(out))
}

func printHelp() {
	fmt.Fprintln(os.Stdout, "tv — TradingView CLI")
	fmt.Fprintln(os.Stdout)
	fmt.Fprintln(os.Stdout, "Commands:")
	for _, name := range order {
		cmd := registry[name]
		fmt.Fprintf(os.Stdout, "  %-20s %s\n", name, cmd.Description)
	}
}

// parseFlags extracts --key=value and --key value pairs from args,
// returning the flag map and the remaining positional arguments.
func parseFlags(args []string) (map[string]string, []string) {
	opts := make(map[string]string)
	var rest []string
	i := 0
	for i < len(args) {
		a := args[i]
		if strings.HasPrefix(a, "--") {
			key := a[2:]
			if k, v, ok := strings.Cut(key, "="); ok {
				opts[k] = v
			} else if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				i++
				opts[key] = args[i]
			} else {
				opts[key] = "true"
			}
		} else {
			rest = append(rest, a)
		}
		i++
	}
	return opts, rest
}
