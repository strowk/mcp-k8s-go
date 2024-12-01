package main

import (
	"context"

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

func getCapabilities() mcp.ServerCapabilities {
	return mcp.ServerCapabilities{
		Resources: &mcp.ServerCapabilitiesResources{
			ListChanged: utils.Ptr(false),
			Subscribe:   utils.Ptr(false),
		},
	}
}

func getImplementation() mcp.Implementation {
	return mcp.Implementation{
		Name:    "mcp-k8s-go",
		Version: "0.0.1",
	}
}

func main() {
	tp := stdio.NewTransport()

	app := fx.New(
		fx.Provide(func() k8s.ClientPool {
			return k8s.NewClientPool()
		}),
		fx.Provide(func() clientcmd.ClientConfig {
			return k8s.GetKubeConfig()
		}),
		fx.Provide(func() (*kubernetes.Clientset, error) {
			return k8s.GetKubeClientset()
		}),
		fx.Provide(func() server.Transport {
			return tp
		}),
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
		fxctx.ProvideToolMux(),
		fx.Provide(fxctx.AsTool(tools.NewListContextsTool)),
		fx.Provide(fxctx.AsTool(tools.NewListPodsTool)),
		fx.Provide(fxctx.AsTool(tools.NewListEventsTool)),
		fxctx.ProvideResourceMux(),
		fx.Provide(fxctx.AsResourceProvider(resources.NewContextsResourceProvider)),
		fx.Invoke(func(
			lc fx.Lifecycle,
			tp server.Transport,
			toolMux fxctx.ToolMux,
			resourceMux fxctx.ResourceMux,
		) {
			shut := make(chan struct{})
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					go func() {
						// TODO: figure out how to stop this gracefully
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
					}()
					return nil
				},
				OnStop: func(ctx context.Context) error {
					close(shut)
					return nil
				},
			})
		}),
	)

	app.Run()
}
