package main

import (
	"context"
	"os"

	"github.com/strowk/mcp-k8s-go/internal/k8s"
	"github.com/strowk/mcp-k8s-go/internal/resources"
	"github.com/strowk/mcp-k8s-go/internal/tools"
	"github.com/strowk/mcp-k8s-go/internal/utils"

	"github.com/strowk/foxy-contexts/pkg/fxctx"
	"github.com/strowk/foxy-contexts/pkg/mcp"
	"github.com/strowk/foxy-contexts/pkg/server"
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
	}
}

func getImplementation() *mcp.Implementation {
	return &mcp.Implementation{
		Name:    "mcp-k8s-go",
		Version: "0.0.1",
	}
}

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	if len(os.Args) > 1 {
		if os.Args[1] == "--version" {
			println(version)
		}
		if os.Args[1] == "version" {
			println("Version: ", version)
			println("Commit: ", commit)
			println("Date: ", date)
		}
		return
	}

	app := fx.New(
		// k8s client
		fx.Provide(func() clientcmd.ClientConfig {
			return k8s.GetKubeConfig()
		}),
		fx.Provide(func() (*kubernetes.Clientset, error) {
			return k8s.GetKubeClientset()
		}),
		fx.Provide(func() k8s.ClientPool {
			return k8s.NewClientPool()
		}),

		// transport
		fx.Provide(func() server.Transport {
			return stdio.NewTransport()
		}),

		// logging configuration
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

		// tools
		fxctx.ProvideToolMux(),
		fx.Provide(fxctx.AsTool(tools.NewListContextsTool)),
		fx.Provide(fxctx.AsTool(tools.NewListPodsTool)),
		fx.Provide(fxctx.AsTool(tools.NewListEventsTool)),
		fx.Provide(fxctx.AsTool(tools.NewPodLogsTool)),
		fx.Provide(fxctx.AsTool(tools.NewListServicesTool)),

		// resources
		fxctx.ProvideResourceMux(),
		fx.Provide(fxctx.AsResourceProvider(resources.NewContextsResourceProvider)),

		// server
		fx.Invoke(func(
			lc fx.Lifecycle,
			tp server.Transport,
			shutdowner fx.Shutdowner,
			toolMux fxctx.ToolMux,
			resourceMux fxctx.ResourceMux,
		) {
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					go func() {
						tp.Run(
							getCapabilities(),
							getImplementation(),
							server.ServerStartCallbackOption{
								Callback: func(s server.Server) {
									toolMux.RegisterHandlers(s)
									resourceMux.RegisterHandlers(s)
								},
							},
						)

						// if transport is stopped, this means that standard
						// input was closed, so we should shutdown the application
						shutdowner.Shutdown()
					}()
					return nil
				},
				OnStop: func(ctx context.Context) error {
					return tp.Shutdown(ctx)
				},
			})
		}),
	)

	// fx would handle the lifecycle of the application
	// including starting and stopping the server, handling
	// termination signals and so on
	app.Run()
}
