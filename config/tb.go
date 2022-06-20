package config

import (
	"github.com/massbitprotocol/turbo/logger"
	"github.com/massbitprotocol/turbo/utils"
	"github.com/urfave/cli/v2"
)

// TurboConfig represent generic node configuration
type TurboConfig struct {
	NodeType utils.NodeType

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

	nodeType, err := utils.FromStringToNodeType(ctx.String(utils.NodeTypeFlag.Name))
	if err != nil {
		return nil, err
	}

	config := &TurboConfig{
		NodeType:   nodeType,
		Env:        env,
		Config:     log,
		TxTraceLog: txTraceLog,
	}

	return config, nil
}
