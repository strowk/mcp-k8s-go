package deployment

import (
	"github.com/strowk/foxy-contexts/pkg/fxctx"
	"go.uber.org/fx"
)

var Registration = fx.Options(
	fx.Provide(fxctx.AsTool(NewGetDeploymentTool)),
	fx.Provide(fxctx.AsTool(NewListDeploymentsTool)),
)
