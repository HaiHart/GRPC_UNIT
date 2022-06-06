package network

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"time"
)

var networkMapping = map[string]EthConfig{
	"Ethereum-Mainnet": newEthereumMainnetConfig(),
}

// NewEthereumPreset returns an Ethereum configuration for the given network name. For most of these presets, the client will present itself as only having the genesis block, but that shouldn't matter too much.
func NewEthereumPreset(network string) (EthConfig, error) {
	config, ok := networkMapping[network]
	if !ok {
		return unknownConfig(), fmt.Errorf("network %v did not have an available configuration", network)
	}
	return config, nil
}

func newEthereumMainnetConfig() EthConfig {
	td, ok := new(big.Int).SetString("0400000000", 16)
	if !ok {
		panic("could not load Ethereum Mainnet configuration")
	}

	return EthConfig{
		Network:            1,
		TotalDifficulty:    td,
		Head:               common.HexToHash("0xd4e56740f876aef8c010b86a40d5f56745a118d0906a34e69aec8c0db1cb8fa3"),
		Genesis:            common.HexToHash("0xd4e56740f876aef8c010b86a40d5f56745a118d0906a34e69aec8c0db1cb8fa3"),
		IgnoreBlockTimeout: 150 * time.Second,
	}
}

func unknownConfig() EthConfig {
	return EthConfig{}
}
