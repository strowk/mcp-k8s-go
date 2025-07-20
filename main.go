package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"os"

	"github.com/strowk/mcp-k8s-go/internal/config"
	"github.com/strowk/mcp-k8s-go/internal/k8s"
	"github.com/strowk/mcp-k8s-go/internal/k8s/apps/v1/deployment"
	"github.com/strowk/mcp-k8s-go/internal/k8s/core/v1/service"
	"github.com/strowk/mcp-k8s-go/internal/k8s/list_mapping"
	"github.com/strowk/mcp-k8s-go/internal/oidc"
	"github.com/strowk/mcp-k8s-go/internal/prompts"
	"github.com/strowk/mcp-k8s-go/internal/resources"
	"github.com/strowk/mcp-k8s-go/internal/tools"
	"github.com/strowk/mcp-k8s-go/internal/utils"

	"github.com/strowk/foxy-contexts/pkg/app"
	"github.com/strowk/foxy-contexts/pkg/auth"
	"github.com/strowk/foxy-contexts/pkg/mcp"
	"github.com/strowk/foxy-contexts/pkg/stdio"
	"github.com/strowk/foxy-contexts/pkg/streamable_http"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/exec"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	"k8s.io/client-go/tools/clientcmd"
)

func getCapabilities() *mcp.ServerCapabilities {
	return &mcp.ServerCapabilities{
		Resources: &mcp.ServerCapabilitiesResources{
			ListChanged: utils.Ptr(false),
			Subscribe:   utils.Ptr(false),
		},
		Prompts: &mcp.ServerCapabilitiesPrompts{
			ListChanged: utils.Ptr(false),
		},
		Tools: &mcp.ServerCapabilitiesTools{
			ListChanged: utils.Ptr(false),
		},
	}
}

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	if len(os.Args) > 1 {
		arg := os.Args[1]
		if arg == "--version" {
			println(version)
			return
		}
		if arg == "version" {
			println("Version: ", version)
			println("Commit: ", commit)
			println("Date: ", date)
			return
		}
		if arg == "help" || arg == "--help" {
			printHelp()
			return
		}
	}

	// Parse configuration flags
	shouldContinue := config.ParseFlags()
	if !shouldContinue {
		return
	}

	foxyApp, err := getApp()
	if err != nil {
		log.Fatalf("Error creating app: %v", err)
	}
	err = foxyApp.Run()
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func printHelp() {
	println("mcp-k8s is an MCP server for Kubernetes")
	println("Read more about it in: https://github.com/strowk/mcp-k8s-go\n")
	println("Usage: <bin> [flags]")
	println("  Run with no flags to start the server\n")
	println("Flags:")
	println("  help, --help: Print this help message")
	println("  --version: Print the version of the server")
	println("  version: Print the version, commit and date of the server")
	println("  --allowed-contexts=<ctx1,ctx2,...>: Comma-separated list of allowed k8s contexts")
	println("      If not specified, all contexts are allowed")
	println("  --readonly: Disables any tool which can write changes to the cluster")
	println("      If not specified, all tools are available")
}

func getApp() (*app.Builder, error) {
	app := app.
		NewBuilder().
		WithFxOptions(
			fx.Provide(func() clientcmd.ClientConfig {
				return k8s.GetKubeConfig()
			}),
			fx.Provide(func() (*kubernetes.Clientset, error) {
				return k8s.GetKubeClientset()
			}),
			fx.Provide(fx.Annotate(
				k8s.NewClientPool,
				fx.ParamTags(list_mapping.MappingResolversTag),
			)),
			fx.Provide(
				list_mapping.AsMappingResolver(func() list_mapping.ListMappingResolver {
					return deployment.NewListMappingResolver()
				}),
			),
			fx.Provide(
				list_mapping.AsMappingResolver(func() list_mapping.ListMappingResolver {
					return service.NewListMappingResolver()
				}),
			),
		).
		WithTool(tools.NewPodLogsTool).
		WithTool(tools.NewListContextsTool).
		WithTool(tools.NewListNamespacesTool).
		WithTool(tools.NewListResourcesTool).
		WithTool(tools.NewGetResourceTool).
		WithTool(tools.NewListNodesTool).
		WithTool(tools.NewListEventsTool).
		WithPrompt(prompts.NewListPodsPrompt).
		WithPrompt(prompts.NewListNamespacesPrompt).
		WithResourceProvider(resources.NewContextsResourceProvider).
		WithServerCapabilities(getCapabilities()).
		// setting up server
		WithName("mcp-k8s-go").
		WithVersion(version).
		// Configuring fx logging to only show errors
		WithLogger(func() *zap.Logger {
			cfg := zap.NewDevelopmentConfig()
			cfg.Level.SetLevel(zap.ErrorLevel)
			logger, _ := cfg.Build()
			return logger
		}())

	if config.GlobalOptions.RemoteHostname == "" {
		app = app.WithTransport(stdio.NewTransport())
	} else {
		remoteHostname := config.GlobalOptions.RemoteHostname
		remotePort := config.GlobalOptions.RemotePort
		remotePath := config.GlobalOptions.RemotePath
		remoteCallbackPath := config.GlobalOptions.RemoteCallbackPath
		oidcCfg := &config.GlobalOptions.OidcConfig
		endpoint := streamable_http.Endpoint{
			Hostname:   remoteHostname,
			Port:       remotePort,
			Path:       remotePath,
			AuthScheme: "http",
			AuthHost: fmt.Sprintf(
				"%s:%d",
				remoteHostname,
				remotePort,
			),
		}
		transportOptions := []streamable_http.TransportOption{
			endpoint,
			oidc.OidcEchoConfigurer(oidcCfg, remoteCallbackPath),
		}
		if config.GlobalOptions.TLSCertificatePath != "" ||
			config.GlobalOptions.TLSCertificateKeyPath != "" {
			cert, err := tls.LoadX509KeyPair(
				config.GlobalOptions.TLSCertificatePath,
				config.GlobalOptions.TLSCertificateKeyPath,
			)
			if err != nil {
				return nil, err
			}
			transportOptions = append(transportOptions,
				&streamable_http.TlsConfig{
					TLS: &tls.Config{
						Certificates: []tls.Certificate{cert},
					},
				},
			)
			endpoint.AuthScheme = "https"
		}
		app = app.
			WithTransport(streamable_http.NewTransport(transportOptions...)).
			WithAuthorization(auth.Must(auth.NewOauth2Authorization(
				auth.WithExcludedPaths(remoteCallbackPath),
				auth.WithAuthorizationHandler(oidc.AuthorizationHandler(oidcCfg)),
			)))
	}

	// Skip registering remaining tools if --readonly detected
	if config.GlobalOptions.Readonly {
		println("Mode=Readonly, Skipping remaining tools")
		return app, nil
	}

	app = app.WithTool(tools.NewApplyK8sResourceTool).WithTool(tools.NewPodExecCommandTool)
	return app, nil
}
