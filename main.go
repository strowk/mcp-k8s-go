package main

import (
	"log"
	"os"

	"github.com/strowk/mcp-k8s-go/internal/config"
	"github.com/strowk/mcp-k8s-go/internal/k8s"
	"github.com/strowk/mcp-k8s-go/internal/prompts"
	"github.com/strowk/mcp-k8s-go/internal/resources"
	"github.com/strowk/mcp-k8s-go/internal/tools"
	"github.com/strowk/mcp-k8s-go/internal/utils"

	"github.com/strowk/foxy-contexts/pkg/app"
	"github.com/strowk/foxy-contexts/pkg/mcp"
	"github.com/strowk/foxy-contexts/pkg/stdio"

	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
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

	foxyApp := getApp()
	err := foxyApp.Run()
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
}

func getApp() *app.Builder {
	return app.
		NewBuilder().
		WithFxOptions(
			fx.Provide(func() clientcmd.ClientConfig {
				return k8s.GetKubeConfig()
			}),
			fx.Provide(func() (*kubernetes.Clientset, error) {
				return k8s.GetKubeClientset()
			}),
			fx.Provide(func() k8s.ClientPool {
				return k8s.NewClientPool()
			}),
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
		WithVersion("0.0.1").
		WithTransport(stdio.NewTransport()).
		// Configuring fx logging to only show errors
		WithFxOptions(
			fx.Provide(func() *zap.Logger {
				cfg := zap.NewDevelopmentConfig()
				cfg.Level.SetLevel(zap.ErrorLevel)
				logger, _ := cfg.Build()
				return logger
			}),
			fx.Option(fx.WithLogger(
				func(logger *zap.Logger) fxevent.Logger {
					return &fxevent.ZapLogger{Logger: logger}
				},
			)),
		)
}
