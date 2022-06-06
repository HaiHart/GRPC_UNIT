package config

import (
	"github.com/massbitprotocol/turbo/logger"
	"github.com/massbitprotocol/turbo/utils"
	"github.com/urfave/cli/v2"
)

// TurboConfig represent generic node configuration
type TurboConfig struct {
	*Env
	*logger.Config
	*TxTraceLog
}

// NewTurboConfigFromCLI builds node config from the CLI context
func NewTurboConfigFromCLI(ctx *cli.Context) (*TurboConfig, error) {
	env, err := NewEnvFromCLI(ctx.String(utils.EnvFlag.Name), ctx)
	if err != nil {
		return nil, err
	}

	log, txTraceLog, err := NewLogFromCLI(ctx)
	if err != nil {
		return nil, err
	}

	turboConfig := &TurboConfig{
		Env:        env,
		Config:     log,
		TxTraceLog: txTraceLog,
	}

	return turboConfig, nil
}
