package config

import (
	"flag"
	"slices"
	"strings"

	"github.com/strowk/mcp-k8s-go/internal/oidc"
)

// Options represents the global configuration options
type Options struct {
	// AllowedContexts is a list of k8s contexts that users are allowed to access
	// If empty, all contexts are allowed
	AllowedContexts []string

	// Readonly determine if tools that 'write' to the cluster are
	// registered and advertised to clients.
	Readonly bool

	OidcConfig oidc.OIDCConfig

	RemoteHostname        string // Hostname of the remote server, if applicable
	RemotePort            int    // Port of the remote server, if applicable
	RemotePath            string // Path of the remote server, if applicable
	RemoteCallbackPath    string // Path for the remote server to call back to after OIDC authentication
	TLSCertificateKeyPath string // Path to the TLS certificate key file, if applicable
	TLSCertificatePath    string // Path to the TLS certificate file, if applicable
}

// GlobalOptions contains the parsed command line options
var GlobalOptions = &Options{
	OidcConfig: oidc.OIDCConfig{},
}

// ParseFlags parses the command line flags
func ParseFlags() bool {
	var allowedContextsStr string
	flag.StringVar(
		&allowedContextsStr,
		"allowed-contexts",
		"",
		"Comma-separated list of allowed k8s contexts. If empty, all contexts are allowed",
	)
	flag.BoolVar(
		&GlobalOptions.Readonly,
		"readonly",
		false,
		"Disables any tool which can write changes to the cluster. If not specified, all tools are allowed",
	)
	flag.StringVar(
		&GlobalOptions.OidcConfig.AuthStyle,
		"oidc-auth-style",
		"",
		"OIDC authentication style. Can be 'in_params', 'in_header', or 'auto_detect'. Defaults to 'in_header'. OIDC configuration is only used for remote deployments.",
	)
	flag.StringVar(
		&GlobalOptions.OidcConfig.AuthURL,
		"oidc-auth-url",
		"",
		"OIDC authentication URL. OIDC configuration is only used for remote deployments.",
	)
	flag.StringVar(
		&GlobalOptions.OidcConfig.TokenURL,
		"oidc-token-url",
		"",
		"OIDC token URL. OIDC configuration is only used for remote deployments.",
	)
	flag.StringVar(
		&GlobalOptions.OidcConfig.ClientID,
		"oidc-client-id",
		"",
		"OIDC client ID. OIDC configuration is only used for remote deployments.",
	)
	flag.StringVar(
		&GlobalOptions.OidcConfig.ClientSecret,
		"oidc-client-secret",
		"",
		"OIDC client secret. OIDC configuration is only used for remote deployments.",
	)
	flag.StringVar(
		&GlobalOptions.OidcConfig.RedirectURL,
		"oidc-redirect-url",
		"",
		"OIDC redirect URL. OIDC configuration is only used for remote deployments.",
	)
	flag.StringVar(
		&GlobalOptions.RemoteHostname,
		"remote-hostname",
		"",
		"Hostname of the remote server. If not specified, the server is assumed to be running locally.",
	)
	flag.IntVar(
		&GlobalOptions.RemotePort,
		"remote-port",
		80,
		"Port of the remote server.",
	)
	flag.StringVar(
		&GlobalOptions.RemotePath,
		"remote-path",
		"/mcp",
		"Path of the remote server.",
	)
	flag.StringVar(
		&GlobalOptions.RemoteCallbackPath,
		"remote-callback-path",
		"/callback",
		"Path for the remote server to call back to after OIDC authentication. This is used to handle the OIDC authentication flow.",
	)

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
