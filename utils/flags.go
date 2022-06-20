package utils

import "github.com/urfave/cli/v2"

// CLI flag variable definitions
var (
	PortFlag = &cli.IntFlag{
		Name:    "port",
		Usage:   "port for accepting gateway connection",
		Aliases: []string{"p"},
		Value:   1809,
	}
	EnvFlag = &cli.StringFlag{
		Name:  "env",
		Usage: "development environment (testnet, mainnet)",
		Value: "mainnet",
	}
	BlockchainNetworkFlag = &cli.StringFlag{
		Name:  "blockchain-network",
		Usage: "determine the blockchain network (Ethereum-Mainnet or BSC-Mainnet)",
		Value: "Ethereum-Mainnet",
	}
	DisableProfilingFlag = &cli.BoolFlag{
		Name:  "disable-profiling",
		Usage: "true to disable the pprof http server (for relays, where profiling is enabled by default)",
		Value: false,
	}
	DataDirFlag = &cli.StringFlag{
		Name:  "data-dir",
		Usage: "directory for storing various persistent files (e.g. private SSL certs)",
		Value: "datadir",
	}
	LogLevelFlag = &cli.StringFlag{
		Name:    "log-level",
		Usage:   "log level for stdout",
		Aliases: []string{"l"},
		Value:   "info",
	}
	LogFileLevelFlag = &cli.StringFlag{
		Name:  "log-file-level",
		Usage: "log level for the log file",
		Value: "info",
	}
	LogMaxSizeFlag = &cli.IntFlag{
		Name:  "log-max-size",
		Usage: "maximum size in megabytes of the log file before it gets rotated",
		Value: 100,
	}
	LogMaxAgeFlag = &cli.IntFlag{
		Name:  "log-max-age",
		Usage: "maximum number of days to retain old log files based on the timestamp encoded in their filename",
		Value: 10,
	}
	LogMaxBackupsFlag = &cli.IntFlag{
		Name:  "log-max-backups",
		Usage: "maximum number of old log files to retain",
		Value: 10,
	}
	TxTraceEnabledFlag = &cli.BoolFlag{
		Name:  "txtrace",
		Usage: "for gateways only, enables transaction trace logging",
		Value: false,
	}
	TxTraceMaxFileSizeFlag = &cli.IntFlag{
		Name:  "txtrace-max-file-size",
		Usage: "for gateways only, sets max size of individual tx trace log file (megabytes)",
		Value: 100,
	}
	TxTraceMaxBackupFilesFlag = &cli.IntFlag{
		Name:  "txtrace-max-backup-files",
		Usage: "for gateways only, sets max number of backup tx trace log files retained (0 enables unlimited backups)",
		Value: 3,
	}
	NodeTypeFlag = &cli.StringFlag{
		Name:  "node-type",
		Usage: "set node type",
		Value: "gateway",
	}
)
