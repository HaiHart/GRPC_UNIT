package config

import (
	"fmt"
	"github.com/urfave/cli/v2"
)

const (
	Mainnet = "mainnet"
)

// Env represents configuration pertaining to a specific development environment
type Env struct {
	Environment string
}

// NewEnvFromCLI parses an environment from a CLI provided string
// provided arguments override defaults from the --env argument
func NewEnvFromCLI(env string, ctx *cli.Context) (*Env, error) {
	gatewayEnv, err := NewEnv(env)
	if err != nil {
		return gatewayEnv, err
	}
	return gatewayEnv, nil
}

// NewEnv returns the preconfigured environment without the need for CLI overrides
func NewEnv(env string) (*Env, error) {
	var gatewayEnv Env
	switch env {
	case Mainnet:
		gatewayEnv = MainnetEnv
	default:
		return nil, fmt.Errorf("could not parse unrecognized env: %v", env)
	}
	return &gatewayEnv, nil
}
