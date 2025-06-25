package config

import (
	"flag"
	"slices"
	"strings"
)

// Options represents the global configuration options
type Options struct {
	// AllowedContexts is a list of k8s contexts that users are allowed to access
	// If empty, all contexts are allowed
	AllowedContexts []string

	// Readonly determine if tools that 'write' to the cluster are
	// registered and advertised to clients.
	Readonly bool
}

// GlobalOptions contains the parsed command line options
var GlobalOptions = &Options{}

// ParseFlags parses the command line flags
func ParseFlags() bool {
	var allowedContextsStr string
	flag.StringVar(&allowedContextsStr, "allowed-contexts", "", "Comma-separated list of allowed k8s contexts. If empty, all contexts are allowed")
	flag.BoolVar(&GlobalOptions.Readonly, "readonly", false, "Disables any tool which can write changes to the cluster. If not specified, all tools are allowed")

	// Add other flags here

	// Parse the flags
	flag.Parse()

	// Check if the flag is --version, version, help or --help
	// If so, we don't need to continue processing
	if len(flag.Args()) > 0 {
		arg := flag.Args()[0]
		if arg == "--version" || arg == "version" || arg == "help" || arg == "--help" {
			return false
		}
	}

	// Process allowed contexts
	if allowedContextsStr != "" {
		GlobalOptions.AllowedContexts = strings.Split(allowedContextsStr, ",")
		for i, ctx := range GlobalOptions.AllowedContexts {
			GlobalOptions.AllowedContexts[i] = strings.TrimSpace(ctx)
		}
	}

	return true
}

// IsContextAllowed checks if a context is allowed based on the configuration
func IsContextAllowed(contextName string) bool {
	// If the allowed contexts list is empty, all contexts are allowed
	if len(GlobalOptions.AllowedContexts) == 0 {
		return true
	}

	return slices.Contains(GlobalOptions.AllowedContexts, contextName)
}
